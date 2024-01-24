package game

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/qbradq/after/lib/util"
)

// ActorGen generates a single actor from a set of possibilities.
type ActorGen []string

// ActorGens is the mapping of generator names to objects.
var ActorGens = map[string]ActorGen{}

func (g *ActorGen) UnmarshalJSON(in []byte) error {
	var src = map[string]int{}
	json.Unmarshal(in, &src)
	for k, n := range src {
		_, found := ActorDefs[k]
		if !found {
			panic(fmt.Errorf("ActorGen referenced non-existent actor %s", k))
		}
		for i := 0; i < n; i++ {
			*g = append(*g, k)
		}
	}
	return nil
}

// Generate returns a pointer to the selected tile def after procedural
// generation.
func (g ActorGen) Generate(t time.Time) *Actor {
	r := g[util.Random(0, len(g))]
	return NewActor(r, t)
}
