package util

import "math"

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

// Sub returns the result of subtracting a from p.
func (p Point) Sub(a Point) Point {
	return Point{p.X - a.X, p.Y - a.Y}
}

// Multiply returns the result of multiplying a and p.
func (p Point) Multiply(a int) Point {
	return Point{p.X * a, p.Y * a}
}

// Divide returns the result of dividing p by a.
func (p Point) Divide(a int) Point {
	return Point{p.X / a, p.Y / a}
}

// DirectionTo returns the direction code that most closely matches the
// direction of the argument point.
func (p Point) DirectionTo(a Point) Direction {
	r := math.Atan2(float64(a.X-p.X), float64(a.Y-p.Y)) * 180 / math.Pi
	b := -157.5
	if r < b+45*0 {
		return DirectionNorth
	}
	if r < b+45*1 {
		return DirectionNorthWest
	}
	if r < b+45*2 {
		return DirectionWest
	}
	if r < b+45*3 {
		return DirectionSouthWest
	}
	if r < b+45*4 {
		return DirectionSouth
	}
	if r < b+45*5 {
		return DirectionSouthEast
	}
	if r < b+45*6 {
		return DirectionEast
	}
	if r < b+45*7 {
		return DirectionNorthEast
	}
	return DirectionNorth
}

// Distance returns the maximum distance from p to d along either the X or Y
// axis.
func (p Point) Distance(d Point) int {
	dx := p.X - d.X
	dy := p.Y - d.Y
	if dx < 0 {
		dx = dx * -1
	}
	if dy < 0 {
		dy = dy * -1
	}
	if dx > dy {
		return dx
	}
	return dy
}
