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
	Name              string          // Descriptive name of the scenario
	Description       string          // Full descriptive text for the scenario
	StartingChunkType string          // Type of chunk to select for the starting location
	Equipment         []itemStatement // Item and item generator expressions to equip to the player on spawn
	Inventory         []itemStatement // Item and item generator expressions to add to the player's inventory on spawn
	Weapon            itemStatement   // Item or item generator expression of the item to wield as a weapon
	SafeZoneRadius    int             // Radius of the "safe zone" surrounding the starting chunk which has all actors removed at spawn
}

// Execute sets up the city map and player according to the parameters of the
// scenario.
func (s *Scenario) Execute(m *game.CityMap) {
	m.Player = game.NewPlayer(m.Now)
	// Starting equipment
	for _, statement := range s.Equipment {
		i := statement.evaluate(m.Now)
		if i == nil {
			continue
		}
		r := m.Player.WearItem(i)
		if r != "" {
			panic(errors.New(r))
		}
	}
	// Starting inventory
	for _, statement := range s.Inventory {
		i := statement.evaluate(m.Now)
		if i == nil {
			continue
		}
		m.Player.AddItemToInventory(i)
	}
	// Starting weapon
	if len(s.Weapon) != 0 {
		i := s.Weapon.evaluate(m.Now)
		if i != nil {
			r := m.Player.WieldItem(i)
			if r != "" {
				panic(errors.New(r))
			}
		}
	}
	// Scan the map for suitable starting locations and pick one at random
	var cs []*game.Chunk
	for _, c := range m.Chunks {
		if c.Generator.GetID() == s.StartingChunkType {
			cs = append(cs, c)
		}
	}
	c := util.RandomValue[*game.Chunk](cs)
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
