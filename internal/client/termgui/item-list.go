package termgui

import (
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// itemList implements a list menu that allows the player to select one item.
type itemList struct {
	Title    string           // Title of the item list
	Selected func(*game.Item) // Callback function on selection
	list     *termui.List     // List menu used
	items    []*game.Item     // The list of items to choose from
	ld       util.Point       // List dimensions
}

// newItemList constructs a new ItemList for use.
func newItemList() *itemList {
	var ret *itemList
	ret = &itemList{
		list: &termui.List{
			Boxed: true,
			Selected: func(s termui.TerminalDriver, i int) error {
				ret.Selected(ret.items[i])
				return termui.ErrorQuit
			},
		},
	}
	return ret
}

// SetItems sets the list of items the list selects from. If the container
// parameter is non-nil it will be considered for inclusion in the list.
func (m *itemList) SetItems(items []*game.Item, container *game.Item) {
	m.items = items
	m.list.Items = m.list.Items[:0]
	m.ld.X = len(m.Title) + 2
	m.ld.Y = len(items)
	if container != nil && container.Container && !container.Fixed {
		m.items = append([]*game.Item{container, nil}, items...)
		m.list.Items = append(m.list.Items, "-"+container.Name, "_hbar_")
	}
	for _, i := range items {
		n := " " + i.Name
		if i.Container && len(i.Inventory) > 0 {
			n = "+" + i.Name
		}
		m.list.Items = append(m.list.Items, n)
		if m.ld.X < len(i.Name) {
			m.ld.X = len(i.Name)
		}
	}
}

// HandleEvent implements the termui.Mode interface.
func (m *itemList) HandleEvent(s termui.TerminalDriver, e any) error {
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
func (m *itemList) Draw(s termui.TerminalDriver) {
	sb := util.NewRectWH(s.Size())
	lb := sb.CenterRect(m.ld.X+2, m.ld.Y+2)
	m.list.Bounds = sb.Contain(lb)
	m.list.Title = m.Title
	m.list.Draw(s)
}
