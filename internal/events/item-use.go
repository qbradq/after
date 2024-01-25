package events

import (
	"strings"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/util"
)

func init() {
	rue("OpenDoor", openDoor)
	rue("CloseDoor", closeDoor)
}

func openDoor(i *game.Item, src *game.Actor, m *game.CityMap) error {
	ff := util.FloodFill{
		Matches: func(p util.Point) bool {
			items := m.ItemsAt(p)
			for _, item := range items {
				if item.TemplateID == i.TemplateID {
					return true
				}
			}
			return false
		},
		Set: func(p util.Point) {
			items := m.ItemsAt(p)
			for _, item := range items {
				if item.TemplateID == i.TemplateID {
					m.RemoveItem(item)
					ni := game.NewItem("Open"+item.TemplateID, m.Now, false)
					ni.Position = item.Position
					m.PlaceItem(ni, true)
				}
			}
		},
	}
	ff.Execute(i.Position)
	return nil
}

func closeDoor(i *game.Item, src *game.Actor, m *game.CityMap) error {
	ff := util.FloodFill{
		Matches: func(p util.Point) bool {
			items := m.ItemsAt(p)
			for _, item := range items {
				if item.TemplateID == i.TemplateID {
					return true
				}
			}
			return false
		},
		Set: func(p util.Point) {
			items := m.ItemsAt(p)
			for _, item := range items {
				if item.TemplateID == i.TemplateID {
					m.RemoveItem(item)
					s, _ := strings.CutPrefix(i.TemplateID, "Open")
					ni := game.NewItem(s, m.Now, false)
					ni.Position = item.Position
					m.PlaceItem(ni, true)
				}
			}
		},
	}
	ff.Execute(i.Position)
	return nil
}
