package termgui

import (
	"errors"
	"time"

	"github.com/qbradq/after/internal/events"
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// gameMode implements the top-level client interface.
type gameMode struct {
	CityMap       *game.CityMap  // The city we are playing
	logMode       *logMode       // Log display
	mapMode       *mapMode       // Map display
	minimap       *minimap       // Small mini-map
	status        *statusPanel   // Status panel
	inventory     *inventoryMenu // Inventory menu
	escapeMenu    *escapeMenu    // Escape menu
	itemList      *itemList      // Item list menu
	confirmDialog *confirmDialog // Confirmation dialog
	modeStack     []termui.Mode  // Internal stack of mode that overlay the main game mode, like the escape menu or inventory screen
	quit          bool           // If true we should quit
	inTarget      bool           // If true we are in targeting mode
	debug         bool           // If true display debug information
}

// newGameMode returns a new game mode.
func newGameMode(m *game.CityMap) *gameMode {
	gm := &gameMode{
		CityMap: m,
		logMode: &logMode{},
		mapMode: &mapMode{
			CityMap: m,
			Center:  m.Player.Position,
		},
		minimap: &minimap{
			CityMap:     m,
			CursorStyle: 1,
		},
		status: &statusPanel{
			CityMap: m,
		},
		inventory:     newInventoryMenu(&m.Player.Actor),
		itemList:      newItemList(),
		confirmDialog: newConfirmDialog(),
	}
	gm.escapeMenu = newEscapeMenu(gm)
	game.Log = gm.logMode
	game.Log.Log(termui.ColorTeal, "Welcome to the aftermath!")
	return gm
}

func (m *gameMode) handleEventInternal(s termui.TerminalDriver, e any) error {
	getAt := func(p util.Point) {
		// If there's nothing there we can skip it
		items := m.CityMap.ItemsAt(p)
		if len(items) < 1 {
			return
		}
		// If there is only a single container we auto-open it
		if len(items) == 1 && items[0].Container {
			sbs := newSideBySide(m, m.CityMap, &m.CityMap.Player.Actor)
			sbs.OnRight = true
			sbs.RightContainer = items[0]
			m.modeStack = append(m.modeStack, sbs)
			return
		}
		// Otherwise we just open the side-by-side for the ground point
		sbs := newSideBySide(m, m.CityMap, &m.CityMap.Player.Actor)
		sbs.OnRight = true
		sbs.RightPoint = p
		m.modeStack = append(m.modeStack, sbs)
	}
	dropAt := func(p util.Point) {
		sbs := newSideBySide(m, m.CityMap, &m.CityMap.Player.Actor)
		sbs.RightPoint = p
		m.modeStack = append(m.modeStack, sbs)
	}
	// Check every event for end game conditions
	if m.CityMap.Player.Dead {
		if ev, ok := e.(*termui.EventKey); ok {
			if ev.Key == '\033' {
				return termui.ErrorQuit
			}
		}
		return nil
	}
	// Normal input processing
	dir := util.DirectionInvalid
	switch ev := e.(type) {
	case *termui.EventKey:
		switch ev.Key {
		case 'u': // Walk North East
			dir = util.DirectionNorthEast
		case 'y': // Walk North West
			dir = util.DirectionNorthWest
		case 'n': // Walk South East
			dir = util.DirectionSouthEast
		case 'b': // Walk South West
			dir = util.DirectionSouthWest
		case 'l': // Walk East
			dir = util.DirectionEast
		case 'h': // Walk West
			dir = util.DirectionWest
		case 'j': // Walk South
			dir = util.DirectionSouth
		case 'k': // Walk North
			dir = util.DirectionNorth
		case '.': // Wait one second
			m.CityMap.PlayerTookTurn(time.Second)
			s.FlushEvents()
			return nil
		case 'x': // eXamine surroundings
			m.inTarget = true
			m.mapMode.Callback = func(p util.Point, b bool) error {
				m.inTarget = false
				return nil
			}
			m.mapMode.Center = m.CityMap.Player.Position
			m.mapMode.CursorPos = m.CityMap.Player.Position
			m.mapMode.CursorRange = 0
			return nil
		case 'm': // Minimap
			termui.RunMode(s, &minimap{
				CityMap:     m.CityMap,
				Bounds:      util.NewRectWH(s.Size()),
				Center:      util.NewPoint(m.CityMap.Player.Position.X/game.ChunkWidth, m.CityMap.Player.Position.Y/game.ChunkHeight),
				CursorStyle: 2,
				DrawInfo:    true,
			})
			return nil
		case 'U': // Use item in surroundings
			m.inTarget = true
			m.mapMode.Callback = func(p util.Point, b bool) error {
				m.inTarget = false
				if !b {
					return nil
				}
				items := m.CityMap.ItemsAt(p)
				if len(items) > 0 {
					err, used := events.ExecuteItemUseEvent("Use", items[len(items)-1], &m.CityMap.Player.Actor, m.CityMap)
					if err != nil {
						return err
					}
					if used {
						m.CityMap.PlayerTookTurn(time.Second)
					}
				}
				return nil
			}
			m.mapMode.Center = m.CityMap.Player.Position
			m.mapMode.CursorPos = m.CityMap.Player.Position
			m.mapMode.CursorRange = 1
			return nil
		case 'a': // Attack
			m.inTarget = true
			m.mapMode.Callback = func(p util.Point, b bool) error {
				m.inTarget = false
				if !b {
					return nil
				}
				a := m.CityMap.ActorAt(p)
				if a != nil {
					a.Damage(m.CityMap.Player.MinDamage, m.CityMap.Player.MaxDamage, m.CityMap.Now, &m.CityMap.Player.Actor)
					m.CityMap.PlayerTookTurn(time.Second)
				}
				return nil
			}
			m.mapMode.Center = m.CityMap.Player.Position
			m.mapMode.CursorPos = m.CityMap.Player.Position
			m.mapMode.CursorRange = 1
			return nil
		case 'w': // Wear / Un-wear / Wield / Un-wield equipment
			m.inventory.Selected = func(i *game.Item, equipped bool) {
				if equipped {
					if i.Weapon {
						if !m.CityMap.Player.UnWieldItem(i) {
							return
						}
						m.CityMap.Player.AddItemToInventory(i)
						m.logMode.Log(termui.ColorAqua, "Stopped wielding %s.", i.Name)
					} else {
						if !m.CityMap.Player.UnWearItem(i) {
							return
						}
						m.CityMap.Player.AddItemToInventory(i)
						m.logMode.Log(termui.ColorAqua, "Took off %s.", i.Name)
					}
				} else {
					if !m.CityMap.Player.RemoveItemFromInventory(i) {
						return
					}
					if i.Weapon {
						if r := m.CityMap.Player.WieldItem(i); r != "" {
							m.logMode.Log(termui.ColorYellow, r)
							m.CityMap.Player.AddItemToInventory(i)
							return
						}
						m.logMode.Log(termui.ColorAqua, "Wielded %s.", i.Name)
					} else if i.Wearable {
						if r := m.CityMap.Player.WearItem(i); r != "" {
							m.logMode.Log(termui.ColorYellow, r)
							m.CityMap.Player.AddItemToInventory(i)
							return
						}
						m.logMode.Log(termui.ColorAqua, "Wore %s.", i.Name)
					} else {
						m.logMode.Log(termui.ColorYellow, "That item is not wearable.")
						m.CityMap.Player.AddItemToInventory(i)
						return
					}
				}
				m.CityMap.PlayerTookTurn(time.Duration(float64(time.Second) * m.CityMap.Player.ActSpeed()))
			}
			m.inventory.Title = "Wear / Un Wear Item"
			m.inventory.IncludeEquipment = true
			if m.inventory.PopulateList() > 0 {
				m.modeStack = append(m.modeStack, m.inventory)
			}
			return nil
		case 'd': // Drop item from inventory
			dropAt(m.CityMap.Player.Position)
			return nil
		case 'D': // targeted Drop from inventory
			m.inventory.Selected = func(i *game.Item, equipped bool) {
				m.inTarget = true
				m.mapMode.Callback = func(p util.Point, confirmed bool) error {
					m.inTarget = false
					if !confirmed {
						return nil
					}
					dropAt(p)
					return nil
				}
				m.mapMode.Center = m.CityMap.Player.Position
				m.mapMode.CursorPos = m.CityMap.Player.Position
				m.mapMode.CursorRange = 1
			}
			m.inventory.IncludeEquipment = false
			m.inventory.Title = "Drop Item"
			if m.inventory.PopulateList() > 0 {
				m.modeStack = append(m.modeStack, m.inventory)
			}
			m.logMode.Log(termui.ColorPurple, "Drop where?")
			return nil
		case ',': // get items at feet (,)
			getAt(m.CityMap.Player.Position)
			return nil
		case 'g': // Get items within reach
			m.inTarget = true
			m.mapMode.Callback = func(p util.Point, confirmed bool) error {
				m.inTarget = false
				if !confirmed {
					return nil
				}
				getAt(p)
				return nil
			}
			m.mapMode.Center = m.CityMap.Player.Position
			m.mapMode.CursorPos = m.CityMap.Player.Position
			m.mapMode.CursorRange = 1
			m.logMode.Log(termui.ColorPurple, "Get where?")
			return nil
		case 'c': // Climb
			m.inTarget = true
			m.mapMode.Callback = func(p util.Point, confirmed bool) error {
				m.inTarget = false
				if !confirmed {
					return nil
				}
				d := m.CityMap.Player.Position.DirectionTo(p)
				m.CityMap.StepPlayer(true, d)
				return nil
			}
			m.mapMode.Center = m.CityMap.Player.Position
			m.mapMode.CursorPos = m.CityMap.Player.Position
			m.mapMode.CursorRange = 1
			m.logMode.Log(termui.ColorPurple, "Climb where?")
			return nil
		case '\033':
			m.modeStack = append(m.modeStack, newEscapeMenu(m))
			return nil
		default:
			// Unhandled key, just ignore it
			return nil
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	dir = dir.Bound()
	if dir != util.DirectionInvalid {
		if !m.CityMap.StepPlayer(false, dir) {
			// Bump handling
			if m.CityMap.PlayerCanClimb(dir) {
				// Offer to climb over whatever is blocking us
				m.confirmDialog.Title = "Confirm Climb"
				m.confirmDialog.Prompt = "Do you wish to climb?"
				m.confirmDialog.Confirmed = func() {
					m.CityMap.StepPlayer(true, dir)
				}
				m.modeStack = append(m.modeStack, m.confirmDialog)
				return nil
			}
			np := m.CityMap.Player.Position.Step(dir)
			a := m.CityMap.ActorAt(np)
			if a != nil {
				a.Damage(m.CityMap.Player.MinDamage, m.CityMap.Player.MaxDamage, m.CityMap.Now, &m.CityMap.Player.Actor)
				m.CityMap.PlayerTookTurn(time.Duration(float64(time.Second) * m.CityMap.Player.ActSpeed()))
				s.FlushEvents()
			} else {
				items := m.CityMap.ItemsAt(np)
				if len(items) > 0 {
					err, used := events.ExecuteItemUseEvent("Use", items[len(items)-1], &m.CityMap.Player.Actor, m.CityMap)
					if err != nil {
						return err
					}
					if used {
						m.CityMap.PlayerTookTurn(time.Duration(float64(time.Second) * m.CityMap.Player.ActSpeed()))
					}
					s.FlushEvents()
				}
			}
		} else {
			s.FlushEvents()
		}
	}
	return nil
}

// HandleEvent implements the termui.Mode interface.
func (m *gameMode) HandleEvent(s termui.TerminalDriver, e any) error {
	if len(m.modeStack) > 0 {
		err := m.modeStack[len(m.modeStack)-1].HandleEvent(s, e)
		if m.quit {
			return termui.ErrorQuit
		}
		if errors.Is(err, termui.ErrorQuit) {
			m.modeStack = m.modeStack[:len(m.modeStack)-1]
			return nil
		}
		return err
	}
	m.logMode.HandleEvent(s, e)
	if m.inTarget {
		return m.mapMode.HandleEvent(s, e)
	} else {
		return m.handleEventInternal(s, e)
	}
}

// Draw implements the termui.Mode interface.
func (m *gameMode) Draw(s termui.TerminalDriver) {
	// Draw the root window elements
	termui.DrawClear(s)
	sw, sh := s.Size()
	m.logMode.Bounds = util.NewRectXYWH(sw-38, 21, 38, sh-21)
	m.logMode.Draw(s)
	m.mapMode.Bounds = util.NewRectXYWH(0, 0, sw-39, sh)
	if m.inTarget {
		m.mapMode.CursorStyle = 2
		m.mapMode.DrawInfo = true
	} else {
		m.mapMode.Center = m.CityMap.Player.Position
		m.mapMode.CursorStyle = 0
		m.mapMode.DrawInfo = false
	}
	m.mapMode.DrawPaths = m.debug
	m.mapMode.Draw(s)
	termui.DrawVLine(s, util.NewPoint(sw-39, 0), sh, termui.CurrentTheme.Normal)
	m.minimap.Bounds = util.NewRectXYWH(sw-22, 0, 21, 21)
	m.minimap.Center = util.NewPoint(m.CityMap.Player.Position.X/game.ChunkWidth, m.CityMap.Player.Position.Y/game.ChunkHeight)
	m.minimap.Draw(s)
	m.status.Position = util.NewPoint(sw-38, 0)
	m.status.Draw(s)
	// Render the mode stack
	for _, m := range m.modeStack {
		m.Draw(s)
	}
}
