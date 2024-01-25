package termgui

import (
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// sideBySide implements a side-by-side view of the player's inventory and a
// location on the ground or a container.
type sideBySide struct {
	InContainer       bool                   // In container flag, if false the focus should be on the left-hand menu
	InventorySelected func(*game.Item, bool) // Called when an item from the inventory was selected
	ItemListSelected  func(*game.Item)       // Called when an item from the item list is selected
	Done              bool                   // If true we should exit the side-by-side mode
	im                *inventoryMenu         // Menu used to display the player's inventory on the left-hand side
	il                *itemList              // Item list used to display the location or container contents
	c                 *game.Item             // Container we are looking into, if any
	p                 util.Point             // Point on the map we are inspecting, only valid if c is nil
	m                 *game.CityMap          // City map we are working on, only valid if c is nil
}

// newSideBySide constructs a new side-by-side menu for use. The a parameter is
// the player's actor, c is the container item if any, and p is the point on the
// map (only valid if c is nil).
func newSideBySide(m *game.CityMap, a *game.Actor, c *game.Item, p util.Point) *sideBySide {
	ret := &sideBySide{
		im: newInventoryMenu(a),
		il: newItemList(),
		c:  c,
		p:  p,
		m:  m,
	}
	ret.im.Title = "Inventory"
	ret.im.Selected = func(i *game.Item, equipped bool) {
		ret.InventorySelected(i, equipped)
		ret.repopulateLists()
	}
	ret.il.Selected = func(i *game.Item) {
		ret.ItemListSelected(i)
		ret.repopulateLists()
	}
	ret.repopulateLists()
	return ret
}

func (m *sideBySide) repopulateLists() {
	m.im.PopulateList()
	if m.c != nil {
		m.il.SetItems(m.c.Inventory, m.c)
		m.il.Title = m.c.Name + " Contents"
	} else {
		m.il.SetItems(m.m.ItemsAt(m.p), nil)
		m.il.Title = "Items on Ground"
	}
}

// HandleEvent implements the termui.Mode interface.
func (m *sideBySide) HandleEvent(s termui.TerminalDriver, e any) error {
	switch ev := e.(type) {
	case *termui.EventKey:
		switch ev.Key {
		case 'h':
			fallthrough
		case 'l':
			m.InContainer = !m.InContainer
			return nil
		case '\033':
			return termui.ErrorQuit
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	if m.InContainer {
		m.il.HandleEvent(s, e)
	} else {
		m.im.HandleEvent(s, e)
	}
	if m.Done {
		return termui.ErrorQuit
	}
	return nil
}

// Draw implements the termui.Mode interface.
func (m *sideBySide) Draw(s termui.TerminalDriver) {
	if m.InContainer {
		m.im.list.HideCursor = true
		m.il.list.HideCursor = false
	} else {
		m.im.list.HideCursor = false
		m.il.list.HideCursor = true
	}
	sb := util.NewRectWH(s.Size())
	ssb := sb.CenterRect(67, 22)
	m.im.Bounds = ssb
	m.im.Bounds.BR = ssb.TL.Add(util.NewPoint(33, 21))
	m.il.Bounds = ssb
	m.il.Bounds.TL = ssb.BR.Sub(util.NewPoint(33, 21))
	m.im.Draw(s)
	m.il.Draw(s)
}
