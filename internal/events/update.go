package events

import (
	"time"

	"github.com/qbradq/after/internal/game"
)

func init() {
	rpue("ResurrectCorpse", resurrectCorpse)

}

func resurrectCorpse(i *game.Item, m *game.CityMap, d time.Duration) error {
	if !m.Now.Before(i.TArg) && m.ActorAt(i.Position) == nil {
		a := game.NewActor(i.SArg, m.Now)
		a.Position = i.Position
		m.PlaceActor(a, true)
		i.Destroyed = true
	}
	return nil
}
