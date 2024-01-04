package game

import (
	"bytes"
	"container/heap"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/kelindar/bitmap"
	"github.com/qbradq/after/lib/util"
)

const (
	CityMapWidth              int = 660  // Width of a city map in chunks
	CityMapHeight             int = 660  // Height of a city map in chunks
	maxInMemoryChunks         int = 1024 // Max chunks to keep in hot memory
	purgeInMemoryChunksTarget int = 512  // Number of chunks to keep in hot memory after purging least-recently used chunks
	chunkUpdateRadius         int = 10   // Number of chunks away from the player to update actors
	chunkLoadRadius           int = 12   // Number of chunks away from the player to keep chunks hot-loaded
)

// Return slice for GetActors
var gaRet []*Actor

// Buffer for visibility set
var visBuf bitmap.Bitmap

// Buffer for remembered set
var remBuf bitmap.Bitmap

// CityMap represents the entire world of the game in terms of which chunks go
// where.
type CityMap struct {
	// Dynamic persistent data
	Player *Player   // Player actor
	Now    time.Time // Current in-game time
	// Static persistent data
	chunks []*Chunk // The chunks of the map
	// Reconstructed values
	Bounds     util.Rect // Bounds of the city map in chunks
	TileBounds util.Rect // Bounds of the city map in tiles
	// Working variables
	inMemoryChunks      bitmap.Bitmap    // Bitmap of all chunks loaded into memory
	inMemoryChunksCount int              // Running count of in-memory chunks to avoid excessive calls to bitmap.Count()
	chunksGenerated     bitmap.Bitmap    // Bitmap of all chunks that have been generated
	cgDirty             bool             // ChunksGenerated has been altered since the last call to SaveBitmaps
	updateSet           map[int]struct{} // Set of all chunks in the current update set
	usNewCache          []int            // Cache of chunk indexes of newly added chunks to the update set
	usOldCache          []int            // Cache of chunk indexes of newly removed chunks to the update set
	aq                  actorQueue       // Queue of all actors within update range
	itemsWithinCache    []*Item          // Return slice for ItemsWithin()
	actorsWithinCache   []*Actor         // Return slice for ActorsWithin()
}

// NewCityMap allocates and returns a new CityMap structure.
func NewCityMap() *CityMap {
	m := &CityMap{
		Bounds:     util.NewRectWH(CityMapWidth, CityMapHeight),
		TileBounds: util.NewRectWH(CityMapWidth*ChunkWidth, CityMapHeight*ChunkHeight),
		chunks:     make([]*Chunk, CityMapWidth*CityMapHeight),
		updateSet:  map[int]struct{}{},
		usNewCache: make([]int, 0, chunkUpdateRadius*chunkUpdateRadius),
		usOldCache: make([]int, 0, chunkUpdateRadius*chunkUpdateRadius),
		aq:         actorQueue{},
	}
	// Configure the starting time as two years from now at 0800
	t := time.Now().Add(time.Hour * 24 * 730)
	m.Now = time.Date(t.Year(), t.Month(), t.Day(), 8, 0, 0, 0, t.Location())
	for i := range m.chunks {
		m.chunks[i] = NewChunk(i%CityMapWidth, i/CityMapWidth, uint32(i))
	}
	heap.Init(&m.aq)
	return m
}

// SaveCityPlan saves the city plan in the current save database.
func (m *CityMap) SaveCityPlan() {
	// Write
	var w = bytes.NewBuffer(nil)
	m.Write(w)
	SaveValue("CityMap.Plan", w.Bytes())
}

// LoadCityPlan loads the city plan from the current save database.
func (m *CityMap) LoadCityPlan() {
	r := LoadValue("CityMap.Plan")
	m.Read(r)
	m.LoadBitmaps()
}

