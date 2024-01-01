package util

// Flood fill offsets
var ffOfs = []Point{
	{0, -1},
	{1, 0},
	{0, 1},
	{-1, 0},
}

// FloodFill implements a flood-fill algorithm.
type FloodFill struct {
	Matches func(Point) bool // Returns true if the point matches the source
	Set     func(Point)      // Sets the point
}

// Execute executes the flood fill algorithm. Matches and Set must be non-nil.
func (f *FloodFill) Execute(p Point) {
	if f.Matches == nil || f.Set == nil {
		return
	}
	f.fill(p)
}

// fill tests a single point and if it matches, fills that point and tests the
// neighbors.
func (f *FloodFill) fill(p Point) {
	if !f.Matches(p) {
		return
	}
	f.Set(p)
	f.fill(p.Add(ffOfs[0]))
	f.fill(p.Add(ffOfs[1]))
	f.fill(p.Add(ffOfs[2]))
	f.fill(p.Add(ffOfs[3]))
}
