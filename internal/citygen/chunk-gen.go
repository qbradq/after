package citygen

import (
	"fmt"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

func init() {
	game.GetChunkGen = func(s, v string) game.ChunkGen {
		return ChunkGenGroups[s].Variants[v]
	}
}

// ChunkGenGroup represents a group of chunk generators.
type ChunkGenGroup struct {
	ID          string               // ID of the group.
	Variants    map[string]*ChunkGen // Map of chunk gens by variant name.
	VariantList []*ChunkGen          // List of chunk gens
}

// NewChunkGenGroup creates a new ChunkGenGroup ready for use.
func NewChunkGenGroup(id string) *ChunkGenGroup {
	return &ChunkGenGroup{
		ID:       id,
		Variants: map[string]*ChunkGen{},
	}
}

// Add adds a variant to the group.
func (g *ChunkGenGroup) Add(v *ChunkGen) error {
	if _, duplicate := g.Variants[v.Variant]; duplicate {
		return fmt.Errorf("duplicate variant %s in chunk gen group %s", v.Variant, g.ID)
	}
	g.Variants[v.Variant] = v
	g.VariantList = append(g.VariantList, v)
	return nil
}

// Get returns a pointer to one of the variant chunk generators at random.
func (g *ChunkGenGroup) Get() *ChunkGen {
	if len(g.VariantList) < 1 {
		return nil
	}
	return g.VariantList[util.Random(0, len(g.VariantList))]
}

// Global chunk generators index
var ChunkGenGroups = map[string]*ChunkGenGroup{}

// ChunkGen defines what the chunk looks like on the city map and controls tile,
// item and actor placements. Note that the zero value is *not sane*. Only get
// ChunkGen objects from the ChunkGens map.
type ChunkGen struct {
	Group   string                       // ID of the group this chunk generator belongs to
	Variant string                       // Name of the variant, must be unique within the group
	Name    string                       // Descriptive name of the chunk
	Width   int                          // Width of the chunk generator in chunks
	Height  int                          // Height of the chunk generator in chunks
	Minimap []string                     // Minimap
	Fg      termui.Color                 // Foreground color
	Bg      termui.Color                 // Background color
	Map     []string                     // Map of characters that define how to procedurally generate each tile, the map is selected at random
	Tiles   map[string]game.GenStatement // Mapping of map characters to value generator statements
}

// GetGroup returns the unique identifier of the generator.
func (g *ChunkGen) GetGroup() string { return g.Group }

// GetVariant returns the variant name.
func (g *ChunkGen) GetVariant() string { return g.Variant }

// AssignStaticInfo inserts all the static chunk info.
func (g *ChunkGen) AssignStaticInfo(c *game.Chunk) {
	c.Name = g.Name
	c.MinimapForeground = g.Fg
	c.MinimapBackground = g.Bg
	c.MinimapRune = string(g.Minimap[c.ChunkGenOffset.Y][c.ChunkGenOffset.X])
}

// Generate handles all of the procedural generation for the chunk.
func (g *ChunkGen) Generate(c *game.Chunk, m *game.CityMap) {
	var sp util.Point
	var dp util.Point
	cb := util.NewRectWH(game.ChunkWidth, game.ChunkHeight)
	// Generate tile matrix
	for sp.Y = c.ChunkGenOffset.Y * game.ChunkHeight; sp.Y < (c.ChunkGenOffset.Y+1)*game.ChunkHeight; sp.Y++ {
		dp.X = 0
		for sp.X = c.ChunkGenOffset.X * game.ChunkWidth; sp.X < (c.ChunkGenOffset.X+1)*game.ChunkWidth; sp.X++ {
			r := string(g.Map[sp.Y][sp.X])
			rp := cb.RotatePointRelative(dp, c.Facing)
			g.Tiles[r].Tile.Evaluate(c, rp, m)
			dp.X++
		}
		dp.Y++
	}
	// Generate vehicles
	dp = util.Point{}
	for sp.Y = c.ChunkGenOffset.Y * game.ChunkHeight; sp.Y < (c.ChunkGenOffset.Y+1)*game.ChunkHeight; sp.Y++ {
		dp.X = 0
		for sp.X = c.ChunkGenOffset.X * game.ChunkWidth; sp.X < (c.ChunkGenOffset.X+1)*game.ChunkWidth; sp.X++ {
			r := string(g.Map[sp.Y][sp.X])
			gen := g.Tiles[r].Vehicle
			if gen == nil {
				dp.X++
				continue
			}
			rp := cb.RotatePointRelative(dp, c.Facing)
			gen.Evaluate(c, rp, m)
			dp.X++
		}
		dp.Y++
	}
	// Generate items
	dp = util.Point{}
	for sp.Y = c.ChunkGenOffset.Y * game.ChunkHeight; sp.Y < (c.ChunkGenOffset.Y+1)*game.ChunkHeight; sp.Y++ {
		dp.X = 0
		for sp.X = c.ChunkGenOffset.X * game.ChunkWidth; sp.X < (c.ChunkGenOffset.X+1)*game.ChunkWidth; sp.X++ {
			r := string(g.Map[sp.Y][sp.X])
			rp := cb.RotatePointRelative(dp, c.Facing)
			for _, gen := range g.Tiles[r].Items {
				gen.Evaluate(c, rp, m)
			}
			dp.X++
		}
		dp.Y++
	}
	// Generate actors
	dp = util.Point{}
	for sp.Y = c.ChunkGenOffset.Y * game.ChunkHeight; sp.Y < (c.ChunkGenOffset.Y+1)*game.ChunkHeight; sp.Y++ {
		dp.X = 0
		for sp.X = c.ChunkGenOffset.X * game.ChunkWidth; sp.X < (c.ChunkGenOffset.X+1)*game.ChunkWidth; sp.X++ {
			r := string(g.Map[sp.Y][sp.X])
			gen := g.Tiles[r].Actor
			if gen == nil {
				dp.X++
				continue
			}
			rp := cb.RotatePointRelative(dp, c.Facing)
			gen.Evaluate(c, rp, m)
			dp.X++
		}
		dp.Y++
	}
}
