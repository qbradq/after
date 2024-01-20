package termgui

import (
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// InventoryMenu implements a menu to select one item from the inventory or
// set of currently equipped items.
type InventoryMenu struct {
	Actor    *game.Actor            // Pointer to the actor who's inventory we are exploring
	List     termui.List            // List used for the display and input
	Selected func(*game.Item, bool) // Function called on valid selection, the first argument will never be nil and the second argument is true if the item is currently equipped by the actor
	items    []*game.Item           // Last list of items
	names    []string               // Last list of names
	fii      int                    // First inventory index
}

// NewInventoryMenu configures a new InventoryMenu ready for use.
func NewInventoryMenu(a *game.Actor, t string) *InventoryMenu {
	var ret *InventoryMenu
	ret = &InventoryMenu{
		Actor: a,
		List: termui.List{
			Boxed: true,
			Title: t,
			Selected: func(td termui.TerminalDriver, i int) error {
				item := ret.items[i]
				if item == nil {
					return nil
				}
				ret.Selected(item, i < ret.fii)
				return termui.ErrorQuit
			},
		},
	}
	ret.PopulateList()
	return ret
}

func (m *InventoryMenu) PopulateList() {
	m.items = m.items[:0]
	m.names = m.names[:0]
	for _, i := range m.Actor.Equipment {
		if i == nil {
			continue
		}
		m.names = append(m.names, i.Name)
		m.items = append(m.items, i)
	}
	if m.Actor.Weapon != nil {
		m.names = append(m.names, m.Actor.Weapon.Name)
		m.items = append(m.items, m.Actor.Weapon)
	}
	m.fii = len(m.items) + 1
	if len(m.items) == 0 {
		m.fii = 0
	} else {
		m.names = append(m.names, "_hbar_")
		m.items = append(m.items, nil)
	}
	for _, i := range m.Actor.Inventory {
		m.names = append(m.names, i.Name)
		m.items = append(m.items, i)
	}
	m.List.CursorPos = 0
}

// HandleEvent implements the termui.Mode interface.
func (m *InventoryMenu) HandleEvent(s termui.TerminalDriver, e any) error {
	if err := m.List.HandleEvent(s, e); err != nil {
		return err
	}
	switch e.(type) {
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return nil
}

// Draw implements the termui.Mode interface.
func (m *InventoryMenu) Draw(s termui.TerminalDriver) {
	_, h := s.Size()
	m.List.Bounds = util.NewRectWH(60, h-6)
	m.List.Items = m.names
	m.List.Draw(s)
}
