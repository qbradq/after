package game

import (
	"math"

	"github.com/qbradq/after/lib/util"
)

// Point offsets for neighbor checks
var dMapPointOfs = []struct {
	P util.Point
	D util.Direction
}{
	{P: util.Point{X: 0, Y: -1}, D: util.DirectionNorth},
	{P: util.Point{X: 1, Y: -1}, D: util.DirectionNorthEast},
	{P: util.Point{X: 1, Y: 0}, D: util.DirectionEast},
	{P: util.Point{X: 1, Y: 1}, D: util.DirectionSouthEast},
	{P: util.Point{X: 0, Y: 1}, D: util.DirectionSouth},
	{P: util.Point{X: -1, Y: 1}, D: util.DirectionSouthWest},
	{P: util.Point{X: -1, Y: 0}, D: util.DirectionWest},
	{P: util.Point{X: -1, Y: -1}, D: util.DirectionNorthWest},
}

// DMapRank represents the rank of a position in the Dijkstra map.
type DMapRank uint16

const (
	DMapRankMin DMapRank = 0
	DMapRankMax DMapRank = math.MaxUint16 - 1
)

// DMap implements a "Dijkstra Map" as described by Brian Walker - the developer
// of Brouge - in this article: https://www.roguebasin.com/index.php/The_Incredible_Power_of_Dijkstra_Maps
type DMap struct {
	Bounds util.Rect  // Bounds of the map relative to the city map
	Map    []DMapRank // Value map
}

// NewDMap constructs a new DMap struct with the given bounds.
func NewDMap(b util.Rect) *DMap {
	return &DMap{
		Bounds: b,
		Map:    make([]DMapRank, b.Width()*b.Height()),
	}
}

// SetBounds changes the bounds of the map. Reallocation is done only if the
// area of the new bounds exceeds the area of the old.
func (m *DMap) SetBounds(b util.Rect) {
	os := len(m.Map)
	ns := b.Area()
	if os == ns {
		m.Bounds = b
		return
	}
	if ns < os {
		m.Map = m.Map[:ns]
		return
	}
	m.Map = make([]DMapRank, ns)
}

// Relocate changes the position of the top-left corner of the bounds without
// changing the area of the bounds.
func (m *DMap) Relocate(p util.Point) {
	m.Bounds = util.NewRectXYWH(p.X, p.Y, m.Bounds.Width(), m.Bounds.Height())
}

// Clear sets the rank of all points on the map to DMapRankMax.
func (m *DMap) Clear() {
	for i := range m.Map {
		m.Map[i] = DMapRankMax
	}
}

// SetGoal sets the rank of the city-map relative point p to DMapRankMin.
func (m *DMap) SetGoal(p util.Point) {
	m.Map[(p.Y-m.Bounds.TL.Y)*m.Bounds.Width()+(p.X-m.Bounds.TL.X)] = DMapRankMin
}

// Calculate calculates the Dijkstra map based on the current set of goals. In
// order for Calculate to have meaningful output, call Clear() first and
// SetGoal() one or more times beforehand.
func (m *DMap) Calculate(cm *CityMap) {
	var p util.Point
	changed := true
	for changed {
		changed = false
		for p.Y = m.Bounds.TL.Y; p.Y <= m.Bounds.BR.Y; p.Y++ {
			for p.X = m.Bounds.TL.X; p.X <= m.Bounds.BR.X; p.X++ {
				c := cm.GetChunk(p)
				if c.bitmapsDirty {
					c.RebuildBitmaps(cm)
				}
				blocks := c.BlocksWalk.Contains(c.relOfs(p))
				if blocks {
					continue
				}
				lr := DMapRankMax
				for i := 0; i < len(dMapPointOfs); i += 2 {
					ofs := dMapPointOfs[i]
					tp := p.Add(ofs.P)
					if !m.Bounds.Contains(tp) {
						continue
					}
					i := (tp.Y-m.Bounds.TL.Y)*m.Bounds.Width() + (tp.X - m.Bounds.TL.X)
					r := m.Map[i]
					if r < lr {
						lr = r
					}
				}
				idx := (p.Y-m.Bounds.TL.Y)*m.Bounds.Width() + (p.X - m.Bounds.TL.X)
				r := m.Map[idx]
				if r > lr+1 {
					m.Map[idx] = lr + 1
					changed = true
				}
			}
		}
	}
}

// RollDown returns a randomly selected rank from the set of lowest neighbor
// ranks. If p is out of bounds p is returned and the direction returned will
// be util.DirectionInvalid.
func (m *DMap) RollDown(p util.Point) (util.Point, util.Direction) {
	if !m.Bounds.Contains(p) {
		return p, util.DirectionInvalid
	}
	idx := (p.Y-m.Bounds.TL.Y)*m.Bounds.Width() + (p.X - m.Bounds.TL.X)
	r := m.Map[idx]
	lr := r
	si := util.Random(0, 8)
	lni := -1
	for i := si; i < si+8; i++ {
		ofs := dMapPointOfs[i%8]
		tp := p.Add(ofs.P)
		if !m.Bounds.Contains(tp) {
			continue
		}
		nr := m.Map[(tp.Y-m.Bounds.TL.Y)*m.Bounds.Width()+(tp.X-m.Bounds.TL.X)]
		if nr < lr {
			lni = i % 8
			lr = nr
		}
	}
	if lni < 0 {
		return p, util.DirectionInvalid
	}
	ofs := dMapPointOfs[lni]
	return p.Add(ofs.P), ofs.D
}

// Rank returns the rank of the point, or DMapRankMax if out of bounds.
func (m *DMap) Rank(p util.Point) DMapRank {
	if !m.Bounds.Contains(p) {
		return DMapRankMax
	}
	return m.Map[(p.Y-m.Bounds.TL.Y)*m.Bounds.Width()+(p.X-m.Bounds.TL.X)]
}
