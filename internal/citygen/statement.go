package citygen

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/util"
)

// evaluator evaluates a single expression executing its generation function.
type evaluator interface {
	// evaluate evaluates a single expression executing its function.
	evaluate(*game.Chunk, util.Point, time.Time)
}

// itemCreator generates new items based on the expression.
type itemCreator interface {
	// createItem generates the new item and returns it if any.
	createItem(time.Time) *game.Item
}

// tileExpression returns a fixed tile.
type tileExpression struct {
	r game.TileRef // Fixed tile reference
}

// evaluate implements the evaluator interface.
func (e *tileExpression) evaluate(c *game.Chunk, p util.Point, now time.Time) {
	c.Tiles[p.Y*game.ChunkWidth+p.X] = game.TileDefs[e.r]
}

// tileGenExpression returns the result of a tile generator.
type tileGenExpression struct {
	r TileGen // Tile generator to execute
}

// evaluate implements the evaluator interface.
func (e *tileGenExpression) evaluate(c *game.Chunk, p util.Point, now time.Time) {
	c.Tiles[p.Y*game.ChunkWidth+p.X] = e.r.Generate()
}

// itemExpression lays down a fixed item with a given chance.
type itemExpression struct {
	r    string // Item template name
	x, y int    // rng parameters
}

// evaluate implements the evaluator interface.
func (e *itemExpression) evaluate(c *game.Chunk, p util.Point, now time.Time) {
	if util.Random(0, e.y) < e.x {
		i := game.NewItem(e.r, now)
		i.Position = p
		c.PlaceItemRelative(i)
	}
}

// createItem returns just the new item, or nil if none.
func (e *itemExpression) createItem(now time.Time) *game.Item {
	if util.Random(0, e.y) < e.x {
		return game.NewItem(e.r, now)
	}
	return nil
}

// itemGenExpression lays down items with a given chance based on a generator.
type itemGenExpression struct {
	r    ItemGen // Item generator to execute
	x, y int     // rng parameters
}

// evaluate implements the evaluator interface.
func (e *itemGenExpression) evaluate(c *game.Chunk, p util.Point, now time.Time) {
	if util.Random(0, e.y) < e.x {
		i := e.r.Generate(now)
		i.Position = p
		c.PlaceItemRelative(i)
	}
}

// createItem returns just the new item, or nil if none.
func (e *itemGenExpression) createItem(now time.Time) *game.Item {
	if util.Random(0, e.y) < e.x {
		return e.r.Generate(now)
	}
	return nil
}

// actorExpression lays down an actor with a given chance.
type actorExpression struct {
	r    string // Actor template name
	x, y int    // rng parameters
}

func (e *actorExpression) evaluate(c *game.Chunk, p util.Point, now time.Time) {
	if util.Random(0, e.y) < e.x {
		a := game.NewActor(e.r, now)
		a.Position = p
		c.PlaceActorRelative(a)
	}
}

// actorGenExpression lays down an actor with a given chance based on a
// generator.
type actorGenExpression struct {
	r    ActorGen // Actor template name
	x, y int      // rng parameters
}

func (e *actorGenExpression) evaluate(c *game.Chunk, p util.Point, now time.Time) {
	if util.Random(0, e.y) < e.x {
		a := e.r.Generate(now)
		a.Position = p
		c.PlaceActorRelative(a)
	}
}

// genStatement is a list of expressions to run on a single position in the
// chunk at generation time. The text format of an expression is as follows:
// exp[;exp]... Where:
// exp = (tile_exp|item_exp)|(item_exp@XinY) Where:
// tile_name is the name of a tile or tile generator
// item_name is the name of an item or item generator
// Y is the bounded maximum of the half-open range [0-Y)
// X is the value of the random roll [0-Y) below which the item will appear
type genStatement []evaluator

func (s *genStatement) UnmarshalJSON(in []byte) error {
	es, err := parseStatement(in)
	if err != nil {
		return err
	}
	*s = append(*s, es...)
	return s.Validate()
}

