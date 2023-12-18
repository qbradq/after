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
