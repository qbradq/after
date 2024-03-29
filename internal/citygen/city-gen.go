// Package citygen implements city generators. See [game.CityMap].
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

// Generate generates a new CityMap for use with the named city generator and
// scenario.
func Generate(cityGen, scenario string) *game.CityMap {
	m := CityGens[cityGen]()
	Scenarios[scenario].Execute(m)
	return m
}
