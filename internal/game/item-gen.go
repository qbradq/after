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
		if len(k) < 1 {
			continue
		}
		for i := 0; i < n; i++ {
			*g = append(*g, k)
		}
	}
	return nil
}

// Validate validates the item generator, making sure all references will
// resolve at runtime.
func (g *ItemGen) Validate() error {
	for _, s := range *g {
		if len(s) < 1 {
			continue
		}
		if s[0] == '$' {
			if _, found := ItemGens[s[1:]]; !found {
				return fmt.Errorf("ItemGen referenced non-existent item generator %s", s[1:])
			}
		} else {
			if _, found := ItemDefs[s]; !found {
				return fmt.Errorf("ItemGen referenced non-existent item %s", s)
			}
		}
	}
	return nil
}

// Generate returns a new item created from the generator.
// generation.
func (g ItemGen) Generate(now time.Time) *Item {
	r := g[util.Random(0, len(g))]
	if len(r) < 1 {
		return nil
	}
	if r[0] == '$' {
		gen, found := ItemGens[r[1:]]
		if !found {
			return nil
		}
		return gen.Generate(now)
	}
	return NewItem(r, now, true)
}
