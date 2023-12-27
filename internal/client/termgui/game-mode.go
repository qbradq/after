package termgui

import (
	"errors"

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
	ModeStack []termui.Mode // Internal stack of mode that overlay the main game mode, like the escape menu or inventory screen
	Quit      bool          // If true we should quit
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
	}
	game.Log = gm.LogMode
	game.Log.Log(termui.ColorTeal, "Welcome to the Aftermath!")
	return gm
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
	switch ev := e.(type) {
	case *termui.EventKey:
		switch ev.Key {
		case 'u':
			m.CityMap.Player.Position.X++
			m.CityMap.Player.Position.Y--
		case 'y':
			m.CityMap.Player.Position.X--
			m.CityMap.Player.Position.Y--
		case 'n':
			m.CityMap.Player.Position.X++
			m.CityMap.Player.Position.Y++
		case 'b':
			m.CityMap.Player.Position.X--
			m.CityMap.Player.Position.Y++
		case 'l':
			m.CityMap.Player.Position.X++
		case 'h':
			m.CityMap.Player.Position.X--
		case 'j':
			m.CityMap.Player.Position.Y++
		case 'k':
			m.CityMap.Player.Position.Y--
		case 'm':
			termui.RunMode(s, &Minimap{
				CityMap:     m.CityMap,
				Bounds:      util.NewRectWH(s.Size()),
				Center:      util.NewPoint(m.CityMap.Player.Position.X/game.ChunkWidth, m.CityMap.Player.Position.Y/game.ChunkHeight),
				CursorStyle: 2,
				DrawInfo:    true,
			})
		case '\033':
			m.ModeStack = append(m.ModeStack, NewEscapeMenu(m))
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	m.LogMode.HandleEvent(s, e)
	return m.MapMode.HandleEvent(s, e)
}

// Draw implements the termui.Mode interface.
func (m *GameMode) Draw(s termui.TerminalDriver) {
	// Draw the root window elements
	termui.DrawClear(s)
	sw, sh := s.Size()
	m.LogMode.Bounds = util.NewRectXYWH(sw-38, 21, 38, sh-21)
	m.LogMode.Draw(s)
	m.MapMode.Bounds = util.NewRectXYWH(0, 0, sw-39, sh)
	m.MapMode.Center = m.CityMap.Player.Position
	m.MapMode.CursorStyle = 0
	m.MapMode.Draw(s)
	termui.DrawVLine(s, util.NewPoint(sw-39, 0), sh, termui.CurrentTheme.Normal)
	m.Minimap.Bounds = util.NewRectXYWH(sw-22, 0, 21, 21)
	m.Minimap.Center = util.NewPoint(m.CityMap.Player.Position.X/game.ChunkWidth, m.CityMap.Player.Position.Y/game.ChunkHeight)
	m.Minimap.Draw(s)
	termui.DrawBox(s, util.NewRectXYWH(sw-38, 0, 16, 21), termui.CurrentTheme.Normal)
	termui.DrawStringCenter(s, util.NewRectXYWH(sw-38, 0, 16, 1), "Placeholder", termui.CurrentTheme.Normal)
	// Render the mode stack
	for _, m := range m.ModeStack {
		m.Draw(s)
	}
}
