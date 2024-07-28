package citygen

import (
	"errors"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/util"
)

// Scenarios is the list of all scenarios by name.
var Scenarios = map[string]*Scenario{}

// Scenario controls the generation of the player, their equipment, inventory
// and placement within the city.
type Scenario struct {
	Name              string   // Descriptive name of the scenario
	Description       string   // Full descriptive text for the scenario
	StartingChunkType string   // Type of chunk to select for the starting location
	Equipment         []string // Item statements to equip to the player on spawn
	SafeZoneRadius    int      // Radius of the "safe zone" surrounding the starting chunk which has all actors removed at spawn
}

// Execute sets up the city map and player according to the parameters of the
// scenario.
func (s *Scenario) Execute(m *game.CityMap) {
	// Equipment injection
	game.ActorDefs["Player"].Equipment = s.Equipment
	game.ActorDefs["Player"].CacheEquipmentStatements()
	m.Player = game.NewPlayer(m.Now)
	// Scan the map for suitable starting locations and pick one at random
	cs := []*game.Chunk{}
	for _, c := range m.Chunks {
		if c.Generator.GetGroup() == s.StartingChunkType {
			cs = append(cs, c)
		}
	}
	c := util.RandomValue[*game.Chunk](cs)
	if c == nil {
		// No suitable chunk found, start in the dead center of the city
		c = m.GetChunk(util.NewPoint(
			game.ChunkWidth*game.CityMapWidth/2+game.ChunkWidth/2,
			game.ChunkHeight*game.CityMapHeight/2+game.ChunkHeight/2))
	}
	// Safe zone implementation
	if s.SafeZoneRadius > 0 {
		// Load all the chunks we need to modify
		r := util.NewRectFromRadius(c.Position, s.SafeZoneRadius)
		m.EnsureLoaded(r)
		// Clean out all actors from the safe zone
		for _, a := range m.ActorsWithin(r.Multiply(game.ChunkWidth)) {
			m.RemoveActor(a)
		}
	}
	// Load the chunk and scan for a valid location for the player
	m.LoadChunk(c, m.Now)
	for i := 0; i < 512; i++ {
		p := util.RandomPoint(c.Bounds)
		ws, cs := c.CanStep(&m.Player.Actor, p)
		if ws || cs {
			m.Player.Position = p
			return
		}
	}
	panic(errors.New("exhausted player placement attempts"))
}
