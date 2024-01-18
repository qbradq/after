package termgui

import (
	"errors"
	"time"

	"github.com/qbradq/after/internal/events"
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// GameMode implements the top-level client interface.
type GameMode struct {
	CityMap   *game.CityMap // The city we are playing
	LogMode   *LogMode      // Log display
	MapMode   *MapMode      // Map display
	Minimap   *Minimap      // Small mini-map
	Status    *StatusPanel  // Status panel
	ModeStack []termui.Mode // Internal stack of mode that overlay the main game mode, like the escape menu or inventory screen
	Quit      bool          // If true we should quit
	InTarget  bool          // If true we are in targeting mode
	Debug     bool          // If true display debug information
}

// NewGameMode returns a new game mode.
func NewGameMode(m *game.CityMap) *GameMode {
	gm := &GameMode{
		CityMap: m,
		LogMode: &LogMode{},
		MapMode: &MapMode{
			CityMap: m,
			Center:  m.Player.Position,
		},
		Minimap: &Minimap{
			CityMap:     m,
			CursorStyle: 1,
		},
		Status: &StatusPanel{
			CityMap: m,
		},
	}
	game.Log = gm.LogMode
	game.Log.Log(termui.ColorTeal, "Welcome to the Aftermath!")
	return gm
}

func (m *GameMode) handleEventInternal(s termui.TerminalDriver, e any) error {
	dir := util.DirectionInvalid
	switch ev := e.(type) {
	case *termui.EventKey:
		switch ev.Key {
		case 'u':
			dir = util.DirectionNorthEast
		case 'y':
			dir = util.DirectionNorthWest
		case 'n':
			dir = util.DirectionSouthEast
		case 'b':
			dir = util.DirectionSouthWest
		case 'l':
			dir = util.DirectionEast
		case 'h':
			dir = util.DirectionWest
		case 'j':
			dir = util.DirectionSouth
		case 'k':
			dir = util.DirectionNorth
		case '.':
			m.CityMap.PlayerTookTurn(time.Second)
			s.FlushEvents()
			return nil
		case 'x':
			m.InTarget = true
			m.MapMode.Callback = func(p util.Point, b bool) error {
				m.InTarget = false
				return nil
			}
			m.MapMode.Center = m.CityMap.Player.Position
			m.MapMode.CursorPos = m.CityMap.Player.Position
			m.MapMode.CursorRange = 0
			return nil
		case 'm':
			termui.RunMode(s, &Minimap{
				CityMap:     m.CityMap,
				Bounds:      util.NewRectWH(s.Size()),
				Center:      util.NewPoint(m.CityMap.Player.Position.X/game.ChunkWidth, m.CityMap.Player.Position.Y/game.ChunkHeight),
				CursorStyle: 2,
				DrawInfo:    true,
			})
			return nil
		case 'U':
			m.InTarget = true
			m.MapMode.Callback = func(p util.Point, b bool) error {
				m.InTarget = false
				if !b {
					return nil
				}
				items := m.CityMap.ItemsAt(p)
				if len(items) > 0 {
					err := events.ExecuteItemUseEvent("Use", items[len(items)-1], &m.CityMap.Player.Actor, m.CityMap)
					if err != nil {
						return err
					}
					m.CityMap.PlayerTookTurn(time.Second)
				}
				return nil
			}
			m.MapMode.Center = m.CityMap.Player.Position
			m.MapMode.CursorPos = m.CityMap.Player.Position
			m.MapMode.CursorRange = 1
			return nil
		case 'a':
			m.InTarget = true
			m.MapMode.Callback = func(p util.Point, b bool) error {
				m.InTarget = false
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
			m.MapMode.Center = m.CityMap.Player.Position
			m.MapMode.CursorPos = m.CityMap.Player.Position
			m.MapMode.CursorRange = 1
			return nil
		case '\033':
			m.ModeStack = append(m.ModeStack, NewEscapeMenu(m))
			return nil
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	dir = dir.Bound()
	if dir != util.DirectionInvalid {
		if !m.CityMap.StepPlayer(dir) {
			// Bump handling
			np := m.CityMap.Player.Position.Step(dir)
			a := m.CityMap.ActorAt(np)
			if a != nil {
				a.Damage(m.CityMap.Player.MinDamage, m.CityMap.Player.MaxDamage, m.CityMap.Now, &m.CityMap.Player.Actor)
				m.CityMap.PlayerTookTurn(time.Second)
				s.FlushEvents()
			} else {
				items := m.CityMap.ItemsAt(np)
				if len(items) > 0 {
					err := events.ExecuteItemUseEvent("Use", items[len(items)-1], &m.CityMap.Player.Actor, m.CityMap)
					if err != nil {
						return err
					}
					m.CityMap.PlayerTookTurn(time.Second)
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
func (m *GameMode) HandleEvent(s termui.TerminalDriver, e any) error {
	if len(m.ModeStack) > 0 {
		err := m.ModeStack[len(m.ModeStack)-1].HandleEvent(s, e)
		if m.Quit {
			return termui.ErrorQuit
		}
		if errors.Is(err, termui.ErrorQuit) {
			m.ModeStack = m.ModeStack[:len(m.ModeStack)-1]
			return nil
		}
		return err
	}
	m.LogMode.HandleEvent(s, e)
	if m.InTarget {
		return m.MapMode.HandleEvent(s, e)
	} else {
		return m.handleEventInternal(s, e)
	}
}

// Draw implements the termui.Mode interface.
func (m *GameMode) Draw(s termui.TerminalDriver) {
	// Draw the root window elements
	termui.DrawClear(s)
	sw, sh := s.Size()
	m.LogMode.Bounds = util.NewRectXYWH(sw-38, 21, 38, sh-21)
	m.LogMode.Draw(s)
	m.MapMode.Bounds = util.NewRectXYWH(0, 0, sw-39, sh)
	if m.InTarget {
		m.MapMode.CursorStyle = 2
		m.MapMode.DrawInfo = true
	} else {
		m.MapMode.Center = m.CityMap.Player.Position
		m.MapMode.CursorStyle = 0
		m.MapMode.DrawInfo = false
	}
	m.MapMode.DrawPaths = m.Debug
	m.MapMode.Draw(s)
	termui.DrawVLine(s, util.NewPoint(sw-39, 0), sh, termui.CurrentTheme.Normal)
	m.Minimap.Bounds = util.NewRectXYWH(sw-22, 0, 21, 21)
	m.Minimap.Center = util.NewPoint(m.CityMap.Player.Position.X/game.ChunkWidth, m.CityMap.Player.Position.Y/game.ChunkHeight)
	m.Minimap.Draw(s)
	m.Status.Position = util.NewPoint(sw-38, 0)
	m.Status.Draw(s)
	// Render the mode stack
	for _, m := range m.ModeStack {
		m.Draw(s)
	}
}
