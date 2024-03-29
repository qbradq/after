package game

import (
	"encoding/json"
	"fmt"

	"github.com/qbradq/after/lib/util"
)

// TileGen generates a single tile from a set of possibilities.
type TileGen []TileRef

// TileGens is the mapping of generator names to objects.
var TileGens = map[string]TileGen{}

func (g *TileGen) UnmarshalJSON(in []byte) error {
	var src = map[string]int{}
	json.Unmarshal(in, &src)
	for k, n := range src {
		r, found := TileRefs[k]
		if !found {
			panic(fmt.Errorf("TileGen referenced non-existent tile %s", k))
		}
		for i := 0; i < n; i++ {
			*g = append(*g, r)
		}
	}
	return nil
}

// Generate returns a pointer to the selected tile def after procedural
// generation.
func (g TileGen) Generate() *TileDef {
	r := g[util.Random(0, len(g))]
	return TileDefs[r]
}
