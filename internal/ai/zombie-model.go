package ai

import (
	"time"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
)

// Zombie configures an AIModel to act as a zombie with very slow reaction
// time to sight but instant reaction time to sound.
func init() {
	reg("Zombie", func() *AIModel {
		return &AIModel{
			act: "zmActIdle",
		}
	})
	regFn("zmActIdle", func(ai *AIModel, a *game.Actor, t time.Time, m *game.CityMap) time.Duration {
		if !m.CanSeePlayerFrom(a.Position) {
			return time.Second * 10 // Up to 10 second delay for visual reactions
		}
		// Begin approaching the player
		ai.poi = m.Player.Position
		ai.act = "zmActApproach"
		return ai.Act(a, t, m) // Begin approaching immediately
	})
	regFn("zmActApproach", func(ai *AIModel, a *game.Actor, t time.Time, m *game.CityMap) time.Duration {
		// Close enough to attack, do that
		if a.Position.Distance(m.Player.Position) < 2 {
			return aiFns["zmActAttack"](ai, a, t, m)
		}
		// Need to get closer, try to approach
		_, d := m.PlayerApproachDMap.RollDown(a.Position)
		m.StepActor(a, d)
		return time.Second * time.Duration(a.WalkSpeed)
	})
	regFn("zmActAttack", func(ai *AIModel, a *game.Actor, t time.Time, m *game.CityMap) time.Duration {
		game.Log.Log(termui.ColorRed, "You got attacked!")
		return time.Second
	})
}
