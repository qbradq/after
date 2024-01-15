package util

import "math"

// Return slice for Ray
var rayBuf []Point

// Ray returns a slice of points along the ray, including both the starting
// and ending points. Subsequent calls to Ray reuse the same return slice.
func Ray(p1, p2 Point) []Point {
	rayBuf = rayBuf[:0]
	dx := int(math.Abs(float64(p2.X - p1.X)))
	dy := int(math.Abs(float64(p2.Y - p1.Y)))
	sx, sy := 1, 1
	if p2.X < p1.X {
		sx = -1
	}
	if p2.Y < p1.Y {
		sy = -1
	}
	err := dx - dy
	gx, gy := p1.X, p1.Y
	rayBuf = append(rayBuf, Point{
		X: gx,
		Y: gy,
	})
	for {
		err2 := 2 * err
		if err2 > (dy * -1) {
			err -= dy
			gx += sx
		}
		if err2 < dx {
			err += dx
			gy += sy
		}
		rayBuf = append(rayBuf, Point{
			X: gx,
			Y: gy,
		})
		if gx == p2.X && gy == p2.Y {
			break
		}
	}
	return rayBuf
}
