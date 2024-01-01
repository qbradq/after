package citygen

import (
	"fmt"

	"github.com/qbradq/after/internal/game"
)

// CityGen is a factory function for *game.CityMap that handles procedural
// generation.
type CityGen func() *game.CityMap

// Map of all city generators
var CityGens = map[string]CityGen{}

// reg registers the given CityGen under the given descriptive name.
func reg(name string, g CityGen) {
	if _, found := CityGens[name]; found {
		panic(fmt.Errorf("duplicate CityGen \"%s\"", name))
	}
	CityGens[name] = g
}
