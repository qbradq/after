package events

import (
	"time"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
)

func init() {
	rpue("ResurrectCorpse", resurrectCorpse)

}

func resurrectCorpse(i *game.Item, m *game.CityMap, d time.Duration) error {
	if !m.Now.Before(i.TArg) && m.ActorAt(i.Position) == nil {
		a := game.NewActor(i.SArg, m.Now, false)
		a.Position = i.Position
		for _, c := range i.Inventory {
			if a.WieldItem(c) == "" {
				continue
			}
			if a.WearItem(c) == "" {
				continue
			}
			if !a.AddItemToInventory(c) {
				game.Log.Log(
					termui.ColorRed,
					"Resurrection Error: Unable to stow %s",
					c.DisplayName(),
				)
			}
		}
		m.PlaceActor(a, true)
		i.Destroyed = true
	}
	return nil
}
