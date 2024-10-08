package ai

import (
	"time"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/util"
)

// Zombie configures an AIModel to act as a zombie with very slow reaction
// time to sight but instant reaction time to sound.
func init() {
	reg("Zombie", func() *AIModel {
		return &AIModel{
			act:      "zmActIdle",
			periodic: "nil",
		}
	})
	regActFn("zmActIdle", func(ai *AIModel, a *game.Actor, m *game.CityMap) time.Duration {
		// Wait for the player to step into view
		if !ai.targetPlayer(a, m) {
			return time.Duration(float64(time.Second) * a.ActSpeed())
		}
		// Begin approaching the player
		ai.cd = time.Minute
		ai.act = "zmActApproach"
		return ai.Act(a, m) // Begin approaching immediately
	})
	regActFn("zmActApproach", func(ai *AIModel, a *game.Actor, m *game.CityMap) time.Duration {
		// Close enough to attack, do that
		if a.Position.Distance(m.Player.Position) < 2 {
			return actFns["zmActAttack"](ai, a, m)
		}
		// Try to re-target the player every step
		ai.targetPlayer(a, m)
		// No path to the POI found, just try to randomly advance towards it
		if len(ai.Path) == 0 && a.Position.Distance(ai.POI) > 1 {
			d := a.Position.DirectionTo(ai.POI)
			if ws, cs := m.StepActor(a, true, d); ws || cs {
				return time.Duration(float64(time.Second) * a.WalkSpeed())
			}
			o1s := 1
			o2s := -1
			if util.RandomBool() {
				o1s *= -1
				o2s *= -1
			}
			d = (d + util.Direction((o1s * 1))).Bound()
			if ws, cs := m.StepActor(a, true, d); ws || cs {
				return time.Duration(float64(time.Second) * a.WalkSpeed())
			}
			d = (d + util.Direction((o2s * 2))).Bound()
			if ws, cs := m.StepActor(a, true, d); ws || cs {
				return time.Duration(float64(time.Second) * a.WalkSpeed())
			}
			d = (d + util.Direction((o1s * 3))).Bound()
			if ws, cs := m.StepActor(a, true, d); ws || cs {
				return time.Duration(float64(time.Second) * a.WalkSpeed())
			}
			d = (d + util.Direction((o2s * 4))).Bound()
			if ws, cs := m.StepActor(a, true, d); ws || cs {
				return time.Duration(float64(time.Second) * a.WalkSpeed())
			}
			// Couldn't step in any direction even close to toward the POI, just
			// stand there looking dumb
			return time.Duration(float64(time.Second) * a.ActSpeed())
		}
		// Already at the POI or out of path steps, just wait there
		if len(ai.Path) == 0 || a.Position.Distance(ai.POI) < 1 {
			ai.cd -= time.Second
			if ai.cd <= 0 {
				ai.cd = 0
				ai.act = "zmActIdle"
			}
			return time.Duration(float64(time.Second) * a.ActSpeed())
		}
		// Need to get closer, try to approach
		if ws, cs := m.StepActor(a, true, ai.Path[0]); ws || cs {
			// Step was successful, advance the path
			ai.Path = ai.Path[1:]
			return time.Duration(float64(time.Second) * a.WalkSpeed())
		}
		// Our path is blocked, try to path around it
		ai.Path = ai.Path[:0]
		game.NewPath(a.Position, ai.POI, m, &ai.Path)
		if len(ai.Path) == 0 {
			// No path right now, just wait
			return time.Second
		}
		if ws, cs := m.StepActor(a, true, ai.Path[0]); ws || cs {
			// Step was successful, advance the path
			ai.Path = ai.Path[1:]
		} else {
			// Step was not successful, something has gone terribly wrong in
			// path-finding
			panic("invalid path")
		}
		return time.Duration(float64(time.Second) * a.WalkSpeed())
	})
	regActFn("zmActAttack", func(ai *AIModel, a *game.Actor, m *game.CityMap) time.Duration {
		min, max := a.DamageMinMax()
		m.Player.Damage(min, max, m.Now, a)
		return time.Duration(float64(time.Second) * a.ActSpeed())
	})
}