// Write writes the city-level map information to the writer.
func (m *CityMap) Write(w io.Writer) {
	// Build dictionary of chunk generator IDs
	dict := util.NewDictionary()
	for _, c := range m.chunks {
		dict.Put(c.Generator.GetID())
	}
	// Write the file
	util.PutUint32(w, 0) // Version
	util.PutDictionary(w, dict)
	for _, c := range m.chunks {
		util.PutUint16(w, dict.Get(c.Generator.GetID()))
		util.PutPoint(w, c.ChunkGenOffset)
		util.PutByte(w, byte(c.Facing))
		util.PutByte(w, byte(c.Flags))
	}
}

// SaveDynamicData writes top-level dynamic map data like the player actor's
// current position.
func (m *CityMap) SaveDynamicData() {
	w := bytes.NewBuffer(nil)
	util.PutUint32(w, 0) // Version
	m.Player.Write(w)    // Player
	SaveValue("CityMap.DynamicData", w.Bytes())
}

// LoadDynamicData loads top-level dynamic map data like the player actor's
// current position.
func (m *CityMap) LoadDynamicData() {
	r := LoadValue("CityMap.DynamicData")
	_ = util.GetUint32(r) // Version
	m.Player = NewPlayerFromReader(r)
}

// Read reads the city-level map information from the buffer.
func (m *CityMap) Read(r io.Reader) {
	_ = util.GetUint32(r) // Version
	dict := util.GetDictionary(r)
	for _, c := range m.chunks {
		s := dict.Lookup(util.GetUint16(r))
		c.Generator = GetChunkGen(s)
		c.ChunkGenOffset = util.GetPoint(r)
		c.Facing = util.Facing(util.GetByte(r))
		c.Flags = ChunkFlags(util.GetByte(r))
		c.Generator.AssignStaticInfo(c)
	}
}

// GetChunkFromMapPoint returns the chunk definition at the given map location
// or nil if out of bounds. Note the point is in chunks not tiles.
func (m *CityMap) GetChunkFromMapPoint(p util.Point) *Chunk {
	if !m.Bounds.Contains(p) {
		return nil
	}
	return m.chunks[p.Y*CityMapWidth+p.X]
}

// GetChunk returns the correct chunk for the given absolute tile point or nil
// if the point is out of bounds.
func (m *CityMap) GetChunk(p util.Point) *Chunk {
	if !m.TileBounds.Contains(p) {
		return nil
	}
	return m.chunks[(p.Y/ChunkHeight)*CityMapWidth+(p.X/ChunkWidth)]
}

// GetTile returns the tile at the given absolute tile point or nil if the point
// is out of bounds.
func (m *CityMap) GetTile(p util.Point) *TileDef {
	if !m.TileBounds.Contains(p) {
		return nil
	}
	c := m.chunks[(p.Y/ChunkHeight)*CityMapWidth+(p.X/ChunkWidth)]
	t := c.Tiles[(p.Y%ChunkHeight)*ChunkWidth+(p.X%ChunkWidth)]
	return t
}

// EnsureLoaded ensures that all chunks in the area given in chunk coordinates
// have been generated and are loaded into memory.
func (m *CityMap) EnsureLoaded(r util.Rect) {
	if r.TL.X < 0 {
		r.TL.X = 0
	}
	if r.TL.Y < 0 {
		r.TL.Y = 0
	}
	if r.BR.X >= CityMapWidth {
		r.BR.X = CityMapWidth - 1
	}
	if r.BR.Y >= CityMapHeight {
		r.BR.Y = CityMapHeight - 1
	}
	// Load all chunks within the area
	var p util.Point
	now := time.Now()
	for p.Y = r.TL.Y; p.Y <= r.BR.Y; p.Y++ {
		for p.X = r.TL.X; p.X <= r.BR.X; p.X++ {
			r := chunkRefForPoint(p)
			c := m.chunks[r]
			m.LoadChunk(c, now)
		}
	}
	// After we load chunks we need to make sure to purge old chunks so we don't
	// fill all available RAM with chunk data.
	m.purgeOldChunks()
}

