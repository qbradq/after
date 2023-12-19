package game

import (
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
	// Generate handles all procedural generation for the chunk
	Generate(*Chunk)
}

// ChunkFlag encodes various bit flags of a chunk.
type ChunkFlags uint8

const (
	ChunkFlagsNone      ChunkFlags = 0b00000000 // No flags enabled
	ChunkFlagsOccupied  ChunkFlags = 0b00000001 // Chunk was explicitly placed during the city gen process
	ChunkFlagsGenerated ChunkFlags = 0b00000010 // Chunk has already been generated
)

// Chunk represents the smallest unit of city planning and contains the tiles,
// items and actors within its bounds.
type Chunk struct {
	Generator         ChunkGen     // The chunk generator responsible for procedural generation
	ChunkGenOffset    util.Point   // Offset from the top-left corner of the chunk generator
	Facing            util.Facing  // Facing of the chunk during generation
	Name              string       // Descriptive name of the chunk
	MinimapRune       string       // Rune to display on the minimap
	MinimapForeground termui.Color // Foreground color of the rune on the minimap
	MinimapBackground termui.Color // Background color of the rune on the minimap
	Flags             ChunkFlags   // Flags
	Tiles             []*TileDef   // Tile matrix
	Loaded            bool         // If true Load() has been called and Unload() has not
}

// NewChunk allocates and returns a new Chunk struct. Note that this struct does
// *not* have the Generator field set yet and all of the tile pointers are nil.
// See Load().
func NewChunk() *Chunk {
	c := &Chunk{
		Name:              "an error",
		MinimapRune:       ".",
		MinimapForeground: termui.ColorWhite,
		MinimapBackground: termui.ColorBlack,
	}
	return c
}

// Load ensures that the chunk is fully loaded from the chunk data and ready to
// use. This is a cheap operation for chunks that are already loaded.
func (c *Chunk) Load() {
	if c.Loaded {
		return
	}
	c.Tiles = make([]*TileDef, ChunkWidth*ChunkHeight)
	if c.Flags&ChunkFlagsGenerated == 0 {
		c.Generator.Generate(c)
		c.Flags |= ChunkFlagsGenerated
	}
	c.Loaded = true
}
