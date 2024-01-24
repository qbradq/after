package game

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/qbradq/after/lib/util"
)

// ItemGen generates a single item from a set of possibilities.
type ItemGen []string

// ItemGens is the mapping of generator names to objects.
var ItemGens = map[string]ItemGen{}

func (g *ItemGen) UnmarshalJSON(in []byte) error {
	var src = map[string]int{}
	json.Unmarshal(in, &src)
	for k, n := range src {
		_, found := ItemDefs[k]
		if !found {
			panic(fmt.Errorf("ItemGen referenced non-existent item %s", k))
		}
		for i := 0; i < n; i++ {
			*g = append(*g, k)
		}
	}
	return nil
}

// Generate returns a new item created from the generator.
// generation.
func (g ItemGen) Generate(now time.Time) *Item {
	r := g[util.Random(0, len(g))]
	return NewItem(r, now, true)
}
