package game

import "github.com/qbradq/after/lib/termui"

// TileDefs is the global TileRef-to-*TileDef reference.
var TileDefs = []*TileDef{
	{
		BackRef: 0,
		Name:    "error",
		Rune:    "!",
		Fg:      termui.ColorWhite,
		Bg:      termui.ColorRed,
	},
}

// TileRefs is the global string-to-TileRef reference.
var TileRefs = map[string]TileRef{
	"Error": 0,
}

// TileRef represents a three foot by three foot area of the world and is a
// reference into the global TileDefs slice.
type TileRef uint16

// TileDef represents all of the data associated with a single tile.
type TileDef struct {
	BackRef TileRef      // The TileRef that indexes this TileDef within TileDefs, used to accelerate saving
	Name    string       // Descriptive name of the tile
	Rune    string       // Map display rune
	Fg      termui.Color // Foreground display color
	Bg      termui.Color // Background display color
}