// LoadChunk loads the passed-in chunk or generates it if needed. This function
// is cheap if the chunk is already in memory.
func (m *CityMap) LoadChunk(c *Chunk, now time.Time) {
	c.Loaded = now
	// Bail if we are already loaded
	if c.Tiles != nil {
		if c.bitmapsDirty {
			c.RebuildBitmaps()
		}
		return
	}
	// There is no possibility of error after this point so go ahead
	// and mark the chunk as in-memory
	m.inMemoryChunks.Set(c.Ref)
	m.inMemoryChunksCount++
	// Allocate memory
	c.Tiles = make([]*TileDef, ChunkWidth*ChunkHeight)
	// Generate the chunk if this has never happened before
	if !m.chunksGenerated.Contains(c.Ref) {
		m.chunksGenerated.Set(c.Ref)
		m.cgDirty = true
		c.Generator.Generate(c, m)
		c.RebuildBitmaps()
		w := bytes.NewBuffer(nil)
		c.Write(w)
		SaveValue(fmt.Sprintf("Chunk-%d", c.Ref), w.Bytes())
		return
	}
	// Otherwise load the chunk into memory from the save database
	n := fmt.Sprintf("Chunk-%d", c.Ref)
	buf := LoadValue(n)
	c.Read(buf)
	c.RebuildBitmaps()
}

// purgeOldChunks purges chunks in least-recently-used first order down to the
// target number if the number of chunks in the memory cache is greater than the
// maximum.
func (m *CityMap) purgeOldChunks() {
	// Short-circuit condition
	if m.inMemoryChunksCount <= maxInMemoryChunks {
		return
	}
	// Sort the chunks by time last updated
	cRefs := make([]uint32, 0, m.inMemoryChunksCount)
	m.inMemoryChunks.Range(func(x uint32) {
		cRefs = append(cRefs, x)
	})
	sort.Slice(cRefs, func(i, j int) bool {
		return cRefs[i] < cRefs[j]
	})
	// Persist and unload the oldest chunks until we reach the purge target
	bufs := map[uint32][]byte{}
	for _, cr := range cRefs[:maxInMemoryChunks-purgeInMemoryChunksTarget] {
		w := bytes.NewBuffer(nil)
		c := m.chunks[cr]
		c.Write(w)
		bufs[cr] = w.Bytes()
		c.Unload()
		m.inMemoryChunks.Remove(cr)
		m.inMemoryChunksCount--
	}
	// Save all unloaded chunks to the database
	for k, v := range bufs {
		name := fmt.Sprintf("Chunk-%d", k)
		SaveValue(name, v)
	}
	// If any chunks updated the tile cross references we need to save them
	if crossReferencesDirty {
		SaveTileRefs()
	}
	m.SaveBitmaps()
}

// saveAllChunks saves all in-memory chunks to the current save database without
// freeing memory.
func (m *CityMap) saveAllChunks() {
	// Accumulate all data
	bufs := map[uint32][]byte{}
	m.inMemoryChunks.Range(func(x uint32) {
		w := bytes.NewBuffer(nil)
		c := m.chunks[x]
		c.Write(w)
		bufs[x] = w.Bytes()
	})
	// Write to database
	for r, v := range bufs {
		name := fmt.Sprintf("Chunk-%d", r)
		SaveValue(name, v)
	}
	// If any chunks updated the tile cross references we need to save them
	if crossReferencesDirty {
		SaveTileRefs()
	}
	m.SaveBitmaps()
}

// SaveBitmaps saves all persistent bitmaps.
func (m *CityMap) SaveBitmaps() {
	if !m.cgDirty {
		return
	}
	w := bytes.NewBuffer(nil)
	util.PutUint32(w, 0) // Version
	m.chunksGenerated.WriteTo(w)
	SaveValue("CityMap.ChunksGenerated", w.Bytes())
	m.cgDirty = false
}

