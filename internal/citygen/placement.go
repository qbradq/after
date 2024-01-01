package citygen

import (
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/util"
)

// overlaps returns true if the given rect overlaps any chunks within the city
// map that have already been explicitly placed.
func overlaps(b util.Rect, m *game.CityMap) bool {
	var p util.Point
	for p.Y = b.TL.Y; p.Y <= b.BR.Y; p.Y++ {
		for p.X = b.TL.X; p.X <= b.BR.X; p.X++ {
			c := m.GetChunkFromMapPoint(p)
			if c == nil {
				// Out of bounds
				return true
			}
			if c.Flags&game.ChunkFlagsOccupied != 0 {
				return true
			}
		}
	}
	return false
}

// place attempts to place the given chunk generator with the given facing. This
// function returns true if the chunk generator was able to be placed properly.
func place(m *game.CityMap, g *ChunkGen, p util.Point, f util.Facing) bool {
	// Variable setup based on facing
	var b util.Rect
	switch f.Bound() {
	case util.FacingNorth:
		b = util.NewRectXYWH(p.X, p.Y, g.Width, g.Height)
	case util.FacingEast:
		b = util.NewRectXYWH(p.X-(g.Height-1), p.Y, g.Height, g.Width)
	case util.FacingSouth:
		b = util.NewRectXYWH(p.X-(g.Width-1), p.Y-(g.Height-1), g.Width, g.Height)
	case util.FacingWest:
		b = util.NewRectXYWH(p.X, p.Y-(g.Width-1), g.Height, g.Width)
	}
	// Bounds check
	if overlaps(b, m) {
		return false
	}
	// Function to translate current chunk rendering position to chunk offset
	fn := func(p util.Point) util.Point {
		switch f {
		case util.FacingNorth:
			return p
		case util.FacingEast:
			return util.Point{
				X: p.Y,
				Y: (g.Height - 1) - p.X,
			}
		case util.FacingSouth:
			return util.Point{
				X: (g.Width - 1) - p.X,
				Y: (g.Height - 1) - p.Y,
			}
		default:
			return util.Point{
				X: (g.Width - 1) - p.Y,
				Y: p.X,
			}
		}
	}
	// Chunk writes
	var cp util.Point
	for cp.Y = b.TL.Y; cp.Y <= b.BR.Y; cp.Y++ {
		for cp.X = b.TL.X; cp.X <= b.BR.X; cp.X++ {
			c := m.GetChunkFromMapPoint(cp)
			cgo := fn(util.Point{
				X: cp.X - b.TL.X,
				Y: cp.Y - b.TL.Y,
			})
			c.ChunkGenOffset = cgo
			c.Facing = f
			c.Generator = g
			g.AssignStaticInfo(c)
			c.Flags |= game.ChunkFlagsOccupied
		}
	}
	return true
}

// set sets the given chunk with parameters from the generator and assumes that
// the generator is a 1x1. This does *not* mark the chunk as occupied.
func set(m *game.CityMap, p util.Point, g *ChunkGen, f util.Facing) {
	c := m.GetChunkFromMapPoint(p)
	if c == nil {
		return
	}
	c.ChunkGenOffset = util.Point{}
	c.Facing = f
	c.Generator = g
	g.AssignStaticInfo(c)
	c.Flags &= ^game.ChunkFlagsOccupied
}
