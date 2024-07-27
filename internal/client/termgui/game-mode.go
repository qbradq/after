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
	escapeMenu    *escapeMenu    // Escape menu
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
		confirmDialog: newConfirmDialog(),
	}
	gm.escapeMenu = newEscapeMenu(gm)
	game.Log = gm.logMode
	game.Log.Log(termui.ColorTeal, "Welcome to the aftermath!")
	return gm
}

func (m *gameMode) handleEventInternal(s termui.TerminalDriver, e any) error {
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
			m.CityMap.PlayerTookTurn(time.Second, func() { m.Draw(s) })
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
						m.CityMap.PlayerTookTurn(time.Duration(float64(time.Second)*m.CityMap.Player.ActSpeed()), func() { m.Draw(s) })
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
					m.CityMap.Player.Attack(a, m.CityMap.Now)
					m.CityMap.PlayerTookTurn(time.Second, func() { m.Draw(s) })
				}
				return nil
			}
			m.mapMode.Center = m.CityMap.Player.Position
			m.mapMode.CursorPos = m.CityMap.Player.Position
			m.mapMode.CursorRange = 1
			return nil
		case 'd': // Drop item from inventory
			m.inTarget = true
			m.mapMode.Callback = func(p util.Point, confirmed bool) error {
				m.inTarget = false
				if !confirmed {
					return nil
				}
				inv := newInventoryDialog(m.CityMap, &m.CityMap.Player.Actor, p)
				m.modeStack = append(m.modeStack, inv)
				return nil
			}
			m.mapMode.Center = m.CityMap.Player.Position
			m.mapMode.CursorPos = m.CityMap.Player.Position
			m.mapMode.CursorRange = 1
			m.logMode.Log(termui.ColorPurple, "Drop where?")
			return nil
		case ',': // get items at feet (,)
			inv := newInventoryDialog(
				m.CityMap,
				&m.CityMap.Player.Actor,
				m.CityMap.Player.Position)
			inv.OnRight = true
			return nil
		case 'g': // Get items within reach
			m.inTarget = true
			m.mapMode.Callback = func(p util.Point, confirmed bool) error {
				m.inTarget = false
				if !confirmed {
					return nil
				}
				inv := newInventoryDialog(m.CityMap, &m.CityMap.Player.Actor, p)
				inv.OnRight = true
				m.modeStack = append(m.modeStack, inv)
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
		case 'i': // Inventory
			inv := newInventoryDialog(
				m.CityMap,
				&m.CityMap.Player.Actor,
				m.CityMap.Player.Position)
			m.modeStack = append(m.modeStack, inv)
			return nil
		case 'r': // Rest / Wait
			td := newTimeDialog(m.CityMap)
			td.Title = "Rest How Long?"
			td.Selected = func(d time.Duration) {
				m.CityMap.PlayerTookTurn(d, func() { m.Draw(s) })
			}
			m.modeStack = append(m.modeStack, td)
			m.logMode.Log(termui.ColorPurple, "Rest how long?")
			return nil
		case 'R': // Run / Walk toggle
			if m.CityMap.Player.Running {
				m.logMode.Log(termui.ColorFuchsia, "You slow to a walk.")
			} else {
				m.logMode.Log(termui.ColorFuchsia, "You quicken your pace to a run.")
			}
			m.CityMap.Player.Running = !m.CityMap.Player.Running
			return nil
		case '\033': // Escape menu
			m.modeStack = append(m.modeStack, newEscapeMenu(m))
			return nil
		case '?': // Help menu
			m.modeStack = append(m.modeStack, newHelpDialog())
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
			if err := m.handleBump(dir, s); err != nil {
				return err
			}
		} else {
			s.FlushEvents()
		}
	}
	return nil
}

// handleBump handles the player bumping into something, returning any error.
func (m *gameMode) handleBump(dir util.Direction, s termui.TerminalDriver) error {
	np := m.CityMap.Player.Position.Step(dir)
	// Try attacking first
	a := m.CityMap.ActorAt(np)
	if a != nil {
		m.CityMap.Player.Attack(a, m.CityMap.Now)
		m.CityMap.PlayerTookTurn(time.Duration(float64(time.Second)*m.CityMap.Player.ActSpeed()), func() { m.Draw(s) })
		s.FlushEvents()
		return nil
	}
	// Then try climbing
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
	// Try to use fixed items
	items := m.CityMap.ItemsAt(np)
	for _, i := range items {
		if !i.Fixed || i.Events == nil {
			continue
		}
		if _, found := i.Events["Use"]; !found {
			continue
		}
		err, used := events.ExecuteItemUseEvent("Use", items[len(items)-1], &m.CityMap.Player.Actor, m.CityMap)
		if err != nil {
			return err
		}
		if used {
			m.CityMap.PlayerTookTurn(time.Duration(float64(time.Second)*m.CityMap.Player.ActSpeed()), func() { m.Draw(s) })
		}
		s.FlushEvents()
		// If we reached this point we have successfully used a fixed item.
		break
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
	// Log area
	m.logMode.Bounds = util.NewRectXYWH(sw-38, 23, 38, sh-23)
	m.logMode.Draw(s)
	// Map display
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
	// Mini-map
	mmb := util.NewRectXYWH(sw-23, 0, 23, 23)
	termui.DrawBox(s, mmb, termui.CurrentTheme.Normal)
	m.minimap.Bounds = mmb.Shrink(1)
	m.minimap.Center = util.NewPoint(m.CityMap.Player.Position.X/game.ChunkWidth, m.CityMap.Player.Position.Y/game.ChunkHeight)
	m.minimap.Draw(s)
	// Status display
	m.status.Position = util.NewPoint(sw-39, 0)
	m.status.Draw(s)
	// Render the mode stack
	for _, m := range m.modeStack {
		m.Draw(s)
	}
}