// LoadBitmaps loads all persistent bitmaps.
func (m *CityMap) LoadBitmaps() {
	r := LoadValue("CityMap.ChunksGenerated")
	_ = util.GetUint32(r) // Version
	m.chunksGenerated.ReadFrom(r)
}

// FullSave commits the entire working set to the current save database without
// freeing memory.
func (m *CityMap) FullSave() {
	m.saveAllChunks()
	m.SaveDynamicData()
}

func chunkRefForPoint(p util.Point) uint32 {
	return uint32(p.Y*CityMapWidth + p.X)
}

// GetActors returns a slice of all of the actors within the given bounds. The
// returned slice is reused by subsequent calls to GetActors().
func (m *CityMap) GetActors(b util.Rect) []*Actor {
	gaRet = gaRet[:0]
	cb := util.Rect{
		TL: util.Point{
			X: b.TL.X / ChunkWidth,
			Y: b.TL.Y / ChunkHeight,
		},
		BR: util.Point{
			X: b.BR.X / ChunkWidth,
			Y: b.BR.Y / ChunkHeight,
		},
	}
	if cb.TL.X < 0 {
		cb.TL.X = 0
	}
	if cb.TL.Y < 0 {
		cb.TL.Y = 0
	}
	if cb.BR.X >= CityMapWidth {
		cb.BR.X = CityMapWidth - 1
	}
	if cb.BR.Y >= CityMapHeight {
		cb.BR.Y = CityMapHeight - 1
	}
	var p util.Point
	for p.Y = cb.TL.Y; p.Y <= cb.BR.Y; p.Y++ {
		for p.X = cb.TL.X; p.X <= cb.BR.X; p.X++ {
			c := m.GetChunkFromMapPoint(p)
			for _, a := range c.Actors {
				if b.Contains(a.Position) {
					gaRet = append(gaRet, a)
				}
			}
		}
	}
	return gaRet
}

// PlaceItem adds the item to the city at it's current location.
func (m *CityMap) PlaceItem(i *Item) {
	m.GetChunk(i.Position).PlaceItem(i)
}

// RemoveItem removes the item from the city map.
func (m *CityMap) RemoveItem(i *Item) {
	m.GetChunk(i.Position).RemoveItem(i)
}

// ItemsAt returns the items at the given position.
func (m *CityMap) ItemsAt(p util.Point) []*Item {
	var ret []*Item
	c := m.GetChunk(p)
	for _, i := range c.Items {
		if i.Position == p {
			ret = append(ret, i)
		}
	}
	return ret
}

// ItemsWithin returns the items within the given bounds. Subsequent calls to
// ItemsWithin will re-use the same slice.
func (m *CityMap) ItemsWithin(b util.Rect) []*Item {
	m.itemsWithinCache = m.itemsWithinCache[:0]
	cb := util.NewRectXYWH(b.TL.X/ChunkWidth, b.TL.Y/ChunkHeight, b.Width()/ChunkWidth+1, b.Height()/ChunkHeight+1)
	for cy := cb.TL.Y; cy <= cb.BR.Y; cy++ {
		for cx := cb.TL.X; cx <= cb.BR.X; cx++ {
			c := m.chunks[cy*CityMapWidth+cx]
			for _, i := range c.Items {
				if b.Contains(i.Position) {
					m.itemsWithinCache = append(m.itemsWithinCache, i)
				}
			}
		}
	}
	return m.itemsWithinCache
}

// AddActor adds the actor to the city at it's current location.
func (m *CityMap) AddActor(a *Actor) {
	m.GetChunk(a.Position).PlaceActor(a)
}

// RemoveActor removes the actor from the city at it's current location.
func (m *CityMap) RemoveActor(a *Actor) {
	m.GetChunk(a.Position).RemoveActor(a)
}

// ActorAt returns the actor at the given position or nil.
func (m *CityMap) ActorAt(p util.Point) *Actor {
	c := m.GetChunk(p)
	for _, a := range c.Actors {
		if a.Position == p {
			return a
		}
	}
	return nil
}

