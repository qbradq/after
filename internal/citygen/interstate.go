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
	cgStreet := ChunkGenGroups["Street"].Get()
	cgRoad := ChunkGenGroups["Road"].Get()
	cgInterstateStreetIntersection := ChunkGenGroups["InterstateStreetIntersection"].Get()
	cgHighwayStreetIntersection := ChunkGenGroups["HighwayStreetIntersection"].Get()
	cgStreetStreetIntersection := ChunkGenGroups["StreetStreetIntersection"].Get()
	cgStreetRoadIntersection := ChunkGenGroups["StreetRoadIntersection"].Get()
	cgHighwayRoadIntersection := ChunkGenGroups["HighwayRoadIntersection"].Get()
	cgRoadRoadIntersection := ChunkGenGroups["RoadRoadIntersection"].Get()
	// layStreet lays a street from the starting point in the given facing
	// for the given distance.
	layStreet := func(sp util.Point, f util.Facing, l int) {
		for ; l > 0; l-- {
			// Try to place a straight street chunk
			if !place(m, cgStreet, sp, f, false) {
				// Try to place an intersection
				c := m.GetChunkFromMapPoint(sp)
				switch c.Generator.GetGroup() {
				case "Interstate":
					place(m, cgInterstateStreetIntersection, sp, c.Facing, true)
				case "Highway":
					if c.Facing == util.FacingWest {
						sp.Y++
						place(m, cgHighwayStreetIntersection, sp, c.Facing, true)
						sp.Y--
					} else {
						place(m, cgHighwayStreetIntersection, sp, c.Facing, true)
					}
				case "Street":
					place(m, cgStreetStreetIntersection, sp, c.Facing, true)
				case "Road":
					place(m, cgStreetRoadIntersection, sp, (c.Facing + 1).Bound(), true)
				}
			}
			sp = sp.StepFacing(f)
		}
	}
	// layRoad lays a road from the starting point in the given facing for the
	// given distance.
	layRoad := func(sp util.Point, f util.Facing, l int) {
		for ; l > 0; l-- {
			// Try to place a straight road chunk
			if !place(m, cgRoad, sp, f, false) {
				// Try to place an intersection
				c := m.GetChunkFromMapPoint(sp)
				switch c.Generator.GetGroup() {
				case "Highway":
					if c.Facing == util.FacingWest {
						sp.Y++
						place(m, cgHighwayRoadIntersection, sp, c.Facing, true)
						sp.Y--
					} else {
						place(m, cgHighwayStreetIntersection, sp, c.Facing, true)
					}
				case "Street":
					place(m, cgStreetRoadIntersection, sp, c.Facing, true)
				case "Road":
					place(m, cgRoadRoadIntersection, sp, c.Facing, true)
				}
			}
			sp = sp.StepFacing(f)
		}
	}
	// Lay down the base forest and clearing land pattern
	cgBrushyField := ChunkGenGroups["BrushyField"].Get()
	cgGrassyField := ChunkGenGroups["GrassyField"].Get()
	cgForest := ChunkGenGroups["Forest"].Get()
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
					set(m, p, cgBrushyField, f)
				} else {
					set(m, p, cgGrassyField, f)
				}
			} else {
				set(m, p, cgForest, f)
			}
		}
	}
	// Main interstate artery
	cgInterstateHighwayIntersection := ChunkGenGroups["InterstateHighwayIntersection"].Get()
	cgInterstate := ChunkGenGroups["Interstate"].Get()
	iip = util.NewPoint(m.Bounds.Width()/2, m.Bounds.Height()/2)
	p = iip
	place(m, cgInterstateHighwayIntersection, p, util.FacingNorth, false)
	for ; p.Y >= 0; p.Y-- {
		place(m, cgInterstate, p, util.FacingNorth, false)
	}
	p = iip
	p.Y += 3
	for ; p.Y < m.Bounds.Height(); p.Y++ {
		place(m, cgInterstate, p, util.FacingNorth, false)
	}
	// Crossing highway
	cgHighway := ChunkGenGroups["Highway"].Get()
	p = iip
	p.X--
	p.Y += 2
	for ; p.X >= 0; p.X-- {
		place(m, cgHighway, p, util.FacingWest, false)
	}
	p = iip
	p.X += cgInterstateHighwayIntersection.Width
	p.Y++
	for ; p.X < m.Bounds.Width(); p.X++ {
		place(m, cgHighway, p, util.FacingEast, false)
	}
	// Street and road network
	nsRoadsX := []int{}
	nsStreetsX := []int{}
	ewRoadsY := []int{}
	ewStreetsY := []int{}
	nStreets := 4
	minRoads := 2
	maxRoads := 4
	minBlockWidth := 4
	maxBlockWidth := 8
	// Western N/S streets
	p = iip
	p.X--
	for iStreet := 0; iStreet < nStreets; iStreet++ {
		nRoads := util.Random(minRoads, maxRoads+1)
		for iRoad := 0; iRoad < nRoads; iRoad++ {
			p.X -= util.Random(minBlockWidth, maxBlockWidth+1) + 1
			nsRoadsX = append(nsRoadsX, p.X)
		}
		if iStreet < nStreets-1 {
			p.X -= util.Random(minBlockWidth, maxBlockWidth+1) + 1
			nsStreetsX = append(nsStreetsX, p.X)
		}
	}
	// Eastern N/S streets
	p = iip
	p.X += cgInterstateHighwayIntersection.Width
	for iStreet := 0; iStreet < nStreets; iStreet++ {
		nRoads := util.Random(minRoads, maxRoads+1)
		for iRoad := 0; iRoad < nRoads; iRoad++ {
			p.X += util.Random(minBlockWidth, maxBlockWidth+1) + 1
			nsRoadsX = append(nsRoadsX, p.X)
		}
		if iStreet < nStreets-1 {
			p.X += util.Random(minBlockWidth, maxBlockWidth+1) + 1
			nsStreetsX = append(nsStreetsX, p.X)
		}
	}
	// Northern E/W streets
	p = iip
	for iStreet := 0; iStreet < nStreets; iStreet++ {
		nRoads := util.Random(minRoads, maxRoads+1)
		for iRoad := 0; iRoad < nRoads; iRoad++ {
			p.Y -= util.Random(minBlockWidth, maxBlockWidth+1) + 1
			ewRoadsY = append(ewRoadsY, p.Y)
		}
		if iStreet < nStreets-1 {
			p.Y -= util.Random(minBlockWidth, maxBlockWidth+1) + 1
			ewStreetsY = append(ewStreetsY, p.Y)
		}
	}
	// Southern E/W streets
	p = iip
	p.Y += cgInterstateHighwayIntersection.Height
	for iStreet := 0; iStreet < nStreets; iStreet++ {
		nRoads := util.Random(minRoads, maxRoads+1)
		for iRoad := 0; iRoad < nRoads; iRoad++ {
			p.Y += util.Random(minBlockWidth, maxBlockWidth+1) + 1
			ewRoadsY = append(ewRoadsY, p.Y)
		}
		if iStreet < nStreets-1 {
			p.Y += util.Random(minBlockWidth, maxBlockWidth+1) + 1
			ewStreetsY = append(ewStreetsY, p.Y)
		}
	}
	// Lay down streets and roads
	minRoadX := iip.X
	maxRoadX := iip.X
	minRoadY := iip.Y
	maxRoadY := iip.Y
	for _, x := range nsRoadsX {
		if x < minRoadX {
			minRoadX = x
		}
		if x > maxRoadX {
			maxRoadX = x
		}
	}
	for _, y := range ewRoadsY {
		if y < minRoadY {
			minRoadY = y
		}
		if y > maxRoadY {
			maxRoadY = y
		}
	}
	for _, x := range nsRoadsX {
		layRoad(util.NewPoint(x, minRoadY), util.FacingSouth, (maxRoadY-minRoadY)+1)
	}
	for _, x := range nsStreetsX {
		layStreet(util.NewPoint(x, minRoadY), util.FacingSouth, (maxRoadY-minRoadY)+1)
	}
	for _, y := range ewRoadsY {
		layRoad(util.NewPoint(minRoadX, y), util.FacingEast, (maxRoadX-minRoadX)+1)
	}
	for _, y := range ewStreetsY {
		layStreet(util.NewPoint(minRoadX, y), util.FacingEast, (maxRoadX-minRoadX)+1)
	}
	// Build possible locations
	type potentialLocation struct {
		p util.Point
		f util.Facing
	}
	phl := []potentialLocation{}
	pbl := []potentialLocation{}
	for _, x := range nsRoadsX {
		for y := minRoadY; y < maxRoadY; y++ {
			phl = append(phl,
				potentialLocation{p: util.NewPoint(x-1, y), f: util.FacingEast},
				potentialLocation{p: util.NewPoint(x+1, y), f: util.FacingWest},
			)
		}
	}
	for _, x := range nsStreetsX {
		for y := minRoadY; y < maxRoadY; y++ {
			pbl = append(pbl,
				potentialLocation{p: util.NewPoint(x-1, y), f: util.FacingEast},
				potentialLocation{p: util.NewPoint(x+1, y), f: util.FacingWest},
			)
		}
	}
	for _, y := range ewRoadsY {
		for x := minRoadX; x < maxRoadX; x++ {
			phl = append(phl,
				potentialLocation{p: util.NewPoint(x, y-1), f: util.FacingSouth},
				potentialLocation{p: util.NewPoint(x, y+1), f: util.FacingNorth},
			)
		}
	}
	for _, y := range ewStreetsY {
		for x := minRoadX; x < maxRoadX; x++ {
			pbl = append(pbl,
				potentialLocation{p: util.NewPoint(x, y-1), f: util.FacingSouth},
				potentialLocation{p: util.NewPoint(x, y+1), f: util.FacingNorth},
			)
		}
	}
	for i := range phl {
		j := util.Random(0, i+1)
		phl[i], phl[j] = phl[j], phl[i]
	}
	for i := range pbl {
		j := util.Random(0, i+1)
		pbl[i], pbl[j] = pbl[j], pbl[i]
	}
	sl := float64(iip.Distance(util.NewPoint(maxRoadX, maxRoadY)))
	// Try to place businesses
	for _, l := range pbl {
		r := (1.0 - float64(l.p.Distance(iip))/sl) / 2
		if util.RandomF(0, 1.0) < r {
			place(m, ChunkGenGroups["Business"].Get(), l.p, l.f, false)
		}
	}
	// Fill with houses
	for _, l := range phl {
		r := 1.0 - float64(l.p.Distance(iip))/sl
		if util.RandomF(0, 1.0) < r {
			place(m, ChunkGenGroups["House"].Get(), l.p, l.f, false)
		}
	}
	// Test chunk
	p = iip
	p.X--
	place(m, ChunkGenGroups["Test"].Get(), p, util.FacingSouth, true)
	return m
}
