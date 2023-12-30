package itemgen

import (
	"encoding/json"
	"fmt"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/util"
)

// ItemGen generates a single tile from a set of possibilities.
type ItemGen []string

// ItemGens is the mapping of generator names to objects.
var ItemGens = map[string]ItemGen{}

func (g *ItemGen) UnmarshalJSON(in []byte) error {
	var src = map[string]int{}
	json.Unmarshal(in, &src)
	for k, n := range src {
		_, found := game.ItemDefs[k]
		if !found {
			panic(fmt.Errorf("TileGen referenced non-existent tile %s", k))
		}
		for i := 0; i < n; i++ {
			*g = append(*g, k)
		}
	}
	return nil
}

// Generate returns a pointer to the selected tile def after procedural
// generation.
func (g ItemGen) Generate() *game.Item {
	r := g[util.Random(0, len(g))]
	return game.NewItem(r)
}
