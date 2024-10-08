package game

import (
	"bytes"
	"container/heap"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/kelindar/bitmap"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

const (
	CityMapWidth              int = 660  // Width of a city map in chunks
	CityMapHeight             int = 660  // Height of a city map in chunks
	maxInMemoryChunks         int = 1024 // Max chunks to keep in hot memory
	purgeInMemoryChunksTarget int = 512  // Number of chunks to keep in hot memory after purging least-recently used chunks
	chunkUpdateRadius         int = 4    // Number of chunks away from the player to update actors
	chunkLoadRadius           int = 5    // Number of chunks away from the player to keep chunks hot-loaded
)

// CityMap represents the entire world of the game in terms of which chunks go
// where.
type CityMap struct {
	//
	// Dynamic persistent data
	//

	Player *Player   // Player actor
	Now    time.Time // Current in-game time

	//
	// Static persistent data
	//

	Chunks []*Chunk // The chunks of the map

	//
	// Reconstructed values
	//

	Bounds     util.Rect // Bounds of the city map in chunks
	TileBounds util.Rect // Bounds of the city map in tiles

	//
	// Working variables
	//

	Visibility          bitmap.Bitmap    // Last visibility set calculated for the player
	Remembered          bitmap.Bitmap    // Last remembered set calculated for the player
	BitmapBounds        util.Rect        // Bounds of the Visibility and Remembered bitmaps
	inMemoryChunks      bitmap.Bitmap    // Bitmap of all chunks loaded into memory
	inMemoryChunksCount int              // Running count of in-memory chunks to avoid excessive calls to bitmap.Count()
	chunksGenerated     bitmap.Bitmap    // Bitmap of all chunks that have been generated
	cgDirty             bool             // ChunksGenerated has been altered since the last call to SaveBitmaps
	updateSet           map[int]struct{} // Set of all chunks in the current update set
	usNewCache          []int            // Cache of chunk indexes of newly added chunks to the update set
	usOldCache          []int            // Cache of chunk indexes of newly removed chunks to the update set
	aq                  actorQueue       // Queue of all actors within update range
	gaRet               []*Actor         // Return slice for GetActors
	itemsWithinCache    []*Item          // Return slice for ItemsWithin()
	actorsWithinCache   []*Actor         // Return slice for ActorsWithin()
	chunksWithinCache   []*Chunk         // Return slice for ChunksWithin()
	vehiclesWithinCache []*Vehicle       // Return slice for VehiclesWithin()
	updateBounds        util.Rect        // Bounds of the current update
	loadBounds          util.Rect        // Load bounds of the current update
}

// NewCityMap allocates and returns a new CityMap structure.
func NewCityMap() *CityMap {
	m := &CityMap{
		Bounds:     util.NewRectWH(CityMapWidth, CityMapHeight),
		TileBounds: util.NewRectWH(CityMapWidth*ChunkWidth, CityMapHeight*ChunkHeight),
		Chunks:     make([]*Chunk, CityMapWidth*CityMapHeight),
		updateSet:  map[int]struct{}{},
		usNewCache: make([]int, 0, chunkUpdateRadius*chunkUpdateRadius),
		usOldCache: make([]int, 0, chunkUpdateRadius*chunkUpdateRadius),
		aq:         actorQueue{},
	}
	// Configure the starting time as two years from now at 0800
	t := time.Now().Add(time.Hour * 24 * 730)
	m.Now = time.Date(t.Year(), t.Month(), t.Day(), 8, 0, 0, 0, t.Location())
	for i := range m.Chunks {
		m.Chunks[i] = NewChunk(i%CityMapWidth, i/CityMapWidth, uint32(i))
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
	// Build string dictionary for generator and variant IDs
	dict := util.NewDictionary()
	for _, c := range m.Chunks {
		dict.Put(c.Generator.GetGroup())
		dict.Put(c.Generator.GetVariant())
	}
	// Write the file
	util.PutUint32(w, 0) // Version
	util.PutDictionary(w, dict)
	for _, c := range m.Chunks {
		util.PutUint16(w, dict.Get(c.Generator.GetGroup()))
		util.PutUint16(w, dict.Get(c.Generator.GetVariant()))
		util.PutPoint(w, c.ChunkGenOffset)
		util.PutByte(w, byte(c.Facing))
		util.PutByte(w, byte(c.Flags))
	}
}

// SaveDynamicData writes top-level dynamic map data.
func (m *CityMap) SaveDynamicData() {
	w := bytes.NewBuffer(nil)
	util.PutUint32(w, 0)   // Version
	m.Player.Write(w)      // Player
	util.PutTime(w, m.Now) // Current time
	SaveValue("CityMap.DynamicData", w.Bytes())
}

// LoadDynamicData loads top-level dynamic map data.
func (m *CityMap) LoadDynamicData() {
	r := LoadValue("CityMap.DynamicData")
	_ = util.GetUint32(r)             // Version
	m.Player = NewPlayerFromReader(r) // Player
	m.Now = util.GetTime(r)           // Current time
}

// Read reads the city-level map information from the buffer.
func (m *CityMap) Read(r io.Reader) {
	_ = util.GetUint32(r) // Version
	dict := util.GetDictionary(r)
	for _, c := range m.Chunks {
		s := dict.Lookup(util.GetUint16(r))
		v := dict.Lookup(util.GetUint16(r))
		c.Generator = GetChunkGen(s, v)
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
	return m.Chunks[p.Y*CityMapWidth+p.X]
}

// GetChunk returns the correct chunk for the given absolute tile point or nil
// if the point is out of bounds.
func (m *CityMap) GetChunk(p util.Point) *Chunk {
	if !m.TileBounds.Contains(p) {
		return nil
	}
	return m.Chunks[(p.Y/ChunkHeight)*CityMapWidth+(p.X/ChunkWidth)]
}

// ChunksWithin returns all chunks within the given bounds. The return value
// will be reused on subsequent calls to GetChunksWithin.
func (m *CityMap) ChunksWithin(b util.Rect) []*Chunk {
	m.chunksWithinCache = m.chunksWithinCache[:0]
	cb := m.TileBounds.Overlap(b).Divide(ChunkWidth)
	var p util.Point
	for p.Y = cb.TL.Y; p.Y <= cb.BR.Y; p.Y++ {
		for p.X = cb.TL.X; p.X <= cb.BR.X; p.X++ {
			c := m.GetChunkFromMapPoint(p)
			if c == nil {
				continue
			}
			m.chunksWithinCache = append(m.chunksWithinCache, c)
		}
	}
	return m.chunksWithinCache
}

// GetTile returns the tile at the given absolute tile point or nil if the point
// is out of bounds.
func (m *CityMap) GetTile(p util.Point) *TileDef {
	c := m.Chunks[(p.Y/ChunkHeight)*CityMapWidth+(p.X/ChunkWidth)]
	if c.Loaded.IsZero() {
		return nil
	}
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
			ref := chunkRefForPoint(p)
			c := m.Chunks[ref]
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
		c.bitmapsDirty = true
		w := bytes.NewBuffer(nil)
		c.Write(w)
		SaveValue(fmt.Sprintf("Chunk-%d", c.Ref), w.Bytes())
		return
	}
	// Otherwise load the chunk into memory from the save database
	n := fmt.Sprintf("Chunk-%d", c.Ref)
	buf := LoadValue(n)
	c.Read(buf)
	c.bitmapsDirty = true
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
	slices.SortFunc[[]uint32](cRefs, func(a, b uint32) int {
		if m.Chunks[a].Loaded.Before(m.Chunks[b].Loaded) {
			return -1
		} else if m.Chunks[a].Loaded.After(m.Chunks[b].Loaded) {
			return 1
		}
		return 0
	})
	// Persist and unload the oldest chunks until we reach the purge target
	buffers := map[uint32][]byte{}
	for _, cr := range cRefs[:maxInMemoryChunks-purgeInMemoryChunksTarget] {
		w := bytes.NewBuffer(nil)
		c := m.Chunks[cr]
		c.Write(w)
		buffers[cr] = w.Bytes()
		c.Unload()
		m.inMemoryChunks.Remove(cr)
		m.inMemoryChunksCount--
	}
	// Save all unloaded chunks to the database
	for k, v := range buffers {
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
	buffers := map[uint32][]byte{}
	m.inMemoryChunks.Range(func(x uint32) {
		w := bytes.NewBuffer(nil)
		c := m.Chunks[x]
		c.Write(w)
		buffers[x] = w.Bytes()
	})
	// Write to database
	for r, v := range buffers {
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
	m.gaRet = m.gaRet[:0]
	cb := b.Divide(ChunkWidth)
	cb = m.Bounds.Overlap(cb)
	var p util.Point
	for p.Y = cb.TL.Y; p.Y <= cb.BR.Y; p.Y++ {
		for p.X = cb.TL.X; p.X <= cb.BR.X; p.X++ {
			c := m.GetChunkFromMapPoint(p)
			for _, a := range c.Actors {
				if b.Contains(a.Position) {
					m.gaRet = append(m.gaRet, a)
				}
			}
		}
	}
	return m.gaRet
}

// PlaceItem adds the item to the city at it's current location.
func (m *CityMap) PlaceItem(i *Item, force bool) bool {
	return m.GetChunk(i.Position).PlaceItem(i, force)
}

// RemoveItem removes the item from the city map.
func (m *CityMap) RemoveItem(i *Item) bool {
	return m.GetChunk(i.Position).RemoveItem(i)
}

// ItemsAt returns the items at the given position.
func (m *CityMap) ItemsAt(p util.Point) []*Item {
	var ret []*Item
	c := m.GetChunk(p)
	if c == nil {
		return nil
	}
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
	cb := util.NewRect(b.TL.Divide(ChunkWidth), b.BR.Divide(ChunkWidth))
	cb = m.Bounds.Overlap(cb)
	for cy := cb.TL.Y; cy <= cb.BR.Y; cy++ {
		for cx := cb.TL.X; cx <= cb.BR.X; cx++ {
			c := m.Chunks[cy*CityMapWidth+cx]
			for _, i := range c.Items {
				if b.Contains(i.Position) {
					m.itemsWithinCache = append(m.itemsWithinCache, i)
				}
			}
		}
	}
	return m.itemsWithinCache
}

// PlaceVehicle attempts to place the vehicle into the city.
func (m *CityMap) PlaceVehicle(v *Vehicle) bool {
	// Place the vehicle into the parent chunk
	c := m.GetChunk(v.Bounds.TL)
	if c == nil {
		return false
	}
	if !c.PlaceVehicle(v) {
		return false
	}
	// Flag impacted chunks' bitmaps
	for _, c := range m.ChunksWithin(v.Bounds) {
		c.bitmapsDirty = true
	}
	return true
}

// RemoveVehicle attempts to remove the vehicle from the city.
func (m *CityMap) RemoveVehicle(v *Vehicle) bool {
	// Remove vehicle from parent chunk
	c := m.GetChunk(v.Bounds.TL)
	if c == nil {
		return false
	}
	if !c.RemoveVehicle(v) {
		return false
	}
	// Flag impacted chunks' bitmaps
	for _, c := range m.ChunksWithin(v.Bounds) {
		c.bitmapsDirty = true
	}
	return true
}

// VehiclesWithin returns a list of all of the vehicles who's bounds overlap
// the given bounds. The return value will be reused on future calls to
// VehiclesWithin.
func (m *CityMap) VehiclesWithin(b util.Rect) []*Vehicle {
	m.vehiclesWithinCache = m.vehiclesWithinCache[:0]
	qb := m.TileBounds.Overlap(b.Grow(16))
	for _, c := range m.ChunksWithin(qb) {
		// Skip chunks that are not yet generated or in memory
		var tz time.Time
		if !m.chunksGenerated.Contains(c.Ref) ||
			c.Loaded == tz {
			continue
		}
		for _, v := range c.Vehicles {
			if v.Bounds.Overlaps(b) {
				m.vehiclesWithinCache = append(m.vehiclesWithinCache, v)
			}
		}
	}
	return m.vehiclesWithinCache
}

// VehicleAt returns the vehicle at the given location, or nil.
func (m *CityMap) VehicleAt(p util.Point) *Vehicle {
	b := util.NewRectFromRadius(p, 16)
	for _, c := range m.ChunksWithin(b) {
		for _, v := range c.Vehicles {
			if v.Bounds.Contains(p) {
				return v
			}
		}
	}
	return nil
}

// PlaceActor adds the actor to the city at it's current location.
func (m *CityMap) PlaceActor(a *Actor, climbing bool) {
	m.GetChunk(a.Position).PlaceActor(a, climbing, m)
	if m.updateBounds.Contains(a.Position) {
		// Chunk is within the current update set so we'll have to manually push
		// them onto the action queue
		heap.Push(&m.aq, a)
	}
}

// RemoveActor removes the actor from the city at it's current location.
func (m *CityMap) RemoveActor(a *Actor) {
	m.GetChunk(a.Position).RemoveActor(a)
	if m.updateBounds.Contains(a.Position) {
		// Chunk is within the current update set so we'll have to manually
		// remove them from the action queue
		idx := -1
		for i, aqa := range m.aq {
			if aqa == a {
				idx = i
				break
			}
		}
		if idx >= 0 {
			heap.Remove(&m.aq, idx)
		}
	}
}

// ActorAt returns the actor at the given position or nil.
func (m *CityMap) ActorAt(p util.Point) *Actor {
	c := m.GetChunk(p)
	if c == nil {
		return nil
	}
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
			c := m.Chunks[cy*CityMapWidth+cx]
			for _, a := range c.Actors {
				if b.Contains(a.Position) {
					m.actorsWithinCache = append(m.actorsWithinCache, a)
				}
			}
		}
	}
	return m.actorsWithinCache
}

// StepActor attempts to move the actor in the given direction. The first return
// value is true if the actor was able to step on the location. The second
// return value is true only if the first return value is false and the actor
// successfully climbed to the location.
func (m *CityMap) StepActor(a *Actor, climbing bool, d util.Direction) (bool, bool) {
	if d == util.DirectionInvalid {
		return false, false
	}
	np := a.Position.Add(util.DirectionOffsets[d.Bound()])
	if !m.TileBounds.Contains(np) {
		return false, false
	}
	op := a.Position
	oc := m.GetChunk(op)
	nc := m.GetChunk(np)
	oc.RemoveActor(a)
	a.Position = np
	ws, cs := nc.PlaceActor(a, climbing, m)
	if !ws && !cs {
		a.Position = op
		oc.PlaceActor(a, true, m)
		return false, false
	}
	return ws, cs
}

// StepPlayer attempts to move the player's actor in the given direction
// returning true on success.
func (m *CityMap) StepPlayer(climbing bool, d util.Direction) bool {
	if d == util.DirectionInvalid {
		return false
	}
	np := m.Player.Position.Add(util.DirectionOffsets[d.Bound()])
	if !m.TileBounds.Contains(np) {
		return false
	}
	nc := m.GetChunk(np)
	ws, cs := nc.CanStep(&m.Player.Actor, np, m)
	if !ws && !cs {
		return false
	}
	if cs && !climbing {
		return false
	}
	m.Player.Position = np
	dur := time.Duration(float64(time.Second) * m.Player.WalkSpeed())
	if cs {
		dur *= 4
	} else {
		if m.Player.Running && m.Player.Stamina <= 0 {
			Log.Log(termui.ColorRed, "You are exhausted and slow to a walk.")
			m.Player.Running = false
		}
		if m.Player.Running {
			dur /= 4
			m.Player.Stamina -= float64(dur) / float64(time.Second*30) // Can run for about 30 seconds - the duration is not terribly realistic but the limit is for game play balance
			m.Player.Stamina -= float64(dur) / float64(time.Minute*5)  // Counteract stamina gain
		}
	}
	m.PlayerTookTurn(dur, nil)
	return true
}

// PlayerCanClimb returns true if the player can climb in the given direction.
func (m *CityMap) PlayerCanClimb(d util.Direction) bool {
	if d == util.DirectionInvalid {
		return false
	}
	np := m.Player.Position.Add(util.DirectionOffsets[d.Bound()])
	if !m.TileBounds.Contains(np) {
		return false
	}
	nc := m.GetChunk(np)
	_, cs := nc.CanStep(&m.Player.Actor, np, m)
	return cs
}

// MakeVisibilitySets constructs bitmaps representing the current and remembered
// visibility of each position within the bounds relative to the player. This is
// a no-op if the bounds do not contain the player. Visibility sets are stored
// in Visibility and Remembered members.
func (m *CityMap) MakeVisibilitySets(b util.Rect) {
	var dp util.Point
	// Process one line of visibility calculations
	fn := func(ps []util.Point) {
		// Range over the points excluding the first
		done := false
		for _, p := range ps[1:] {
			// Bail if we've already hit a non-visible position
			if done {
				break
			}
			// If this point blocks visibility we are done
			c := m.GetChunk(p)
			if c.BlocksVis.Contains(c.relOfs(p)) {
				done = true
			}
			// Skip processing points that have already been marked visible
			idx := uint32((p.Y-b.TL.Y)*b.Width() + (p.X - b.TL.X))
			if m.Visibility.Contains(idx) {
				continue
			}
			// Set this point as visible
			m.Visibility.Set(idx)
			// Set all neighbors as visible if they block vis this fixes wall
			// looking issues
			for dp.Y = p.Y - 1; dp.Y <= p.Y+1; dp.Y++ {
				if dp.Y < b.TL.Y || dp.Y > b.BR.Y {
					continue
				}
				for dp.X = p.X - 1; dp.X <= p.X+1; dp.X++ {
					if dp.X < b.TL.X || dp.X > b.BR.X {
						continue
					}
					c := m.GetChunk(dp)
					if c.BlocksVis.Contains(c.relOfs(dp)) {
						m.Visibility.Set(uint32((dp.Y-b.TL.Y)*b.Width() + (dp.X - b.TL.X)))
					}
				}
			}
		}
	}
	// Setup return values for reuse
	m.BitmapBounds = b
	m.Visibility.Clear()
	m.Remembered.Clear()
	// Sanity checks
	if !b.Contains(m.Player.Position) {
		return
	}
	// Mark the starting location as visible always
	p := m.Player.Position
	m.Visibility.Set(uint32((p.Y-b.TL.Y)*b.Width() + (p.X - b.TL.X)))
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
	for p.Y = b.TL.Y; p.Y <= b.BR.Y; p.Y++ {
		for p.X = b.TL.X; p.X <= b.BR.X; p.X++ {
			idx := uint32((p.Y-b.TL.Y)*b.Width() + (p.X - b.TL.X))
			c := m.GetChunk(p)
			cIdx := c.relOfs(p)
			if m.Visibility.Contains(idx) {
				c.HasSeen.Set(cIdx)
			}
			if c.HasSeen.Contains(cIdx) {
				m.Remembered.Set(idx)
			}
		}
	}
}

// CanSeePlayerFrom returns true if there is line of sight between the given
// point and the player.
func (m *CityMap) CanSeePlayerFrom(p util.Point) bool {
	// Position is on-screen, use visibility set
	if m.BitmapBounds.Contains(p) {
		idx := (p.Y-m.BitmapBounds.TL.Y)*m.BitmapBounds.Width() + (p.X - m.BitmapBounds.TL.X)
		return m.Visibility.Contains(uint32(idx))
	}
	// Position is off-screen, use a ray trace as the asymmetry won't be
	// noticeable
	for _, p := range util.Ray(p, m.Player.Position) {
		c := m.GetChunk(p)
		if c.BlocksVis.Contains(c.relOfs(p)) {
			return false
		}
	}
	return true
}

// Update updates the game world for d duration based around point p.
func (m *CityMap) Update(p util.Point, d time.Duration, update func()) {
	m.Now = m.Now.Add(d)
	// Updates of one minute or longer will use the wait handler automatically
	if d < time.Minute {
		m.updatePrepSets(p)
		m.updateShort(d)
		m.updateItemsAndPostProcessing(d)
	} else {
		m.Wait(d, update)
	}
}

func (m *CityMap) updatePrepSets(p util.Point) {
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
	m.loadBounds = lb.Multiply(ChunkWidth)
	m.updateBounds = ub.Multiply(ChunkWidth)
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
		c := m.Chunks[idx]
		for _, a := range c.Actors {
			heap.Remove(&m.aq, a.pqIdx)
		}
	}
	// Add actors in the new chunks to the priority queue and reset their think
	// times so the actors don't take a million turns when the chunk gets
	// reloaded after a long winter
	for _, idx := range m.usNewCache {
		c := m.Chunks[idx]
		for _, a := range c.Actors {
			if a.NextThink.Before(m.Now) {
				a.NextThink = m.Now
			}
			heap.Push(&m.aq, a)
		}
	}
}

// updateShort updates short-term updates for actors and vehicles.
func (m *CityMap) updateShort(d time.Duration) {
	// Process actor queue
	if len(m.aq) > 0 {
		for {
			a := heap.Pop(&m.aq).(*Actor)
			if m.Now.Before(a.NextThink) {
				heap.Push(&m.aq, a)
				break
			}
			if a.Dead {
				continue
			}
			d := a.AIModel.Act(a, m)
			a.NextThink = a.NextThink.Add(d)
			a.AIModel.PeriodicUpdate(a, m, d)
			heap.Push(&m.aq, a)
		}
	}
	// Update all vehicles within the update radius
	for _, v := range m.VehiclesWithin(m.updateBounds) {
		v.Update(d, m)
	}
}

// Wait advances time in the city by the given duration. During the first 90
// seconds of the duration the city will simulate as normal at an interval of
// one update per game second. Following that the simulation halts and only long
// term updates are executed for the remainder of the duration.
func (m *CityMap) Wait(d time.Duration, update func()) {
	sd := time.Second * 90
	if sd > d {
		sd = d
	}
	ld := d - sd
	for ; sd > 0; sd -= time.Second {
		ud := time.Second
		if sd < ud {
			ud = sd
		}
		m.updateShort(ud)
		if update != nil {
			update()
		}
	}
	if ld == 0 {
		return
	}
	for _, a := range m.aq {
		a.AIModel.PeriodicUpdate(a, m, ld)
		// Note that this assignment to the indexed property of the pq does not
		// cause chaos because every index is changed in the exact same way and
		// it does not alter the order of priority
		a.NextThink = a.NextThink.Add(ld)
	}
	m.updateItemsAndPostProcessing(d)
}

func (m *CityMap) updateItemsAndPostProcessing(d time.Duration) {
	var p util.Point
	var actorsToRemove []*Actor
	var itemsToRemove []*Item
	// Update all items in the update radius
	for p.Y = m.updateBounds.TL.Y; p.Y <= m.updateBounds.BR.Y; p.Y += ChunkHeight {
		for p.X = m.updateBounds.TL.X; p.X <= m.updateBounds.BR.X; p.X += ChunkWidth {
			c := m.GetChunk(p)
			for _, i := range c.Items {
				ExecuteItemUpdateEvent("Update", i, m, d)
			}
		}
	}
	// Update all items held by the player
	for idx, i := range m.Player.WornItems {
		if i == nil {
			continue
		}
		ExecuteItemUpdateEvent("Update", i, m, d)
		if i.Destroyed {
			m.Player.WornItems[idx] = nil
		}
	}
	if m.Player.Weapon != nil {
		ExecuteItemUpdateEvent("Update", m.Player.Weapon, m, d)
		if m.Player.Weapon.Destroyed {
			m.Player.Weapon = nil
		}
	}
	itemsToRemove = itemsToRemove[:0]
	for _, i := range m.Player.Inventory {
		if i == nil {
			continue
		}
		ExecuteItemUpdateEvent("Update", i, m, d)
		if i.Destroyed {
			itemsToRemove = append(itemsToRemove, i)
		}
	}
	for _, i := range itemsToRemove {
		m.Player.RemoveItemFromInventory(i)
	}
	// Update all items held by actors that were in the update area
	for p.Y = m.updateBounds.TL.Y; p.Y <= m.updateBounds.BR.Y; p.Y += ChunkHeight {
		for p.X = m.updateBounds.TL.X; p.X <= m.updateBounds.BR.X; p.X += ChunkWidth {
			c := m.GetChunk(p)
			for _, a := range c.Actors {
				for idx, i := range a.WornItems {
					if i == nil {
						continue
					}
					ExecuteItemUpdateEvent("Update", i, m, d)
					if i.Destroyed {
						a.WornItems[idx] = nil
					}
				}
				if a.Weapon != nil {
					ExecuteItemUpdateEvent("Update", a.Weapon, m, d)
					if a.Weapon.Destroyed {
						a.Weapon = nil
					}
				}
				itemsToRemove = itemsToRemove[:0]
				for _, i := range a.Inventory {
					if i == nil {
						continue
					}
					ExecuteItemUpdateEvent("Update", i, m, d)
					if i.Destroyed {
						itemsToRemove = append(itemsToRemove, i)
					}
				}
				for _, i := range itemsToRemove {
					m.Player.RemoveItemFromInventory(i)
				}
			}
		}
	}
	// Post-update cleanup of dead actors that need to turn into corpses and
	// destroyed items that need to be removed
	for p.Y = m.loadBounds.TL.Y; p.Y <= m.loadBounds.BR.Y; p.Y += ChunkHeight {
		for p.X = m.loadBounds.TL.X; p.X <= m.loadBounds.BR.X; p.X += ChunkWidth {
			// Remove dead actors
			actorsToRemove = actorsToRemove[:0]
			c := m.GetChunk(p)
			for _, a := range c.Actors {
				if a.Dead {
					a.DropCorpse(m)
					actorsToRemove = append(actorsToRemove, a)
				}
			}
			for _, a := range actorsToRemove {
				c.RemoveActor(a)
				idx := -1
				for i, qa := range m.aq {
					if qa == a {
						idx = i
						break
					}
				}
				if idx >= 0 {
					heap.Remove(&m.aq, idx)
				}
			}
			// Remove destroyed items
			itemsToRemove = itemsToRemove[:0]
			for _, i := range c.Items {
				if i.Destroyed {
					itemsToRemove = append(itemsToRemove, i)
				}
			}
			for _, i := range itemsToRemove {
				m.RemoveItem(i)
			}
		}
	}
}

// PlayerTookTurn is responsible for updating the city map model for the given
// duration as well as anything else that should happen after the player's turn.
func (m *CityMap) PlayerTookTurn(d time.Duration, update func()) {
	m.Player.TookTurn(m.Now, d)
	m.Update(m.Player.Position, d, update)
	// End conditions check
	if m.Player.Dead {
		Log.Log(termui.ColorRed, "YOU ARE DEAD! Press Escape to return to the main menu.")
	}
}

// FlagBitmapsForVehicle sets bitmaps dirty for all chunks occupied by the given
// vehicle.
func (m *CityMap) FlagBitmapsForVehicle(v *Vehicle, nb util.Rect) {
	b := v.Bounds.ContainingRect(nb)
	for _, c := range m.ChunksWithin(b) {
		c.bitmapsDirty = true
	}
}

// VehicleFits returns true if the vehicle fits within the given bounds.
func (m *CityMap) VehicleFits(v *Vehicle, nb util.Rect) bool {
	if nb.Area() != m.TileBounds.Overlap(nb).Area() {
		// Not totally within the map
		return false
	}
	// Consider tile matrix
	var p util.Point
	for p.Y = nb.TL.Y; p.Y <= nb.BR.Y; p.Y++ {
		for p.X = nb.TL.X; p.X <= nb.BR.X; p.X++ {
			t := m.GetTile(p)
			if t.BlocksWalk {
				return false
			}
		}
	}
	// Consider items
	for _, c := range m.ChunksWithin(nb) {
		for _, i := range c.Items {
			if nb.Contains(i.Position) && i.BlocksWalk {
				return false
			}
		}
	}
	// Consider vehicles
	for _, ov := range m.VehiclesWithin(nb) {
		if v == ov {
			continue
		}
		return false
	}
	return true
}

// MoveVehicle attempts to move the vehicle with the given offset.
func (m *CityMap) MoveVehicle(v *Vehicle, ofs util.Point) bool {
	// Try to move the vehicle
	ob := v.Bounds
	nb := v.Bounds.MoveRelative(ofs)
	if !m.VehicleFits(v, nb) {
		return false
	}
	oc := m.GetChunk(v.Bounds.TL)
	nc := m.GetChunk(nb.TL)
	if !oc.RemoveVehicle(v) {
		return false
	}
	v.Bounds = nb
	if !nc.PlaceVehicle(v) {
		v.Bounds = ob
		oc.PlaceVehicle(v)
		return false
	}
	// Move the player if they are within the vehicle
	if ob.Contains(m.Player.Position) {
		m.Player.Position = m.Player.Position.Add(ofs)
	}
	m.FlagBitmapsForVehicle(v, nb)
	return true
}
