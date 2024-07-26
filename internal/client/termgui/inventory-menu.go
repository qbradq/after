package termgui

import (
	"strconv"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// inventoryMenu implements a menu to select one item from the inventory or
// set of currently equipped items.
type inventoryMenu struct {
	Bounds           util.Rect              // If this is the zero value the menu will be screen center
	Selected         func(*game.Item, bool) // Function called on valid selection, the first argument will never be nil and the second argument is true if the item is currently equipped by the actor
	IncludeEquipment bool                   // If true the actor's current equipment will be included in the list of items separated from the inventory by a horizontal bar
	OnlyEquipment    bool                   // If true only wearable and wield-able items are displayed
	OnlyUsable       bool                   // If true only items with a "Use" event are included
	Title            string                 // Title of the inventory menu
	actor            *game.Actor            // Pointer to the actor who's inventory we are exploring
	list             termui.List            // List used for the display and input
	items            []*game.Item           // Last list of items
	names            []string               // Last list of names
	fii              int                    // First inventory index
	ld               util.Point             // List dimensions without box
}

// newInventoryMenu configures a new InventoryMenu ready for use.
func newInventoryMenu(a *game.Actor) *inventoryMenu {
	var ret *inventoryMenu
	ret = &inventoryMenu{
		actor: a,
		list: termui.List{
			Boxed: true,
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
	return ret
}

// PopulateList (re)-populates the inventory menu from the actor's items,
// returning the number of items displayed.
func (m *inventoryMenu) PopulateList() int {
	fn := func(s string, i *game.Item) string {
		if i.Container {
			if len(i.Inventory) > 0 {
				return "+" + s
			}
			return "-" + s
		}
		return " " + s
	}
	m.ld.X = len(m.Title) + 2
	m.items = m.items[:0]
	m.names = m.names[:0]
	m.fii = 0
	if m.IncludeEquipment {
		for _, i := range m.actor.Equipment {
			if i == nil {
				continue
			}
			if m.OnlyEquipment && !i.Wearable && !i.Weapon {
				continue
			}
			if m.OnlyUsable && i.Events["Use"] == "" {
				continue
			}
			n := fn(i.Name, i)
			m.names = append(m.names, n)
			m.items = append(m.items, i)
			if m.ld.X < len(n) {
				m.ld.X = len(n)
			}
		}
		doWeapon := true
		if m.actor.Weapon == nil {
			doWeapon = false
		}
		if doWeapon && m.OnlyUsable && m.actor.Weapon.Events["Use"] == "" {
			doWeapon = false
		}
		if doWeapon {
			n := fn(m.actor.Weapon.Name, m.actor.Weapon)
			m.names = append(m.names, n)
			m.items = append(m.items, m.actor.Weapon)
			if m.ld.X < len(n) {
				m.ld.X = len(n)
			}
		}
		m.fii = len(m.items) + 1
		if len(m.items) > 0 && len(m.actor.Inventory) > 0 {
			m.names = append(m.names, "_hbar_")
			m.items = append(m.items, nil)
		}
	}
	for _, i := range m.actor.Inventory {
		if m.OnlyEquipment && !i.Wearable && !i.Weapon {
			continue
		}
		if m.OnlyUsable && i.Events["Use"] == "" {
			continue
		}
		n := fn(i.Name, i)
		if i.Amount > 1 {
			n = n + " x" + strconv.FormatInt(int64(i.Amount), 10)
		}
		m.names = append(m.names, n)
		m.items = append(m.items, i)
		if m.ld.X < len(n) {
			m.ld.X = len(n)
		}
	}
	m.ld.Y = len(m.items)
	m.list.CursorPos = 0
	return m.ld.Y
}

// HandleEvent implements the termui.Mode interface.
func (m *inventoryMenu) HandleEvent(s termui.TerminalDriver, e any) error {
	if err := m.list.HandleEvent(s, e); err != nil {
		return err
	}
	switch e.(type) {
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return nil
}

// Draw implements the termui.Mode interface.
func (m *inventoryMenu) Draw(s termui.TerminalDriver) {
	var rz util.Rect
	if m.Bounds == rz {
		w, h := s.Size()
		sb := util.NewRectWH(w, h)
		lb := sb.CenterRect(m.ld.X+2, m.ld.Y+2)
		m.list.Bounds = sb.Contain(lb)
	} else {
		m.list.Bounds = m.Bounds
	}
	m.list.Items = m.names
	m.list.Title = m.Title
	m.list.Draw(s)
}
