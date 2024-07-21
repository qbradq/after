package termgui

import (
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// helpDialog implements a dialog displaying the game controls.
type helpDialog struct {
	page *game.HelpPage
	td   *termui.TextDialog // Text dialog used to display the help text
}

// newHelpDialog constructs a new helpDialog object ready for use.
func newHelpDialog() *helpDialog {
	return &helpDialog{
		page: game.HelpPages["Base/controls.txt"],
		td: &termui.TextDialog{
			Bounds: util.NewRectWH(80, 24),
			Boxed:  true,
			Title:  "Help",
		},
	}
}

// HandleEvent implements the termui.Mode interface.
func (m *helpDialog) HandleEvent(s termui.TerminalDriver, e any) error {
	switch ev := e.(type) {
	case *termui.EventKey:
		switch ev.Key {
		case '\033':
			return termui.ErrorQuit
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return m.td.HandleEvent(s, e)
}

// Draw implements the termui.Mode interface.
func (m *helpDialog) Draw(s termui.TerminalDriver) {
	if s == nil || m.page == nil {
		return
	}
	w, h := s.Size()
	m.td.Bounds = util.NewRectWH(w, h)
	m.td.Title = m.page.Title
	m.td.SetLines(m.page.Contents)
	m.td.Draw(s)
}
