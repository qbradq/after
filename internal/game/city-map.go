package game

import (
	"bytes"
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
)

// Return slice for GetActors
var gaRet []*Actor

// CityMap represents the entire world of the game in terms of which chunks go
// where.
type CityMap struct {
	Bounds              util.Rect     // Bounds of the city map in chunks
	TileBounds          util.Rect     // Bounds of the city map in tiles
	Player              *Player       // Player actor
	Chunks              []*Chunk      // The chunks of the map
	InMemoryChunks      bitmap.Bitmap // Bitmap of all chunks loaded into memory
	InMemoryChunksCount int           // Running count of in-memory chunks to avoid excessive calls to bitmap.Count()
	ChunksGenerated     bitmap.Bitmap // Bitmap of all chunks that have been generated
	cgDirty             bool          // ChunksGenerated has been altered since the last call to SaveBitmaps
}

// NewCityMap allocates and returns a new CityMap structure.
func NewCityMap() *CityMap {
	m := &CityMap{
		Bounds:     util.NewRectWH(CityMapWidth, CityMapHeight),
		TileBounds: util.NewRectWH(CityMapWidth*ChunkWidth, CityMapHeight*ChunkHeight),
		Chunks:     make([]*Chunk, CityMapWidth*CityMapHeight),
	}
	for i := range m.Chunks {
		m.Chunks[i] = NewChunk(i%CityMapWidth, i/CityMapWidth, uint32(i))
	}
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
	for _, c := range m.Chunks {
		dict.Put(c.Generator.GetID())
	}
	// Write the file
	util.PutUint32(w, 0) // Version
	util.PutDictionary(w, dict)
	for _, c := range m.Chunks {
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
	for _, c := range m.Chunks {
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

// GetTile returns the tile at the given absolute tile point or nil if the point
// is out of bounds.
func (m *CityMap) GetTile(p util.Point) *TileDef {
	if !m.TileBounds.Contains(p) {
		return nil
	}
	c := m.Chunks[(p.Y/ChunkHeight)*CityMapWidth+(p.X/ChunkWidth)]
	t := c.Tiles[(p.Y%ChunkHeight)*ChunkWidth+(p.X%ChunkWidth)]
	return t
}

// EnsureLoaded ensures that all chunks in the area given in absolute tile
// coordinates have been generated and are loaded into memory.
func (m *CityMap) EnsureLoaded(area util.Rect) {
	// Calculate bounding area of the rect in terms of chunks and bound it.
	r := util.Rect{
		TL: util.Point{
			X: area.TL.X / ChunkWidth,
			Y: area.TL.Y / ChunkHeight,
		},
		BR: util.Point{
			X: area.BR.X / ChunkWidth,
			Y: area.BR.Y / ChunkHeight,
		},
	}
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
			c := m.Chunks[r]
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
	m.InMemoryChunks.Set(c.Ref)
	m.InMemoryChunksCount++
	// Allocate memory
	c.Tiles = make([]*TileDef, ChunkWidth*ChunkHeight)
	// Generate the chunk if this has never happened before
	if !m.ChunksGenerated.Contains(c.Ref) {
		m.ChunksGenerated.Set(c.Ref)
		m.cgDirty = true
		c.Generator.Generate(c)
		w := bytes.NewBuffer(nil)
		c.Write(w)
		SaveValue(fmt.Sprintf("Chunk-%d", c.Ref), w.Bytes())
		return
	}
	// Otherwise load the chunk into memory from the save database
	n := fmt.Sprintf("Chunk-%d", c.Ref)
	buf := LoadValue(n)
	c.Read(buf)
}

// purgeOldChunks purges chunks in least-recently-used first order down to the
// target number if the number of chunks in the memory cache is greater than the
// maximum.
func (m *CityMap) purgeOldChunks() {
	// Short-circuit condition
	if m.InMemoryChunksCount <= maxInMemoryChunks {
		return
	}
	// Sort the chunks by time last updated
	cRefs := make([]uint32, 0, m.InMemoryChunksCount)
	m.InMemoryChunks.Range(func(x uint32) {
		cRefs = append(cRefs, x)
	})
	sort.Slice(cRefs, func(i, j int) bool {
		return cRefs[i] < cRefs[j]
	})
	// Persist and unload the oldest chunks until we reach the purge target
	bufs := map[uint32][]byte{}
	for _, cr := range cRefs[:maxInMemoryChunks-purgeInMemoryChunksTarget] {
		w := bytes.NewBuffer(nil)
		c := m.Chunks[cr]
		c.Write(w)
		bufs[cr] = w.Bytes()
		c.Unload()
		m.InMemoryChunks.Remove(cr)
		m.InMemoryChunksCount--
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
	m.InMemoryChunks.Range(func(x uint32) {
		w := bytes.NewBuffer(nil)
		c := m.Chunks[x]
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
	m.ChunksGenerated.WriteTo(w)
	SaveValue("CityMap.ChunksGenerated", w.Bytes())
	m.cgDirty = false
}

// LoadBitmaps loads all persistent bitmaps.
func (m *CityMap) LoadBitmaps() {
	r := LoadValue("CityMap.ChunksGenerated")
	_ = util.GetUint32(r) // Version
	m.ChunksGenerated.ReadFrom(r)
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

// AddItem adds the item to the city at it's current location.
func (m *CityMap) AddItem(i *Item) {
	c := m.GetChunk(i.Position)
	c.Items = append(c.Items, i)
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

// ItemsWithin returns the items within the given bounds.
func (m *CityMap) ItemsWithin(b util.Rect) []*Item {
	var ret []*Item
	cb := util.NewRectXYWH(b.TL.X/ChunkWidth, b.TL.Y/ChunkHeight, b.Width()/ChunkWidth+1, b.Height()/ChunkHeight+1)
	for cy := cb.TL.Y; cy <= cb.BR.Y; cy++ {
		for cx := cb.TL.X; cx <= cb.BR.X; cx++ {
			c := m.Chunks[cy*CityMapWidth+cx]
			for _, i := range c.Items {
				if b.Contains(i.Position) {
					ret = append(ret, i)
				}
			}
		}
	}
	return ret
}

// AddActor adds the actor to the city at it's current location.
func (m *CityMap) AddActor(a *Actor) {
	c := m.GetChunk(a.Position)
	c.Actors = append(c.Actors, a)
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
	var ret []*Actor
	cb := util.NewRectXYWH(b.TL.X/ChunkWidth, b.TL.Y/ChunkHeight, b.Width()/ChunkWidth, b.Height()/ChunkHeight)
	for cy := cb.TL.Y; cy <= cb.BR.Y; cy++ {
		for cx := cb.TL.X; cx <= cb.BR.X; cx++ {
			c := m.Chunks[cy*CityMapWidth+cx]
			for _, a := range c.Actors {
				if b.Contains(a.Position) {
					ret = append(ret, a)
				}
			}
		}
	}
	return ret
}
