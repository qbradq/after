package citygen

import (
	"github.com/larspensjo/Go-simplex-noise/simplexnoise"
	"github.com/qbradq/after/internal/chunkgen"
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
	var p util.Point
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
					set(m, p, chunkgen.ChunkGens["BrushyField"], f)
				} else {
					set(m, p, chunkgen.ChunkGens["GrassyField"], f)
				}
			} else {
				set(m, p, chunkgen.ChunkGens["Forest"], f)
			}
		}
	}
	// Main interstate artery
	iip := util.NewPoint(10, 15)
	p = iip
	place(m, chunkgen.ChunkGens["InterstateHighwayIntersection"], p, util.FacingNorth)
	for ; p.Y >= 0; p.Y-- {
		place(m, chunkgen.ChunkGens["Interstate"], p, util.FacingNorth)
	}
	p = iip
	p.Y += 3
	for ; p.Y < m.Bounds.Height(); p.Y++ {
		place(m, chunkgen.ChunkGens["Interstate"], p, util.FacingNorth)
	}
	// Crossing highway
	p = iip
	p.X--
	p.Y += 2
	for ; p.X >= 0; p.X-- {
		place(m, chunkgen.ChunkGens["Highway"], p, util.FacingWest)
	}
	p = iip
	p.X += chunkgen.ChunkGens["InterstateHighwayIntersection"].Width
	p.Y++
	for ; p.X < m.Bounds.Width(); p.X++ {
		place(m, chunkgen.ChunkGens["Highway"], p, util.FacingEast)
	}
	// Test house
	p = iip
	p.X -= 2
	place(m, chunkgen.ChunkGens["House"], p, util.FacingSouth)
	return m
}
