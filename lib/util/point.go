package util

// Point represents an integer point in 2D space.
type Point struct {
	X int // X component
	Y int // Y component
}

// NewPoint returns a new Point value.
func NewPoint(x, y int) Point {
	return Point{
		X: x,
		Y: y,
	}
}

// Add returns the result of adding the two points X and Y values.
func (p Point) Add(a Point) Point {
	return Point{p.X + a.X, p.Y + a.Y}
}
