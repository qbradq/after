package game

import (
	"io"
	"time"

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
	Generate(*Chunk)
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
	Ref               uint32       // Reference index for the chunk
	Bounds            util.Rect    // Bounds of the chunk
	Generator         ChunkGen     // The chunk generator responsible for procedural generation
	ChunkGenOffset    util.Point   // Offset from the top-left corner of the chunk generator
	Facing            util.Facing  // Facing of the chunk during generation
	Name              string       // Descriptive name of the chunk
	MinimapRune       string       // Rune to display on the minimap
	MinimapForeground termui.Color // Foreground color of the rune on the minimap
	MinimapBackground termui.Color // Background color of the rune on the minimap
	Flags             ChunkFlags   // Flags
	Loaded            time.Time    // Time this chunk was loaded, the zero value means it is not in memory
	Tiles             []*TileDef   // Tile matrix
	Items             []*Item      // All items within the chunk
	Actors            []*Actor     // All actors within the chunk
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
	c.Items = append(c.Items, i)
}
