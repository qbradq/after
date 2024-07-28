package game

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/qbradq/after/lib/util"
)

// Evaluator evaluates a single expression executing its generation function.
type Evaluator interface {
	// Evaluate evaluates a single expression executing its function.
	Evaluate(*Chunk, util.Point, time.Time)
}

// ItemsCreator generates new items based on the expression.
type ItemsCreator interface {
	// CreateItems generates the new items, if any, and returns them in a non-
	// nil slice.
	CreateItems(time.Time) []*Item
}

// tileExpression returns a fixed tile.
type tileExpression struct {
	r TileRef // Fixed tile reference
}

// Evaluate implements the evaluator interface.
func (e *tileExpression) Evaluate(c *Chunk, p util.Point, now time.Time) {
	c.Tiles[p.Y*ChunkWidth+p.X] = TileDefs[e.r]
}

// tileGenExpression returns the result of a tile generator.
type tileGenExpression struct {
	r TileGen // Tile generator to execute
}

// Evaluate implements the evaluator interface.
func (e *tileGenExpression) Evaluate(c *Chunk, p util.Point, now time.Time) {
	c.Tiles[p.Y*ChunkWidth+p.X] = e.r.Generate()
}

// itemExpression lays down a fixed item with a given chance.
type itemExpression struct {
	r    string // Item template name
	x, y int    // rng parameters
	n    int    // Number of repetitions
}

// Evaluate implements the evaluator interface.
func (e *itemExpression) Evaluate(c *Chunk, p util.Point, now time.Time) {
	for _, i := range e.CreateItems(now) {
		i.Position = p
		c.PlaceItemRelative(i)
	}
}

// CreateItems implements the ItemsCreator interface.
func (e *itemExpression) CreateItems(now time.Time) []*Item {
	ret := []*Item{}
	for pass := 0; pass < e.n; pass++ {
		if util.Random(0, e.y) < e.x {
			i := NewItem(e.r, now, true)
			if i.Amount < 1 {
				i.Amount = 1
			}
			ret = append(ret, i)
		}
	}
	return ret
}

// itemGenExpression lays down items with a given chance based on a generator.
type itemGenExpression struct {
	r    ItemGen // Item generator to execute
	x, y int     // rng parameters
	n    int     // Number of repetitions
}

// Evaluate implements the evaluator interface.
func (e *itemGenExpression) Evaluate(c *Chunk, p util.Point, now time.Time) {
	for _, i := range e.CreateItems(now) {
		i.Position = p
		c.PlaceItemRelative(i)
	}
}

// CreateItems implements the ItemsCreator interface.
func (e *itemGenExpression) CreateItems(now time.Time) []*Item {
	ret := []*Item{}
	for pass := 0; pass < e.n; pass++ {
		if util.Random(0, e.y) < e.x {
			i := e.r.Generate(now)
			if i.Amount < 1 {
				i.Amount = 1
			}
			ret = append(ret, i)
		}
	}
	return ret
}

// actorExpression lays down an actor with a given chance.
type actorExpression struct {
	r    string // Actor template name
	x, y int    // rng parameters
}

