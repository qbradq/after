package game

import (
	"io"
	"time"

	"github.com/kelindar/bitmap"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

const (
	ChunkWidth  int = 16 // Width of a chunk in tiles
	ChunkHeight int = 16 // Height of a chunk in tiles
)

// ChunkGen is the interface all chunk generators must implement for the game
// library to interact with them.
type ChunkGen interface {
	// Generate handles all initial procedural generation for the chunk
	Generate(*Chunk, *CityMap)
	// AssignStaticInfo inserts all of the non-procedurally generated bits into
	// the chunk, such as name and map rune.
	AssignStaticInfo(*Chunk)
	// GetID returns the unique ID of the generator.
	GetID() string
}

// ChunkFlag encodes various bit flags of a chunk.
type ChunkFlags uint8

const (
	ChunkFlagsNone     ChunkFlags = 0b00000000 // No flags enabled
	ChunkFlagsOccupied ChunkFlags = 0b00000001 // Chunk was explicitly placed during the city gen process
)

// Chunk represents the smallest unit of city planning and contains the tiles,
// items and actors within its bounds.
type Chunk struct {
	Ref               uint32        // Reference index for the chunk
	Bounds            util.Rect     // Bounds of the chunk
	Generator         ChunkGen      // The chunk generator responsible for procedural generation
	ChunkGenOffset    util.Point    // Offset from the top-left corner of the chunk generator
	Facing            util.Facing   // Facing of the chunk during generation
	Name              string        // Descriptive name of the chunk
	MinimapRune       string        // Rune to display on the minimap
	MinimapForeground termui.Color  // Foreground color of the rune on the minimap
	MinimapBackground termui.Color  // Background color of the rune on the minimap
	Flags             ChunkFlags    // Flags
	Loaded            time.Time     // Time this chunk was loaded, the zero value means it is not in memory
	Tiles             []*TileDef    // Tile matrix
	Items             []*Item       // All items within the chunk
	Actors            []*Actor      // All actors within the chunk
	HasSeen           bitmap.Bitmap // Bitmap of all spaces that have been previously viewed by the player
	BlocksWalk        bitmap.Bitmap // Bitmap of all spaces that are blocked for walking
	BlocksVis         bitmap.Bitmap // Bitmap of all spaces that are blocked for visibility
	bitmapsDirty      bool          // If true the BlocksWalk and BlocksVis bitmaps need to be rebuilt before use

}

// NewChunk allocates and returns a new Chunk struct. Note that this struct does
// *not* have the Generator field set yet and all of the tile pointers are nil.
// See Load().
func NewChunk(x, y int, r uint32) *Chunk {
	c := &Chunk{
		Ref:               r,
		Bounds:            util.NewRectXYWH(x*ChunkWidth, y*ChunkHeight, ChunkWidth, ChunkHeight),
		Name:              "an error",
		MinimapRune:       "!",
		MinimapForeground: termui.ColorWhite,
		MinimapBackground: termui.ColorRed,
	}
	return c
}

func (c *Chunk) relOfs(p util.Point) uint32 {
	return uint32((p.Y-c.Bounds.TL.Y)*ChunkWidth + (p.X - c.Bounds.TL.X))
}

// Write writes the chunk to w.
func (c *Chunk) Write(w io.Writer) {
	util.PutUint32(w, 0)        // Version
	for _, t := range c.Tiles { // Tile map
		util.PutUint16(w, uint16(getTileCrossRef(t.BackRef)))
	}
	util.PutUint16(w, uint16(len(c.Items))) // Number of items
	for _, i := range c.Items {             // Items
		i.Write(w)
	}
	util.PutUint16(w, uint16(len(c.Actors))) // Number of actors
	for _, a := range c.Actors {             // Actors
		a.Write(w)
	}
	c.HasSeen.WriteTo(w)
}

// Unload frees chunk-level persistent memory
func (c *Chunk) Unload() {
	c.Tiles = nil
	c.Items = nil
	c.Actors = nil
	c.Loaded = time.Time{}
}

// Read allocates memory and reads the chunk from r.
func (c *Chunk) Read(r io.Reader) {
	c.Tiles = make([]*TileDef, ChunkWidth*ChunkHeight)
	_ = util.GetUint32(r)    // Version
	for i := range c.Tiles { // Tile map
		x := TileCrossRef(util.GetUint16(r))
		c.Tiles[i] = TileCrossRefs[x]
	}
	n := int(util.GetUint16(r)) // Number of items
	for i := 0; i < n; i++ {    // Items
		c.Items = append(c.Items, NewItemFromReader(r))
	}
	n = int(util.GetUint16(r)) // Number of actors
	for i := 0; i < n; i++ {   // Actors
		c.Actors = append(c.Actors, NewActorFromReader(r))
	}
	c.HasSeen.ReadFrom(r)
}

// RebuildBitmaps must be called after chunk load or generation in order to
// rebuild the walk and vis bitmap caches. The HasSeen bitmap is persistent.
func (c *Chunk) RebuildBitmaps() {
	// Clear bitmaps
	c.BlocksVis.Clear()
	c.BlocksWalk.Clear()
	// Consider tiles
	for i, t := range c.Tiles {
		if t.BlocksVis {
			c.BlocksVis.Set(uint32(i))
		}
		if t.BlocksWalk {
			c.BlocksWalk.Set(uint32(i))
		}
	}
	// Consider items
	for _, i := range c.Items {
		if i.BlocksVis {
			c.BlocksVis.Set(c.relOfs(i.Position))
		}
		if i.BlocksWalk {
			c.BlocksWalk.Set(c.relOfs(i.Position))
		}
	}
	// Consider actors
	for _, a := range c.Actors {
		c.BlocksWalk.Set(c.relOfs(a.Position))
	}
	c.bitmapsDirty = false
}

// PlaceItemRelative adds the item to the chunk and adjusts the
// position from chunk-relative to absolute.
func (c *Chunk) PlaceItemRelative(i *Item) {
	i.Position.X += c.Bounds.TL.X
	i.Position.Y += c.Bounds.TL.Y
	c.PlaceItem(i)
}

// PlaceItem places the item within the chunk. This is a no-op if the item's
// current position lies outside the chunk.
func (c *Chunk) PlaceItem(i *Item) {
	if !c.Bounds.Contains(i.Position) {
		return
	}
	if i.BlocksVis {
		c.BlocksVis.Set(c.relOfs(i.Position))
	}
	if i.BlocksWalk {
		c.BlocksWalk.Set(c.relOfs(i.Position))
	}
	c.Items = append(c.Items, i)
}

// RemoveItem removes the item from the chunk. This is a no-op if the item's
// current position lies outside the chunk.
func (c *Chunk) RemoveItem(i *Item) {
	if !c.Bounds.Contains(i.Position) {
		return
	}
	idx := -1
	for n, item := range c.Items {
		if item == i {
			idx = n
			break
		}
	}
	if idx < 0 {
		return
	}
	// Remove from slice while maintaining order
	copy(c.Items[idx:], c.Items[idx+1:])
	c.Items[len(c.Items)-1] = nil
	c.Items = c.Items[:len(c.Items)-1]
	// Bitmap updates
	if i.BlocksVis {
		c.bitmapsDirty = true
	}
	if i.BlocksWalk {
		c.bitmapsDirty = true
	}
}

// PlaceActorRelative adds the actor to the chunk and adjusts the
// position from chunk-relative to absolute.
func (c *Chunk) PlaceActorRelative(a *Actor) {
	a.Position.X += c.Bounds.TL.X
	a.Position.Y += c.Bounds.TL.Y
	c.PlaceActor(a)
}

// PlaceActor places the actor within the chunk. This is a no-op if the
// current position lies outside the chunk. Returns true on success.
func (c *Chunk) PlaceActor(a *Actor) bool {
	if !c.CanStep(a, a.Position) {
		return false
	}
	c.Actors = append(c.Actors, a)
	c.BlocksWalk.Set(c.relOfs(a.Position))
	return true
}

// RemoveActor removes the Actor from the chunk. This is a no-op if the
// current position lies outside the chunk.
func (c *Chunk) RemoveActor(a *Actor) {
	if !c.Bounds.Contains(a.Position) {
		return
	}
	idx := -1
	for n, actor := range c.Actors {
		if actor == a {
			idx = n
			break
		}
	}
	if idx < 0 {
		return
	}
	// Remove from slice, order is not important
	c.Actors[idx] = c.Actors[len(c.Actors)-1]
	c.Actors[len(c.Actors)-1] = nil
	c.Actors = c.Actors[:len(c.Actors)-1]
	c.bitmapsDirty = true
}

// CanStep returns true if the location is valid for an actor.
func (c *Chunk) CanStep(a *Actor, p util.Point) bool {
	if !c.Bounds.Contains(p) {
		return false
	}
	if c.bitmapsDirty {
		c.RebuildBitmaps()
	}
	if c.BlocksWalk.Contains(c.relOfs(p)) {
		return false
	}
	for _, a := range c.Actors {
		if a.Position == p {
			return false
		}
	}
	return true
}
