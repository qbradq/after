package ai

import (
	"time"

	"github.com/qbradq/after/internal/game"
)

// Nil implements an AI model that does nothing.
func init() {
	reg("Nil", func() *AIModel {
		return &AIModel{
			act: "nil",
		}
	})
	regFn("nil", func(a1 *AIModel, a2 *game.Actor, t time.Time, cm *game.CityMap) time.Duration {
		return time.Second * 60
	})
}
