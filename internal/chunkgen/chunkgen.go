package chunkgen

import (
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/internal/tilegen"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

func init() {
	game.GetChunkGen = func(s string) game.ChunkGen {
		return ChunkGens[s]
	}
}

// Global chunk generators index
var ChunkGens = map[string]*ChunkGen{}

// ChunkGen defines what the chunk looks like on the city map and controls tile,
// item and actor placements. Note that the zero value is *not sane*. Only get
// ChunkGen objects from the ChunkGens map.
type ChunkGen struct {
	ID      string            // Unique id of the generator
	Name    string            // Descriptive name of the chunk
	Width   int               // Width of the chunk generator in chunks
	Height  int               // Height of the chunk generator in chunks
	Minimap []string          // Minimap
	Fg      termui.Color      // Foreground color
	Bg      termui.Color      // Background color
	Map     []string          // Map of characters that define how to procedurally generate each tile
	Tiles   map[string]string // Mapping of map characters to tile generators
}

// GetID returns the unique identifier of the generator.
func (g *ChunkGen) GetID() string { return g.ID }

// AssignStaticInfo inserts all the static chunk info.
func (g *ChunkGen) AssignStaticInfo(c *game.Chunk) {
	c.Name = g.Name
	c.MinimapForeground = g.Fg
	c.MinimapBackground = g.Bg
	c.MinimapRune = string(g.Minimap[c.ChunkGenOffset.Y][c.ChunkGenOffset.X])
}

// Generate handles all of the procedural generation for the chunk.
func (g *ChunkGen) Generate(c *game.Chunk) {
	fn := func(p util.Point, f util.Facing) util.Point {
		switch f {
		case util.FacingNorth:
			return p
		case util.FacingEast:
			return util.Point{
				X: p.Y,
				Y: p.X,
			}
		case util.FacingSouth:
			return util.Point{
				X: (game.ChunkWidth - 1) - p.X,
				Y: (game.ChunkHeight - 1) - p.Y,
			}
		default:
			return util.Point{
				X: p.Y,
				Y: (game.ChunkWidth - 1) - p.X,
			}
		}
	}
	var sp util.Point
	var dp util.Point
	for sp.Y = c.ChunkGenOffset.Y * game.ChunkHeight; sp.Y < (c.ChunkGenOffset.Y+1)*game.ChunkHeight; sp.Y++ {
		dp.X = 0
		for sp.X = c.ChunkGenOffset.X * game.ChunkWidth; sp.X < (c.ChunkGenOffset.X+1)*game.ChunkWidth; sp.X++ {
			r := string(g.Map[sp.Y][sp.X])
			gn := g.Tiles[r]
			rp := fn(dp, c.Facing)
			if tg, found := tilegen.TileGens[gn]; found {
				c.Tiles[rp.Y*game.ChunkWidth+rp.X] = tg.Generate()
			} else if t, found := game.TileRefs[gn]; found {
				c.Tiles[rp.Y*game.ChunkWidth+rp.X] = game.TileDefs[t]
			}
			dp.X++
		}
		dp.Y++
	}
}
