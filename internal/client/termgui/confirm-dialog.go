package termgui

import (
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// confirmDialog implements a pop-up style dialog to prompt the user for a yes
// or no answer.
type confirmDialog struct {
	Title     string          // Title of the dialog
	Prompt    string          // The prompt string to present
	Confirmed func()          // Function to call when the player responds to the dialog in the affirmative
	tb        *termui.TextBox // Text box used to display the message
}

// newConfirmDialog constructs a new confirmDialog object ready for use.
func newConfirmDialog() *confirmDialog {
	return &confirmDialog{
		tb: &termui.TextBox{
			Boxed: true,
		},
	}
}

// HandleEvent implements the termui.Mode interface.
func (m *confirmDialog) HandleEvent(s termui.TerminalDriver, e any) error {
	switch ev := e.(type) {
	case *termui.EventKey:
		switch ev.Key {
		case 'y':
			fallthrough
		case 'Y':
			m.Confirmed()
			return termui.ErrorQuit
		case 'n':
			fallthrough
		case 'N':
			fallthrough
		case ' ':
			fallthrough
		case '\n':
			fallthrough
		case '\033':
			return termui.ErrorQuit
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return nil
}

// Draw implements the termui.Mode interface.
func (m *confirmDialog) Draw(s termui.TerminalDriver) {
	w, h := s.Size()
	sb := util.NewRectWH(w, h)
	th := m.tb.SetText(m.Prompt + " [y/N]")
	tb := sb.CenterRect(42, th+2)
	m.tb.Bounds = tb
	m.tb.Title = m.Title
	m.tb.Draw(s)
}
