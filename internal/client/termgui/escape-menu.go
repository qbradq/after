package termgui

import (
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// escapeMenu implements the system menu that appears when you press escape.
type escapeMenu struct {
	m    *gameMode   // Game mode back reference
	list termui.List // Menu list
}

// newEscapeMenu returns a new EscapeMenu ready for use.
func newEscapeMenu(m *gameMode) *escapeMenu {
	ret := &escapeMenu{
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
					m.quit = true
					return termui.ErrorQuit
				case 4:
					m.debug = !m.debug
					return termui.ErrorQuit
				}
				return nil
			},
		},
	}
	if m.debug {
		ret.list.Items = append(ret.list.Items, "Disable Debug Display")
	} else {
		ret.list.Items = append(ret.list.Items, "Enable Debug Display")
	}
	return ret
}

// HandleEvent implements the termui.Mode interface.
func (m *escapeMenu) HandleEvent(s termui.TerminalDriver, e any) error {
	switch ev := e.(type) {
	case *termui.EventKey:
		if ev.Key == '\033' {
			return termui.ErrorQuit
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return m.list.HandleEvent(s, e)
}

// Draw implements the termui.Mode interface.
func (m *escapeMenu) Draw(s termui.TerminalDriver) {
	w, h := s.Size()
	m.list.Bounds = util.NewRectWH(w, h).CenterRect(27, 7)
	m.list.Draw(s)
}
