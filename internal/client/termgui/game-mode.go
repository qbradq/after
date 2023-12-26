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
	MapMode   *MapMode      // Map display
	Minimap   *Minimap      // Small mini-map
	ModeStack []termui.Mode // Internal stack of mode that overlay the main game mode, like the escape menu or inventory screen
	Quit      bool          // If true we should quit
}

// NewGameMode returns a new game mode.
func NewGameMode(m *game.CityMap, c util.Point) *GameMode {
	return &GameMode{
		CityMap: m,
		MapMode: &MapMode{
			CityMap: m,
			Center:  c,
		},
		Minimap: &Minimap{
			CityMap:     m,
			CursorStyle: 1,
		},
	}
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
		if ev.Key == '\033' {
			m.ModeStack = append(m.ModeStack, NewEscapeMenu(m))
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return m.MapMode.HandleEvent(s, e)
}

// Draw implements the termui.Mode interface.
func (m *GameMode) Draw(s termui.TerminalDriver) {
	// Draw the root window elements
	termui.DrawClear(s)
	sw, sh := s.Size()
	m.MapMode.Bounds = util.NewRectXYWH(0, 0, sw-39, sh)
	m.MapMode.CursorStyle = 0
	m.MapMode.Draw(s)
	termui.DrawVLine(s, util.NewPoint(sw-39, 0), sh, termui.CurrentTheme.Normal)
	m.Minimap.Bounds = util.NewRectXYWH(sw-22, 0, 21, 21)
	m.Minimap.Center = util.NewPoint(m.MapMode.Center.X/game.ChunkWidth, m.MapMode.Center.Y/game.ChunkHeight)
	m.Minimap.Draw(s)
	// termui.DrawBox(s, util.NewRectXYWH(sw-38, 0, 16, 21), termui.CurrentTheme.Normal)
	// Render the mode stack
	for _, m := range m.ModeStack {
		m.Draw(s)
	}
}
