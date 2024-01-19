package util

import (
	"math/rand"
	"time"
)

// rng is the global random number generator implementation and is provided by
// the runtime.
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// Random returns a random int within the half-open range [min-max).
func Random(min, max int) int {
	if min < 0 || max < 0 {
		return 0
	}
	return rng.Intn(max-min) + min
}

// RandomF returns a random floating point value in the half-open range
// [min-max).
func RandomF(min, max float64) float64 {
	return rng.Float64()*(max-min) + min
}

// RandomBool returns a random boolean value.
func RandomBool() bool {
	return rng.Intn(2) != 0
}

// RandomValue returns a random value from the slice.
func RandomValue[T any](s []T) T {
	return s[Random(0, len(s))]
}

// RandomPoint returns a random point within the rect.
func RandomPoint(b Rect) Point {
	return Point{
		X: Random(b.TL.X, b.BR.X+1),
		Y: Random(b.TL.Y, b.BR.Y+1),
	}
}
