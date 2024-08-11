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
	Evaluate(*Chunk, util.Point, *CityMap)
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
func (e *tileExpression) Evaluate(c *Chunk, p util.Point, cm *CityMap) {
	c.Tiles[p.Y*ChunkWidth+p.X] = TileDefs[e.r]
}

// tileGenExpression returns the result of a tile generator.
type tileGenExpression struct {
	r TileGen // Tile generator to execute
}

// Evaluate implements the evaluator interface.
func (e *tileGenExpression) Evaluate(c *Chunk, p util.Point, cm *CityMap) {
	c.Tiles[p.Y*ChunkWidth+p.X] = e.r.Generate()
}

// itemExpression lays down a fixed item with a given chance.
type itemExpression struct {
	r    string // Item template name
	x, y int    // rng parameters
	n    int    // Number of repetitions
}

// Evaluate implements the evaluator interface.
func (e *itemExpression) Evaluate(c *Chunk, p util.Point, cm *CityMap) {
	for _, i := range e.CreateItems(cm.Now) {
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
func (e *itemGenExpression) Evaluate(c *Chunk, p util.Point, cm *CityMap) {
	for _, i := range e.CreateItems(cm.Now) {
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
func (e *actorExpression) Evaluate(c *Chunk, p util.Point, cm *CityMap) {
	if util.Random(0, e.y) < e.x {
		a := NewActor(e.r, cm.Now, true)
		a.Position = p
		c.PlaceActorRelative(a, cm)
	}
}

// actorGenExpression lays down an actor with a given chance based on a
// generator.
type actorGenExpression struct {
	r    ActorGen // Actor template name
	x, y int      // rng parameters
}

// Evaluate implements the evaluator interface.
func (e *actorGenExpression) Evaluate(c *Chunk, p util.Point, cm *CityMap) {
	if util.Random(0, e.y) >= e.x {
		return
	}
	a := e.r.Generate(cm.Now)
	a.Position = p
	c.PlaceActorRelative(a, cm)
}

// vehicleGenExpression lays down a vehicle with a given chance based on the
// named vehicle group.
type vehicleGenExpression struct {
	g    *VehicleGenGroup // Vehicle group to pull variants from
	f    util.Facing      // Output vehicle facing
	w, h int              // Vehicle spawn area dimensions
	x, y int              // rng parameters
}

// Evaluate implements the evaluator interface.
func (e *vehicleGenExpression) Evaluate(c *Chunk, p util.Point, cm *CityMap) {
	if util.Random(0, e.y) >= e.x {
		return
	}
	var gb util.Rect
	cb := util.NewRectWH(ChunkWidth, ChunkHeight)
	sb := cb.RotateRect(util.NewRectXYWH(p.X, p.Y, e.w, e.h), c.Facing)
	// Try up to 8 times to select a variant that will fit within the bounds.
	var gen *VehicleGen
	for i := 0; i < 8; i++ {
		gen = e.g.Get()
		gb = cb.RotateRect(util.NewRectXYWH(p.X, p.Y, gen.Width, gen.Height), c.Facing)
		if sb.ContainsRect(gb) {
			break
		}
		gen = nil
	}
	if gen == nil {
		return
	}
	// Randomly move the spawn location within the spawn bounds
	v := gen.Generate(cm.Now)
	v.Facing = e.f.Rotate(c.Facing)
	gb = sb.RandomSubRect(gb.Width(), gb.Height())
	v.Bounds = v.Bounds.Move(c.Bounds.TL.Add(gb.TL))
	c.Vehicles = append(c.Vehicles, v)
}

// GenStatement is a list of expressions to run on a single position in the
// chunk at generation time.
type GenStatement struct {
	Tile    Evaluator
	Vehicle Evaluator
	Actor   Evaluator
	Items   []Evaluator
}

func (s *GenStatement) UnmarshalJSON(in []byte) error {
	es, err := parseStatement(in)
	if err != nil {
		return err
	}
	for _, e := range es {
		switch e.(type) {
		case *tileExpression:
			if s.Tile != nil {
				return errors.New("multiple tile expressions given")
			}
			s.Tile = e
		case *tileGenExpression:
			if s.Tile != nil {
				return errors.New("multiple tile expressions given")
			}
			s.Tile = e
		case *actorExpression:
			if s.Actor != nil {
				return errors.New("multiple actor expressions given")
			}
			s.Actor = e
		case *actorGenExpression:
			if s.Actor != nil {
				return errors.New("multiple actor expressions given")
			}
			s.Actor = e
		case *vehicleGenExpression:
			if s.Vehicle != nil {
				return errors.New("multiple vehicle expressions given")
			}
			s.Vehicle = e
		default:
			s.Items = append(s.Items, e)
		}
	}
	if s.Tile == nil {
		return errors.New("no tile given in expression")
	}
	return nil
}

// parseVehicleExpression parses a vehicle expression string into its parts.
func parseVehicleExpression(s string) (f util.Facing, w, h int, err error) {
	parts := strings.Split(s, "x")
	if len(parts) != 2 || len(parts[0]) < 1 || len(parts[1]) < 1 {
		return util.FacingInvalid, 0, 0, errors.New("bad vehicle expression: " + s)
	}
	fs := parts[0][0]
	ws := parts[0][1:]
	hs := parts[1]
	switch fs {
	case 'N':
		f = util.FacingNorth
	case 'E':
		f = util.FacingEast
	case 'S':
		f = util.FacingSouth
	case 'W':
		f = util.FacingWest
	default:
		return util.FacingInvalid, 0, 0, errors.New("bad facing code in vehicle expression: " + s)
	}
	var v int64
	v, err = strconv.ParseInt(ws, 0, 32)
	if err != nil {
		return
	}
	w = int(v)
	v, err = strconv.ParseInt(hs, 0, 32)
	if err != nil {
		return
	}
	h = int(v)
	return
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
			if gnp := strings.Split(parts[0], "^"); len(gnp) == 2 {
				if group, found := VehicleGenGroups[gnp[0]]; found {
					if n > 1 {
						return nil, fmt.Errorf("'*' symbol not allowed in vehicle expressions")
					}
					f, w, h, err := parseVehicleExpression(gnp[1])
					if err != nil {
						return nil, err
					}
					ret = append(ret, &vehicleGenExpression{
						g: group,
						f: f,
						w: w,
						h: h,
						x: 1,
						y: 1,
					})
				} else {
					return nil, fmt.Errorf("vehicle group %s not found", gnp[0])
				}
			} else if gen, found := ActorGens[parts[0]]; found {
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
			if gnp := strings.Split(parts[0], "^"); len(gnp) == 2 {
				if group, found := VehicleGenGroups[gnp[0]]; found {
					if n > 1 {
						return nil, fmt.Errorf("'*' symbol not allowed in vehicle expressions")
					}
					f, w, h, err := parseVehicleExpression(gnp[1])
					if err != nil {
						return nil, err
					}
					ret = append(ret, &vehicleGenExpression{
						g: group,
						f: f,
						w: w,
						h: h,
						x: int(x),
						y: int(y),
					})
				} else {
					return nil, fmt.Errorf("vehicle group %s not found", gnp[0])
				}
			} else if gen, found := ActorGens[parts[0]]; found {
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
// items generated, if any, in a non-nil slice.
func (s ItemStatement) Evaluate(t time.Time) []*Item {
	ret := []*Item{}
	for _, exp := range s {
		items := exp.CreateItems(t)
		if len(items) > 0 {
			ret = append(ret, items...)
		}
	}
	return ret
}
