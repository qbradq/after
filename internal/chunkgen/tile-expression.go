package chunkgen

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/internal/itemgen"
	"github.com/qbradq/after/internal/tilegen"
	"github.com/qbradq/after/lib/util"
)

// evaluator evaluates a single expression executing its generation function.
type evaluator interface {
	// evaluate evaluates a single expression executing its function.
	evaluate(*game.Chunk, util.Point)
}

// tileExpression returns a fixed tile.
type tileExpression struct {
	r game.TileRef // Fixed tile reference
}

// evaluate implements the evaluator interface.
func (e *tileExpression) evaluate(c *game.Chunk, p util.Point) {
	c.Tiles[p.Y*game.ChunkWidth+p.X] = game.TileDefs[e.r]
}

// tileGenExpression returns the result of a tile generator.
type tileGenExpression struct {
	r tilegen.TileGen // Tile generator to execute
}

// evaluate implements the evaluator interface.
func (e *tileGenExpression) evaluate(c *game.Chunk, p util.Point) {
	c.Tiles[p.Y*game.ChunkWidth+p.X] = e.r.Generate()
}

// itemExpression lays down a fixed item with a given chance.
type itemExpression struct {
	r    string // Item template name
	x, y int    // rng parameters
}

// evaluate implements the evaluator interface.
func (e *itemExpression) evaluate(c *game.Chunk, p util.Point) {
	if util.Random(0, e.y) < e.x {
		i := game.NewItem(e.r)
		i.Position = p
		c.PlaceItemRelative(i)
	}
}

// itemGenExpression lays down items with a given chance based on a generator.
type itemGenExpression struct {
	r    itemgen.ItemGen // Item generator to execute
	x, y int             // rng parameters
}

// evaluate implements the evaluator interface.
func (e *itemGenExpression) evaluate(c *game.Chunk, p util.Point) {
	if util.Random(0, e.y) < e.x {
		i := e.r.Generate()
		i.Position = p
		c.PlaceItemRelative(i)
	}
}

// GenStatement is a list of expressions to run on a single position in the
// chunk at generation time. The text format of an expression is as follows:
// exp[;exp]... Where:
// exp = (tile_exp|item_exp)|(item_exp@XinY) Where:
// tile_name is the name of a tile or tile generator
// item_name is the name of an item or item generator
// Y is the bounded maximum of the half-open range [0-Y)
// X is the value of the random roll [0-Y) below which the item will appear
type GenStatement []evaluator

func (s *GenStatement) UnmarshalJSON(in []byte) error {
	// Expression parsing
	exprs := strings.Split(string(in[1:len(in)-1]), ";")
	for _, expr := range exprs {
		parts := strings.Split(expr, "@")
		switch len(parts) {
		case 1:
			if gen, found := itemgen.ItemGens[parts[0]]; found {
				*s = append(*s, &itemGenExpression{
					r: gen,
					x: 1,
					y: 1,
				})
			} else if _, found := game.ItemDefs[parts[0]]; found {
				*s = append(*s, &itemExpression{
					r: parts[0],
					x: 1,
					y: 1,
				})
			} else if gen, found := tilegen.TileGens[parts[0]]; found {
				*s = append(*s, &tileGenExpression{
					r: gen,
				})
			} else if r, found := game.TileRefs[parts[0]]; found {
				*s = append(*s, &tileExpression{
					r: r,
				})
			} else {
				return fmt.Errorf("unresolved item or tile reference %s", parts[0])
			}
		case 2:
			nParts := strings.Split(parts[1], "n")
			x, err := strconv.ParseInt(nParts[0], 0, 32)
			if err != nil {
				return err
			}
			y, err := strconv.ParseInt(nParts[1], 0, 32)
			if err != nil {
				return err
			}
			if gen, found := itemgen.ItemGens[parts[0]]; found {
				*s = append(*s, &itemGenExpression{
					r: gen,
					x: int(x),
					y: int(y),
				})
			} else if _, found := game.ItemDefs[parts[0]]; found {
				*s = append(*s, &itemExpression{
					r: parts[0],
				})
			} else {
				return fmt.Errorf("unresolved item reference %s", parts[0])
			}
		default:
			return fmt.Errorf("found %d parts, expected 1 or 2", len(parts))
		}
	}
	// Validate the statement
	tilesFound := 0
	for _, exp := range *s {
		switch e := exp.(type) {
		case *tileExpression:
			tilesFound++
		case *tileGenExpression:
			tilesFound++
		case *itemExpression:
			if e.x < 1 || e.y < 1 {
				return errors.New("generator statement item expressions must use positive whole numbers")
			}
		case *itemGenExpression:
			if e.x < 1 || e.y < 1 {
				return errors.New("generator statement item expressions must use positive whole numbers")
			}
		}
	}
	if tilesFound != 1 {
		return fmt.Errorf("generator statements must contain exactly 1 tile or tile generator")
	}
	return nil
}

// evaluate evaluates each expression in the statement in order.
func (s GenStatement) evaluate(c *game.Chunk, p util.Point) {
	for _, exp := range s {
		exp.evaluate(c, p)
	}
}