// ActorsWithin returns the items within the given bounds.
func (m *CityMap) ActorsWithin(b util.Rect) []*Actor {
	m.actorsWithinCache = m.actorsWithinCache[:0]
	cb := util.NewRectXYWH(b.TL.X/ChunkWidth, b.TL.Y/ChunkHeight, b.Width()/ChunkWidth+1, b.Height()/ChunkHeight+1)
	for cy := cb.TL.Y; cy <= cb.BR.Y; cy++ {
		for cx := cb.TL.X; cx <= cb.BR.X; cx++ {
			c := m.chunks[cy*CityMapWidth+cx]
			for _, a := range c.Actors {
				if b.Contains(a.Position) {
					m.actorsWithinCache = append(m.actorsWithinCache, a)
				}
			}
		}
	}
	return m.actorsWithinCache
}

// StepActor attempts to move the actor in the given direction, returning true
// on success.
func (m *CityMap) StepActor(a *Actor, d util.Direction) bool {
	if d == util.DirectionInvalid {
		return false
	}
	np := a.Position.Add(util.DirectionOffsets[d.Bound()])
	if !m.TileBounds.Contains(np) {
		return false
	}
	op := a.Position
	oc := m.GetChunk(op)
	nc := m.GetChunk(np)
	oc.RemoveActor(a)
	a.Position = np
	if !nc.PlaceActor(a) {
		a.Position = op
		oc.PlaceActor(a)
		return false
	}
	return true
}

// StepPlayer attempts to move the player's actor in the given direction,
// returning true on success.
func (m *CityMap) StepPlayer(d util.Direction) bool {
	if d == util.DirectionInvalid {
		return false
	}
	np := m.Player.Position.Add(util.DirectionOffsets[d.Bound()])
	if !m.TileBounds.Contains(np) {
		return false
	}
	nc := m.GetChunk(np)
	if !nc.CanStep(&m.Player.Actor, np) {
		return false
	}
	m.Player.Position = np
	m.playerTookTurn(time.Second)
	return true
}

// MakeVisibilitySets constructs bitmaps representing the current and remembered
// visibility of each position within the bounds relative to the player. This is
// a no-op if the bounds do not contain the player. Subsequent calls to
// MakeVisibilitySets reuse the same bitmaps.
func (m *CityMap) MakeVisibilitySets(b util.Rect) (vis, rem bitmap.Bitmap) {
	// Process one line of visibility calculations
	fn := func(ps []util.Point) {
		// Range over the points
		for _, p := range ps {
			// If this point blocks visibility we are done
			c := m.GetChunk(p)
			idx := c.relOfs(p)
			if c.BlocksVis.Contains(idx) {
				return
			}
			// If not we need to mark the point and all neighbors as visible
			for iy := p.Y - 1; iy <= p.Y+1; iy++ {
				if iy < 0 {
					continue
				}
				if iy >= m.TileBounds.Height() {
					break
				}
				for ix := p.X - 1; ix <= p.X+1; ix++ {
					if ix < 0 {
						continue
					}
					if ix >= m.TileBounds.Width() {
						break
					}
					p := util.Point{
						X: ix - b.TL.X,
						Y: iy - b.TL.Y,
					}
					idx := uint32(p.Y*b.Width() + p.X)
					vis.Set(idx)
				}
			}
		}
	}
	// Setup return values for reuse
	vis = visBuf
	rem = remBuf
	vis.Clear()
	rem.Clear()
	// Sanity checks
	if !b.Contains(m.Player.Position) {
		return
	}
	// Cast rays to the boarders of the rect
	for i := 0; i < b.Width(); i++ {
		fn(util.Ray(m.Player.Position, util.Point{
			X: b.TL.X + i,
			Y: b.TL.Y,
		}))
		fn(util.Ray(m.Player.Position, util.Point{
			X: b.TL.X + i,
			Y: b.BR.Y,
		}))
	}
	for i := 1; i < b.Height()-1; i++ {
		fn(util.Ray(m.Player.Position, util.Point{
			X: b.TL.X,
			Y: b.TL.Y + i,
		}))
		fn(util.Ray(m.Player.Position, util.Point{
			X: b.BR.X,
			Y: b.TL.Y + i,
		}))
	}
	// Construct remembered set for chunks and return value
	var p util.Point
	for p.Y = b.TL.Y; p.Y <= b.BR.Y; p.Y++ {
		for p.X = b.TL.X; p.X <= b.BR.X; p.X++ {
			idx := uint32((p.Y-b.TL.Y)*b.Width() + (p.X - b.TL.X))
			c := m.GetChunk(p)
			cIdx := c.relOfs(p)
			if vis.Contains(idx) {
				c.HasSeen.Set(cIdx)
			}
			if c.HasSeen.Contains(cIdx) {
				rem.Set(idx)
			}
		}
	}
	return
}

