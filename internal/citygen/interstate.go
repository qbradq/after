package citygen

import (
	"github.com/larspensjo/Go-simplex-noise/simplexnoise"
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/util"
)

func init() {
	reg("Interstate Town", interstate)
}

// interstate implements a CityGen that creates a town centered on an
// interstate highway intersection with a state highway.
func interstate() *game.CityMap {
	m := game.NewCityMap()
	var p, iip util.Point
	var l int
	// layStreet lays a street from the starting point in the given facing
	// for the given distance.
	layStreet := func(sp util.Point, f util.Facing, l int) {
		for ; l > 0; l-- {
			// Try to place a straight street chunk
			if !place(m, ChunkGens["Street"], sp, f, false) {
				// Try to place an intersection
				c := m.GetChunkFromMapPoint(sp)
				switch c.Generator.GetID() {
				case "Interstate":
					place(m, ChunkGens["InterstateStreetIntersection"], sp, c.Facing, true)
				case "Highway":
					sp.Y++
					place(m, ChunkGens["HighwayIntersection"], sp, c.Facing, true)
					sp.Y--
				case "Street":
					place(m, ChunkGens["StreetStreetIntersection"], sp, c.Facing, true)
				}
			}
			sp = sp.StepFacing(f)
		}
	}
	// Lay down the base forest and clearing land pattern
	nox := util.RandomF(0, 1024)
	noy := util.RandomF(0, 1024)
	for p.Y = 0; p.Y < m.Bounds.Height(); p.Y++ {
		for p.X = 0; p.X < m.Bounds.Width(); p.X++ {
			n := simplexnoise.Noise2(
				float64(p.X)/32+nox,
				float64(p.Y)/32+noy,
			)
			f := util.Facing(util.Random(0, 4))
			if n > 0.25 {
				if util.Random(0, 16) == 0 {
					set(m, p, ChunkGens["BrushyField"], f)
				} else {
					set(m, p, ChunkGens["GrassyField"], f)
				}
			} else {
				set(m, p, ChunkGens["Forest"], f)
			}
		}
	}
	// Main interstate artery
	iip = util.NewPoint(m.Bounds.Width()/2, m.Bounds.Height()/2)
	p = iip
	place(m, ChunkGens["InterstateHighwayIntersection"], p, util.FacingNorth, false)
	for ; p.Y >= 0; p.Y-- {
		place(m, ChunkGens["Interstate"], p, util.FacingNorth, false)
	}
	p = iip
	p.Y += 3
	for ; p.Y < m.Bounds.Height(); p.Y++ {
		place(m, ChunkGens["Interstate"], p, util.FacingNorth, false)
	}
	// Crossing highway
	p = iip
	p.X--
	p.Y += 2
	for ; p.X >= 0; p.X-- {
		place(m, ChunkGens["Highway"], p, util.FacingWest, false)
	}
	p = iip
	p.X += ChunkGens["InterstateHighwayIntersection"].Width
	p.Y++
	for ; p.X < m.Bounds.Width(); p.X++ {
		place(m, ChunkGens["Highway"], p, util.FacingEast, false)
	}
	// Western N/S streets
	p = iip
	p.X--
	l = 256
	for i := 0; i < 4; i++ {
		p.X -= 32 + util.Random(0, 32)
		p.Y = iip.Y - l/2
		layStreet(p, util.FacingSouth, l)
		l -= 16
	}
	// Eastern N/S streets
	p = iip
	p.X += ChunkGens["InterstateHighwayIntersection"].Width
	l = 256
	for i := 0; i < 4; i++ {
		p.X += 32 + util.Random(0, 32)
		p.Y = iip.Y - l/2
		layStreet(p, util.FacingSouth, l)
		l -= 16
	}
	// Northern E/W streets
	p = iip
	p.X--
	l = 256
	for i := 0; i < 4; i++ {
		p.Y -= 32 + util.Random(0, 32)
		p.X = iip.X - l/2
		layStreet(p, util.FacingEast, l)
		l -= 16
	}
	// Southern E/W streets
	p = iip
	p.X += ChunkGens["InterstateHighwayIntersection"].Width
	l = 256
	for i := 0; i < 4; i++ {
		p.Y += 32 + util.Random(0, 32)
		p.X = iip.X - l/2
		layStreet(p, util.FacingEast, l)
		l -= 16
	}
	// // Test chunk
	// p = iip
	// p.X--
	// place(m, ChunkGens["Test"], p, util.FacingSouth)
	// // Test house
	// p = iip
	// p.X -= 2
	// place(m, ChunkGens["House"], p, util.FacingSouth)
	return m
}
