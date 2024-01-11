package termgui

import (
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// EscapeMenu implements the system menu that appears when you press escape.
type EscapeMenu struct {
	m    *GameMode   // Game mode back reference
	list termui.List // Menu list
}

// NewEscapeMenu returns a new EscapeMenu ready for use.
func NewEscapeMenu(m *GameMode) *EscapeMenu {
	ret := &EscapeMenu{
		m: m,
		list: termui.List{
			Boxed: true,
			Items: []string{"Resume", "Force Save", "Save and Quit", "_hbar_"},
			Title: "Game Menu",
			Selected: func(td termui.TerminalDriver, i int) error {
				switch i {
				case 0:
					return termui.ErrorQuit
				case 1:
					m.CityMap.FullSave()
					return termui.ErrorQuit
				case 2:
					m.CityMap.FullSave()
					m.Quit = true
					return termui.ErrorQuit
				case 4:
					m.Debug = !m.Debug
					return termui.ErrorQuit
				}
				return nil
			},
		},
	}
	if m.Debug {
		ret.list.Items = append(ret.list.Items, "Disable Debug Display")
	} else {
		ret.list.Items = append(ret.list.Items, "Enable Debug Display")
	}
	return ret
}

// HandleEvent implements the termui.Mode interface.
func (m *EscapeMenu) HandleEvent(s termui.TerminalDriver, e any) error {
	switch ev := e.(type) {
	case *termui.EventKey:
		if ev.Key == '\033' {
			return termui.ErrorQuit
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return m.list.HandleInput(s, e)
}

// Draw implements the termui.Mode interface.
func (m *EscapeMenu) Draw(s termui.TerminalDriver) {
	w, h := s.Size()
	m.list.Bounds = util.NewRectWH(w, h).CenterRect(27, 7)
	m.list.Draw(s)
}