// CanSeePlayerFrom returns true if there is line of sight between the given
// point and the player.
func (m *CityMap) CanSeePlayerFrom(p util.Point) bool {
	for _, p := range util.Ray(p, m.Player.Position) {
		c := m.GetChunk(p)
		if !c.BlocksVis.Contains(c.relOfs(p)) {
			return false
		}
	}
	return true
}

// Update updates the game world for d duration based around point p.
func (m *CityMap) Update(p util.Point, d time.Duration) {
	fn := func(p util.Point) int {
		return p.Y*CityMapWidth + p.X
	}
	// Reset caches
	m.usNewCache = m.usNewCache[:0]
	m.usOldCache = m.usOldCache[:0]
	// Establish working parameters
	cp := util.Point{
		X: p.X / ChunkWidth,
		Y: p.Y / ChunkHeight,
	}
	lb := util.NewRectFromRadius(cp, chunkLoadRadius).Overlap(m.Bounds)
	ub := util.NewRectFromRadius(cp, chunkUpdateRadius).Overlap(m.Bounds)
	// Load chunks
	m.EnsureLoaded(lb)
	// Prep new chunks set
	newSet := map[int]struct{}{}
	for p.Y = ub.TL.Y; p.Y <= ub.BR.Y; p.Y++ {
		for p.X = ub.TL.X; p.X <= ub.BR.X; p.X++ {
			idx := fn(p)
			newSet[idx] = struct{}{}
			if _, found := m.updateSet[idx]; !found {
				m.usNewCache = append(m.usNewCache, idx)
			}
		}
	}
	// Prep old chunks set
	for k := range m.updateSet {
		if _, found := newSet[k]; !found {
			m.usOldCache = append(m.usOldCache, k)
		}
	}
	m.updateSet = newSet
	// Remove actors in the old chunks from the priority queue
	for _, idx := range m.usOldCache {
		c := m.chunks[idx]
		for _, a := range c.Actors {
			heap.Remove(&m.aq, a.pqIdx)
		}
	}
	// Add actors in the new chunks to the priority queue
	for _, idx := range m.usNewCache {
		c := m.chunks[idx]
		for _, a := range c.Actors {
			heap.Push(&m.aq, a)
		}
	}
	// Step time and process the priority queue
	m.Now = m.Now.Add(d)
	for {
		a := heap.Pop(&m.aq).(*Actor)
		if m.Now.Before(a.NextThink) {
			heap.Push(&m.aq, a)
			break
		}
		a.NextThink = a.NextThink.Add(a.AIModel.Act(a, a.NextThink, m))
		heap.Push(&m.aq, a)
	}
}

// playerTookTurn is responsible for updating the city map model for the given
// duration as well as anything else that should happen after the player's turn.
func (m *CityMap) playerTookTurn(d time.Duration) {
	m.Update(m.Player.Position, d)
}