// Evaluate implements the evaluator interface.
func (e *actorExpression) Evaluate(c *Chunk, p util.Point, now time.Time) {
	if util.Random(0, e.y) < e.x {
		a := NewActor(e.r, now, true)
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

// Evaluate implements the evaluator interface.
func (e *actorGenExpression) Evaluate(c *Chunk, p util.Point, now time.Time) {
	if util.Random(0, e.y) < e.x {
		a := e.r.Generate(now)
		a.Position = p
		c.PlaceActorRelative(a)
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
type GenStatement []Evaluator

func (s *GenStatement) UnmarshalJSON(in []byte) error {
	es, err := parseStatement(in)
	if err != nil {
		return err
	}
	*s = append(*s, es...)
	return s.Validate()
}

// parseStatement parses a series of generator expressions from an input string
// and returns them.
func parseStatement(in []byte) ([]Evaluator, error) {
	var ret []Evaluator
	exprs := strings.Split(string(in[1:len(in)-1]), ";")
	for _, expr := range exprs {
		// "Times" symbol handling
		n := 1
		parts := strings.Split(expr, "*")
		if len(parts) == 2 {
			v, err := strconv.ParseInt(parts[1], 0, 64)
			if err != nil {
				return nil, err
			}
			n = int(v)
		} else if len(parts) != 1 {
			return nil, fmt.Errorf("only one '*' symbol allowed per expression")
		}
		if n > 1024 {
			return nil, fmt.Errorf("excessive n value, capped at 1024")
		}
		// Chance expression handling
		parts = strings.Split(parts[0], "@")
		switch len(parts) {
		case 1:
			if gen, found := ActorGens[parts[0]]; found {
				if n > 1 {
					return nil, fmt.Errorf("'*' symbol not allowed in actor or tile expressions")
				}
				ret = append(ret, &actorGenExpression{
					r: gen,
					x: 1,
					y: 1,
				})
			} else if _, found := ActorDefs[parts[0]]; found {
				if n > 1 {
					return nil, fmt.Errorf("'*' symbol not allowed in actor or tile expressions")
				}
				ret = append(ret, &actorExpression{
					r: parts[0],
					x: 1,
					y: 1,
				})
			} else if gen, found := ItemGens[parts[0]]; found {
				ret = append(ret, &itemGenExpression{
					r: gen,
					x: 1,
					y: n,
					n: n,
				})
			} else if _, found := ItemDefs[parts[0]]; found {
				ret = append(ret, &itemExpression{
					r: parts[0],
					x: 1,
					y: n,
					n: n,
				})
			} else if gen, found := TileGens[parts[0]]; found {
				if n > 1 {
					return nil, fmt.Errorf("'*' symbol not allowed in actor or tile expressions")
				}
				ret = append(ret, &tileGenExpression{
					r: gen,
				})
			} else if r, found := TileRefs[parts[0]]; found {
				if n > 1 {
					return nil, fmt.Errorf("'*' symbol not allowed in actor or tile expressions")
				}
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
				if n > 1 {
					return nil, fmt.Errorf("'*' symbol not allowed in actor or tile expressions")
				}
				ret = append(ret, &actorGenExpression{
					r: gen,
					x: int(x),
					y: int(y),
				})
			} else if _, found := ActorDefs[parts[0]]; found {
				if n > 1 {
					return nil, fmt.Errorf("'*' symbol not allowed in actor or tile expressions")
				}
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
					n: n,
				})
			} else if _, found := ItemDefs[parts[0]]; found {
				ret = append(ret, &itemExpression{
					r: parts[0],
					x: int(x),
					y: int(y),
					n: n,
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
func validateExpressionIntegrity(s []Evaluator) error {
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
func (s *GenStatement) Validate() error {
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

// Evaluate evaluates each expression in the statement in order.
func (s GenStatement) Evaluate(c *Chunk, p util.Point, t time.Time) {
	for _, exp := range s {
		exp.Evaluate(c, p, t)
	}
}

// ItemStatement is a generation statement that is only allowed to produce
// items.
type ItemStatement []ItemsCreator

func (s *ItemStatement) UnmarshalJSON(in []byte) error {
	es, err := parseStatement(in)
	if err != nil {
		return err
	}
	for _, e := range es {
		if ic, ok := e.(ItemsCreator); ok {
			*s = append(*s, ic)
		} else {
			return errors.New("item statement contained non-item expression")
		}
	}
	// No validation steps to perform that haven't already been done
	return nil
}

// Evaluate evaluates each expression in the statement in order and returns the
// items generated by the first item expression, if any, in a non-nil slice.
func (s ItemStatement) Evaluate(t time.Time) []*Item {
	ret := []*Item{}
	for _, exp := range s {
		items := exp.CreateItems(t)
		if len(items) > 0 {
			ret = items
			break
		}
	}
	return ret
}
