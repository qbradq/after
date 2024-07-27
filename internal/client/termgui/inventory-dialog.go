package termgui

import (
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// inventoryDialogLine contains the information for one entry in a dialog panel.
type inventoryDialogLine struct {
	item *game.Item // Item being referenced, if any
	text string     // Text of the line
}

// inventoryDialogPanel contains the common inventory panel methods.
type inventoryDialogPanel struct {
	top      int                   // Top line index
	selected int                   // Index of the selected line
	source   any                   // Source of the inventory
	title    string                // Title for the panel
	lines    []inventoryDialogLine // All of the lines of the panel
	size     util.Point            // Size of the panel
}

// newInventoryDialogPanel creates a new inventoryDialogPanel ready for use.
func newInventoryDialogPanel(source any, cm *game.CityMap) *inventoryDialogPanel {
	ret := &inventoryDialogPanel{
		size: util.NewPoint(35, 21),
	}
	ret.setSource(source, cm)
	ret.setSelected(0, false, true)
	return ret
}

// setTop sets the top line based on the selected line.
func (m *inventoryDialogPanel) setTop() {
	m.top = m.selected - (m.size.Y / 2)
	// Bound bottom
	if m.top+m.size.Y >= len(m.lines) {
		m.top = len(m.lines) - m.size.Y
	}
	// Bound top
	if m.top < 0 {
		m.top = 0
	}
}

// setSelected sets the selected line relative to the top. If the up parameter
// is true, labels will be skipped going up, otherwise down. If the popUp
// parameter is true, the initial index is bound to the bottom.
func (m *inventoryDialogPanel) setSelected(idx int, up bool, popUp bool) {
	// Empty case
	if len(m.lines) == 0 {
		m.selected = -1
		m.top = 0
		return
	}
	// Indexing and popup handling
	idx += m.top
	if idx >= len(m.lines) {
		idx = 0
	}
	if idx < 0 {
		idx = len(m.lines) - 1
	}
	if popUp {
		if m.lines[idx].item == nil {
			idx++
		}
	}
	// Line skipping
	for {
		if idx < 0 {
			idx = len(m.lines) - 1
		}
		if idx >= len(m.lines) {
			idx = 0
		}
		l := m.lines[idx]
		if l.item != nil {
			break
		}
		if up {
			idx--
		} else {
			idx++
		}
	}
	m.selected = idx
	m.setTop()
}

// linesForActor returns a slice of the dialog items appropriate for an actor.
func (m *inventoryDialogPanel) linesForActor(a *game.Actor) []inventoryDialogLine {
	// Equipped items
	ret := []inventoryDialogLine{
		{
			text: "Equipment",
		},
	}
	if a.Weapon != nil {
		ret = append(ret, inventoryDialogLine{
			item: a.Weapon,
			text: a.Weapon.UIDisplayName(),
		})
	}
	for _, i := range a.Equipment {
		if i == nil {
			continue
		}
		ret = append(ret, inventoryDialogLine{
			item: i,
			text: i.UIDisplayName(),
		})
	}
	// Trim empty equipment list
	if len(ret) == 1 {
		ret = ret[:0]
	}
	// Top-level inventory
	if len(a.Inventory) > 0 {
		// Skip empty inventory header
		ret = append(ret, inventoryDialogLine{
			text: "Inventory",
		})
	}
	for _, i := range a.Inventory {
		ret = append(ret, inventoryDialogLine{
			item: i,
			text: i.UIDisplayName(),
		})
	}
	return ret
}

// linesForContainer returns a slice of the dialog items appropriate for the
// container.
func (m *inventoryDialogPanel) linesForContainer(c *game.Item) []inventoryDialogLine {
	ret := []inventoryDialogLine{}
	for _, i := range c.Inventory {
		ret = append(ret, inventoryDialogLine{
			item: i,
			text: i.UIDisplayName(),
		})
	}
	return ret
}

// linesForPoint returns a slice of the dialog items appropriate for the map
// location.
func (m *inventoryDialogPanel) linesForPoint(cm *game.CityMap, p util.Point) []inventoryDialogLine {
	ret := []inventoryDialogLine{}
	// Tile contents
	for _, i := range cm.ItemsAt(p) {
		ret = append(ret, inventoryDialogLine{
			item: i,
			text: i.UIDisplayName(),
		})
	}
	return ret
}

// setSource sets the source for the panel and updates the lines.
func (m *inventoryDialogPanel) setSource(source any, cm *game.CityMap) {
	m.source = source
	m.refreshSource(cm)
}

// refreshSource refreshes the internal data cache from the source.
func (m *inventoryDialogPanel) refreshSource(cm *game.CityMap) {
	if m.source == nil {
		return
	}
	switch t := m.source.(type) {
	case *game.Actor:
		m.lines = m.linesForActor(t)
		m.title = t.Name
	case *game.Item:
		m.lines = m.linesForContainer(t)
		m.title = t.Name
	case util.Point:
		m.lines = m.linesForPoint(cm, t)
		tile := cm.GetTile(t)
		m.title = tile.Name
	}
	if len(m.lines) == 0 {
		return
	}
	if m.selected < 0 {
		m.selected = 0
	}
	if m.selected >= len(m.lines) {
		m.selected = len(m.lines) - 1
	}
	if m.lines[m.selected].item == nil {
		m.selected++
	}
	if m.selected >= len(m.lines) {
		m.selected = len(m.lines) - 1
	}
}

// getSelectedItem returns the currently selected item from the source.
func (m *inventoryDialogPanel) getSelectedItem() *game.Item {
	if len(m.lines) < 1 || m.selected >= len(m.lines) {
		return nil
	}
	return m.lines[m.selected].item
}

// removeSelectedItem removes the currently selected item from the source.
func (m *inventoryDialogPanel) removeSelectedItem(cm *game.CityMap) *game.Item {
	i := m.getSelectedItem()
	if i.Fixed {
		game.Log.Log(termui.ColorYellow, "You cannot pick that up.")
		return nil
	}
	switch t := m.source.(type) {
	case *game.Actor:
		if !t.RemoveItemFromInventory(i) {
			if i == t.Weapon || i == t.Equipment[i.WornBodyPart] {
				game.Log.Log(termui.ColorYellow, "You must take that off first.")
			}
			return nil
		}
		return i
	case *game.Item:
		if !t.RemoveItem(i) {
			return nil
		}
		return i
	case util.Point:
		if i.Position != t {
			return nil
		}
		if !cm.RemoveItem(i) {
			return nil
		}
		return i
	default:
		return nil
	}
}

// addItem adds the given item to the source.
func (m *inventoryDialogPanel) addItem(i *game.Item, cm *game.CityMap) bool {
	switch t := m.source.(type) {
	case *game.Actor:
		return t.AddItemToInventory(i)
	case *game.Item:
		return t.AddItem(i)
	case util.Point:
		i.Position = t
		return cm.PlaceItem(i, false)
	default:
		return false
	}
}

// Draw draws the panel at the given point. If the parameter si is non-negative,
// it is the index of the line to highlight.
func (m *inventoryDialogPanel) Draw(s termui.TerminalDriver, tl util.Point, si int, cm *game.CityMap, focused bool) {
	// Frame drawing
	b := util.NewRectXYWH(tl.X, tl.Y, m.size.X, m.size.Y)
	ns := termui.CurrentTheme.Normal
	if focused {
		ns = ns.Background(termui.ColorNavy)
	}
	termui.DrawBox(s, b, ns)
	termui.DrawStringCenter(s, b, m.title, ns)
	// Draw items list
	b.TL.X++
	b.TL.Y++
	b.BR.X--
	b.BR.Y--
	for y := 0; y < b.Height(); y++ {
		idx := y + m.top
		if idx >= len(m.lines) {
			break
		}
		i := m.lines[idx]
		if i.item == nil {
			// Section title
			termui.DrawStringCenter(s, b, i.text, termui.CurrentTheme.Emphasis)
		} else {
			// Item
			ns := termui.CurrentTheme.Normal
			if idx == si {
				// Selected item name
				ns = termui.CurrentTheme.Highlight
			}
			// Mark fixed items
			if i.item.Fixed {
				ns = ns.Foreground(termui.ColorSilver)
			}
			db := b
			db.BR.Y = db.TL.Y
			termui.DrawFill(s, db, termui.Glyph{
				Rune:  ' ',
				Style: ns,
			})
			termui.DrawStringLeft(s, b, i.item.UIDisplayName(), ns)
		}
		b.TL.Y++
	}
}

// inventoryDialog implements a dialog for managing and interacting with the
// player's inventory.
type inventoryDialog struct {
	OnRight      bool                    // If true, the cursor is in the right-hand pane
	SelectedLine int                     // Index of the selected index
	m            *game.CityMap           // CityMap we are getting the player from
	left         []*inventoryDialogPanel // Left-hand panel
	right        []*inventoryDialogPanel // Right-hand panel
}

// newInventoryDialog creates a new inventoryDialog ready to use.
func newInventoryDialog(m *game.CityMap, left, right any) *inventoryDialog {
	ret := &inventoryDialog{
		m:     m,
		left:  []*inventoryDialogPanel{newInventoryDialogPanel(left, m)},
		right: []*inventoryDialogPanel{newInventoryDialogPanel(right, m)},
	}
	return ret
}

// HandleEvent implements the termui.Mode interface.
func (m *inventoryDialog) HandleEvent(s termui.TerminalDriver, e any) error {
	switchSource := func(source any) {
		if p, ok := source.(util.Point); ok {
			source = m.m.Player.Position.Add(p)
		}
		if m.OnRight {
			m.right = m.right[:0]
			m.right = append(m.right, newInventoryDialogPanel(source, m.m))
		} else {
			m.left = m.left[:0]
			m.left = append(m.left, newInventoryDialogPanel(source, m.m))
		}
	}
	left := m.left[len(m.left)-1]
	right := m.right[len(m.right)-1]
	switch evt := e.(type) {
	case *termui.EventKey:
		switch evt.Key {
		case 'h': // Cursor left
			if !m.OnRight {
				break
			}
			m.OnRight = false
			left.setSelected(right.selected-right.top, false, true)
		case 'l': // Cursor right
			if m.OnRight {
				break
			}
			m.OnRight = true
			right.setSelected(left.selected-left.top, false, true)
		case 'j': // Cursor down
			if m.OnRight {
				right.setSelected(right.selected+1, false, false)
			} else {
				left.setSelected(left.selected+1, false, false)
			}
		case 'k': // Cursor up
			if m.OnRight {
				right.setSelected(right.selected-1, true, false)
			} else {
				left.setSelected(left.selected-1, true, false)
			}
		case ' ':
			fallthrough
		case '\n': // Open container
			var c *game.Item
			if m.OnRight {
				c = right.lines[right.selected].item
			} else {
				c = left.lines[left.selected].item
			}
			if c != nil && c.Container {
				if m.OnRight {
					m.right = append(m.right, newInventoryDialogPanel(c, m.m))
				} else {
					m.left = append(m.left, newInventoryDialogPanel(c, m.m))
				}
			}
		case 'm':
			var i *game.Item
			var s *inventoryDialogPanel
			var t *inventoryDialogPanel
			if m.OnRight {
				s = right
				t = left
			} else {
				s = left
				t = right
			}
			if i = s.removeSelectedItem(m.m); i == nil {
				break
			}
			if !t.addItem(i, m.m) {
				s.addItem(i, m.m)
			}
			left.refreshSource(m.m)
			right.refreshSource(m.m)
		case 'w':
			// Source and i selection
			var i *game.Item
			var s *inventoryDialogPanel
			if m.OnRight {
				s = right
			} else {
				s = left
			}
			i = s.getSelectedItem()
			if i == nil || (!i.Weapon && !i.Wearable) {
				break
			}
			if i.Weapon {
				// Weapon handling
				if i == m.m.Player.Weapon {
					// Unwield request
					if !m.m.Player.UnWieldItem(i) {
						if !s.addItem(i, m.m) {
							game.Log.Log(termui.ColorYellow, "Unable to stow %s.", i.Name)
							m.m.Player.WieldItem(i)
						}
					} else {
						game.Log.Log(termui.ColorAqua, "You stowed %s.", i.Name)
					}
				} else {
					// Wield request
					if m.m.Player.Weapon != nil {
						game.Log.Log(termui.ColorYellow, "Already wielding an item.")
					} else if s.removeSelectedItem(m.m) == nil {
						game.Log.Log(termui.ColorYellow, "Unable to get %s from inventory.", i.Name)
					} else if r := m.m.Player.WieldItem(i); r != "" {
						game.Log.Log(termui.ColorYellow, r)
						s.addItem(i, m.m)
					} else {
						game.Log.Log(termui.ColorAqua, "You wielded %s.", i.Name)
					}
				}
			} else {
				// Equipment handling
				if i == m.m.Player.Equipment[i.WornBodyPart] {
					// Unwear request
					if !m.m.Player.UnWearItem(i) {
						game.Log.Log(termui.ColorYellow, "Unable to take off %s.", i.Name)
					} else if m.m.Player.AddItemToInventory(i) {
						game.Log.Log(termui.ColorAqua, "You took off %s.", i.Name)
					} else {
						game.Log.Log(termui.ColorYellow, "Unable to stow %s.", i.Name)
						m.m.Player.WearItem(i)
					}
				} else {
					// Wear request
					if m.m.Player.Equipment[i.WornBodyPart] != nil {
						game.Log.Log(termui.ColorYellow, "You are already wearing something there.")
					} else if s.removeSelectedItem(m.m) == nil {
						game.Log.Log(termui.ColorYellow, "Unable to get %s from inventory.", i.Name)
					} else if r := m.m.Player.WearItem(i); r != "" {
						game.Log.Log(termui.ColorYellow, r)
						s.addItem(i, m.m)
					} else {
						game.Log.Log(termui.ColorAqua, "You put on %s", i.Name)
					}
				}
			}
			left.refreshSource(m.m)
			right.refreshSource(m.m)
		case '\010':
			if m.OnRight {
				if len(m.right) > 1 {
					m.right = m.right[:len(m.right)-1]
					m.right[len(m.right)-1].refreshSource(m.m)
				}
			} else {
				if len(m.left) > 1 {
					m.left = m.left[:len(m.left)-1]
					m.left[len(m.left)-1].refreshSource(m.m)
				}
			}
		case ',':
			switchSource(util.NewPoint(0, 0))
		case '1':
			switchSource(util.NewPoint(-1, 1))
		case '2':
			switchSource(util.NewPoint(0, 1))
		case '3':
			switchSource(util.NewPoint(1, 1))
		case '4':
			switchSource(util.NewPoint(-1, 0))
		case '5':
			switchSource(util.NewPoint(0, 0))
		case '6':
			switchSource(util.NewPoint(1, 0))
		case '7':
			switchSource(util.NewPoint(-1, -1))
		case '8':
			switchSource(util.NewPoint(0, -1))
		case '9':
			switchSource(util.NewPoint(1, -1))
		case 'i':
			switchSource(&m.m.Player.Actor)
		case '\033':
			return termui.ErrorQuit
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return nil
}

// Draw implements the termui.Mode interface.
func (m *inventoryDialog) Draw(s termui.TerminalDriver) {
	sb := util.NewRectWH(s.Size())
	b := sb.CenterRect(70, 24)
	termui.DrawFill(s, b, termui.Glyph{
		Rune:  ' ',
		Style: termui.CurrentTheme.Normal,
	})
	// Help frame
	db := b
	db.TL.Y += 21
	termui.DrawBox(s, db, termui.CurrentTheme.Normal)
	termui.DrawStringCenter(s, db,
		"[hjkl] Navigate [SPACE] Open [BACK] Close [m] Move Item [w] (Un)Wear",
		termui.CurrentTheme.Normal.Foreground(termui.ColorLime),
	)
	db.TL.Y++
	termui.DrawStringCenter(s, db,
		"Quick Change: [i] Inventory [,] At Feet [NUMPAD] Adjacent",
		termui.CurrentTheme.Normal.Foreground(termui.ColorLime),
	)
	// Left-hand display
	left := m.left[len(m.left)-1]
	db = b
	idx := -1
	if !m.OnRight {
		idx = left.selected - left.top
	}
	left.Draw(s, db.TL, idx, m.m, !m.OnRight)
	// Right-hand display
	right := m.right[len(m.right)-1]
	db = b
	db.TL.X += 35
	idx = -1
	if m.OnRight {
		idx = right.selected - right.top
	}
	right.Draw(s, db.TL, idx, m.m, m.OnRight)
}