// parseStatement parses a series of generator expressions from an input string
// and returns them.
func parseStatement(in []byte) ([]evaluator, error) {
	var ret []evaluator
	exprs := strings.Split(string(in[1:len(in)-1]), ";")
	for _, expr := range exprs {
		parts := strings.Split(expr, "@")
		switch len(parts) {
		case 1:
			if gen, found := ActorGens[parts[0]]; found {
				ret = append(ret, &actorGenExpression{
					r: gen,
					x: 1,
					y: 1,
				})
			} else if _, found := game.ActorDefs[parts[0]]; found {
				ret = append(ret, &actorExpression{
					r: parts[0],
					x: 1,
					y: 1,
				})
			} else if gen, found := ItemGens[parts[0]]; found {
				ret = append(ret, &itemGenExpression{
					r: gen,
					x: 1,
					y: 1,
				})
			} else if _, found := game.ItemDefs[parts[0]]; found {
				ret = append(ret, &itemExpression{
					r: parts[0],
					x: 1,
					y: 1,
				})
			} else if gen, found := TileGens[parts[0]]; found {
				ret = append(ret, &tileGenExpression{
					r: gen,
				})
			} else if r, found := game.TileRefs[parts[0]]; found {
				ret = append(ret, &tileExpression{
					r: r,
				})
			} else {
				return nil, fmt.Errorf("unresolved actor, item or tile reference %s", parts[0])
			}
		case 2:
			nParts := strings.Split(parts[1], "n")
			x, err := strconv.ParseInt(nParts[0], 0, 32)
			if err != nil {
				return nil, err
			}
			y, err := strconv.ParseInt(nParts[1], 0, 32)
			if err != nil {
				return nil, err
			}
			if gen, found := ActorGens[parts[0]]; found {
				ret = append(ret, &actorGenExpression{
					r: gen,
					x: int(x),
					y: int(y),
				})
			} else if _, found := game.ActorDefs[parts[0]]; found {
				ret = append(ret, &actorExpression{
					r: parts[0],
					x: int(x),
					y: int(y),
				})
			} else if gen, found := ItemGens[parts[0]]; found {
				ret = append(ret, &itemGenExpression{
					r: gen,
					x: int(x),
					y: int(y),
				})
			} else if _, found := game.ItemDefs[parts[0]]; found {
				ret = append(ret, &itemExpression{
					r: parts[0],
					x: int(x),
					y: int(y),
				})
			} else {
				return nil, fmt.Errorf("unresolved actor or item reference %s", parts[0])
			}
		default:
			return nil, fmt.Errorf("found %d parts, expected 1 or 2", len(parts))
		}
	}
	return ret, nil
}

// validateExpressionIntegrity returns an error if the given statement contains
// invalid parameters for any expression.
func validateExpressionIntegrity(s []evaluator) error {
	for _, exp := range s {
		switch e := exp.(type) {
		case *actorExpression:
			if e.x < 1 || e.y < 1 {
				return errors.New("generator statement actor expressions must use positive whole numbers")
			}
		case *actorGenExpression:
			if e.x < 1 || e.y < 1 {
				return errors.New("generator statement actor expressions must use positive whole numbers")
			}
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
	return nil
}

// Validate validates the generator statement.
func (s *genStatement) Validate() error {
	if err := validateExpressionIntegrity(*s); err != nil {
		return err
	}
	// Make sure we have exactly one tile expression in the statement
	tilesFound := 0
	for _, exp := range *s {
		switch exp.(type) {
		case *tileExpression:
			tilesFound++
		case *tileGenExpression:
			tilesFound++
		}
	}
	if tilesFound != 1 {
		return fmt.Errorf("generator statements must contain exactly 1 tile or tile generator")
	}
	return nil
}

// evaluate evaluates each expression in the statement in order.
func (s genStatement) evaluate(c *game.Chunk, p util.Point, t time.Time) {
	for _, exp := range s {
		exp.evaluate(c, p, t)
	}
}

// itemStatement is a generation statement that is only allowed to produce
// items.
type itemStatement []itemCreator

func (s *itemStatement) UnmarshalJSON(in []byte) error {
	es, err := parseStatement(in)
	if err != nil {
		return err
	}
	for _, e := range es {
		if ic, ok := e.(itemCreator); ok {
			*s = append(*s, ic)
		} else {
			return errors.New("item statement contained non-item expression")
		}
	}
	// No validation steps to perform that haven't already been done
	return nil
}

// evaluate evaluates each expression in the statement in order and returns the
// first item created.
func (s itemStatement) evaluate(t time.Time) *game.Item {
	for _, exp := range s {
		if i := exp.createItem(t); i != nil {
			return i
		}
	}
	return nil
}
