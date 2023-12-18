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

// RandomBool returns a random boolean value.
func RandomBool() bool {
	return rng.Intn(2) != 0
}
