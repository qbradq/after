package ai

import (
	"time"

	"github.com/qbradq/after/internal/game"
)

// Nil implements an AI model that does nothing.
func init() {
	reg("Nil", func() *AIModel {
		return &AIModel{
			act:      "nil",
			periodic: "nil",
		}
	})
	regActFn("nil", func(ai *AIModel, a *game.Actor, m *game.CityMap) time.Duration {
		// Never act
		return time.Hour
	})
	regPUFn("nil", func(ai *AIModel, a *game.Actor, m *game.CityMap, d time.Duration) {
		// Standard regeneration
		days := float64(d) / float64(time.Hour*24)
		for i, p := range a.BodyParts {
			p.Health += days * 0.5 // Body parts heal in two days
			if p.Health > 1 {
				p.Health = 1
			}
			if !p.BrokenUntil.IsZero() && !m.Now.Before(p.BrokenUntil) {
				p.Broken = false
				p.BrokenUntil = time.Time{}
			}
			a.BodyParts[i] = p
		}
	})
}
