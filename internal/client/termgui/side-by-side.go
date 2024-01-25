package termgui

import (
	"fmt"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// sideBySide implements a side-by-side view of the player's inventory and a
// location on the ground or a container.
type sideBySide struct {
	OnRight        bool          // If false the focus should be on the left-hand menu
	LeftParent     *game.Item    // Parent of LeftContainer
	LeftContainer  *game.Item    // Container contents we are displaying at the left, nil means the actor's inventory
	RightParent    *game.Item    // Parent of RightContainer
	RightContainer *game.Item    // Container contents we are displaying at the right
	RightPoint     util.Point    // Location we are displaying items from at the right, only valid if RightContainer is nil
	done           bool          // If true we should exit the side-by-side mode
	ll             *itemList     // Left-side list
	rl             *itemList     // Right-side list
	m              *game.CityMap // City map we are using
	a              *game.Actor   // Actor who's inventory we are exploring on the left
}

// newSideBySide constructs a new side-by-side menu for use. The a parameter is
// the player's actor, c is the container item if any, and p is the point on the
// map (only valid if c is nil).
func newSideBySide(p *gameMode, m *game.CityMap, a *game.Actor) *sideBySide {
	ret := &sideBySide{
		ll: newItemList(),
		rl: newItemList(),
		m:  m,
		a:  a,
	}
	ret.ll.Selected = func(i *game.Item) {
		// Open sub-containers
		if i.Container && i != ret.LeftContainer {
			sbs := newSideBySide(p, m, a)
			sbs.LeftParent = ret.LeftContainer
			sbs.LeftContainer = i
			sbs.RightParent = ret.RightParent
			sbs.RightContainer = ret.RightContainer
			sbs.RightPoint = ret.RightPoint
			p.modeStack = append(p.modeStack, sbs)
			return
		}
		// Scan for selection of equipped items
		for _, e := range a.Equipment {
			if e == i {
				game.Log.Log(termui.ColorYellow, "You must take that off first.")
				return
			}
		}
		// Remove item from parent
		if i == ret.LeftContainer {
			ret.done = true
			if ret.LeftParent != nil {
				if !ret.LeftParent.RemoveItem(i) {
					return
				}
			} else {
				if !a.RemoveItemFromInventory(i) {
					return
				}
			}
		} else {
			if ret.LeftContainer != nil {
				if !ret.LeftContainer.RemoveItem(i) {
					ret.done = true
					return
				}
			} else {
				if !a.RemoveItemFromInventory(i) {
					ret.done = true
					return
				}
			}
		}
		// Drop item
		if ret.RightContainer != nil {
			if !ret.RightContainer.AddItem(i) {
				ret.done = true
				return
			}
		} else {
			i.Position = ret.RightPoint
			if !m.PlaceItem(i, false) {
				ret.done = true
				return
			}
		}
		// Logging
		game.Log.Log(termui.ColorAqua, fmt.Sprintf("You dropped %s.", i.Name))
	}
	ret.rl.Selected = func(i *game.Item) {
		// Open sub-containers
		if i.Container && i != ret.RightContainer {
			sbs := newSideBySide(p, m, a)
			sbs.OnRight = true
			sbs.LeftContainer = ret.LeftContainer
			sbs.LeftParent = ret.LeftParent
			sbs.RightParent = ret.RightContainer
			sbs.RightContainer = i
			sbs.RightPoint = ret.RightPoint
			p.modeStack = append(p.modeStack, sbs)
			return
		}
		// Make sure we don't pick up a fixed item
		if i.Fixed {
			game.Log.Log(termui.ColorYellow, "You cannot pick that up.")
			return
		}
		// Remove item from parent
		if i == ret.RightContainer {
			ret.done = true
			if ret.RightParent != nil {
				if !ret.RightParent.RemoveItem(i) {
					return
				}
			} else {
				m.RemoveItem(i)
			}
		} else if ret.RightContainer != nil {
			if !ret.RightContainer.RemoveItem(i) {
				ret.done = true
				return
			}
		} else {
			m.RemoveItem(i)
		}
		// Add item to left-hand container
		if ret.LeftContainer != nil {
			if !ret.LeftContainer.AddItem(i) {
				ret.done = true
				return
			}
		} else {
			if !a.AddItemToInventory(i) {
				ret.done = true
			}
		}
		// Log and cleanup
		game.Log.Log(termui.ColorAqua, fmt.Sprintf("You picked up %s.", i.Name))
	}
	return ret
}

// RepopulateLists (re)-populates the left and right lists
func (m *sideBySide) RepopulateLists() {
	if m.LeftContainer != nil {
		m.ll.SetItems(m.LeftContainer.Inventory, m.LeftContainer)
		m.ll.Title = m.LeftContainer.Name + " Contents"
	} else {
		// Build actor item list
		var items []*game.Item
		for _, i := range m.a.Equipment {
			if i == nil || !i.Container {
				continue
			}
			items = append(items, i)
		}
		if len(items) > 0 {
			items = append(items, nil)
		}
		items = append(items, m.a.Inventory...)
		m.ll.SetItems(items, nil)
		m.ll.Title = "Inventory"
	}
	if m.RightContainer != nil {
		m.rl.SetItems(m.RightContainer.Inventory, m.RightContainer)
		m.rl.Title = m.RightContainer.Name + " Contents"
	} else {
		m.rl.SetItems(m.m.ItemsAt(m.RightPoint), nil)
		m.rl.Title = "Items on Ground"
	}
}

// HandleEvent implements the termui.Mode interface.
func (m *sideBySide) HandleEvent(s termui.TerminalDriver, e any) error {
	m.RepopulateLists()
	switch ev := e.(type) {
	case *termui.EventKey:
		switch ev.Key {
		case 'h':
			fallthrough
		case 'l':
			m.OnRight = !m.OnRight
			return nil
		case '\033':
			return termui.ErrorQuit
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	if m.OnRight {
		m.rl.HandleEvent(s, e)
	} else {
		m.ll.HandleEvent(s, e)
	}
	if m.done {
		return termui.ErrorQuit
	}
	return nil
}

// Draw implements the termui.Mode interface.
func (m *sideBySide) Draw(s termui.TerminalDriver) {
	m.RepopulateLists()
	if len(m.ll.items) < 1 {
		m.OnRight = true
	}
	if len(m.rl.items) < 1 {
		m.OnRight = false
	}
	if m.OnRight {
		m.ll.list.HideCursor = true
		m.rl.list.HideCursor = false
	} else {
		m.ll.list.HideCursor = false
		m.rl.list.HideCursor = true
	}
	sb := util.NewRectWH(s.Size())
	ssb := sb.CenterRect(67, 22)
	m.ll.Bounds = ssb
	m.ll.Bounds.BR = ssb.TL.Add(util.NewPoint(33, 21))
	m.rl.Bounds = ssb
	m.rl.Bounds.TL = ssb.BR.Sub(util.NewPoint(33, 21))
	m.ll.Draw(s)
	m.rl.Draw(s)
}
