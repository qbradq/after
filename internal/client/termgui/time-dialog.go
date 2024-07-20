package termgui

import (
	"time"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// timeDialog implements a dialog to select an amount of time for waiting or
// sleeping.
type timeDialog struct {
	Selected func(time.Duration) // Callback function when the player selects a duration
	Title    string              // Title for the dialog
	list     *termui.List        // List of durations
}

// newTimeDialog creates a new timeDialog ready for use.
func newTimeDialog(m *game.CityMap) *timeDialog {
	durations := []time.Duration{
		time.Minute,
		time.Minute * 5,
		time.Minute * 15,
		time.Hour,
		time.Hour * 4,
		time.Hour * 8,
		time.Hour * 24,
	}
	yyyy, mm, dd := m.Now.Date()
	hh := m.Now.Hour()
	if hh < 6 {
		durations = append(durations, time.Date(yyyy, mm, dd, 6, 0, 0, 0, m.Now.Location()).Sub(m.Now))
	} else {
		durations = append(durations, time.Date(yyyy, mm, dd+1, 6, 0, 0, 0, m.Now.Location()).Sub(m.Now))
	}
	if hh < 12 {
		durations = append(durations, time.Date(yyyy, mm, dd, 12, 0, 0, 0, m.Now.Location()).Sub(m.Now))
	} else {
		durations = append(durations, time.Date(yyyy, mm, dd+1, 12, 0, 0, 0, m.Now.Location()).Sub(m.Now))
	}
	if hh < 18 {
		durations = append(durations, time.Date(yyyy, mm, dd, 18, 0, 0, 0, m.Now.Location()).Sub(m.Now))
	} else {
		durations = append(durations, time.Date(yyyy, mm, dd+1, 18, 0, 0, 0, m.Now.Location()).Sub(m.Now))
	}
	durations = append(durations, time.Date(yyyy, mm, dd+1, 0, 0, 0, 0, m.Now.Location()).Sub(m.Now))
	var ret *timeDialog
	ret = &timeDialog{
		list: &termui.List{
			Boxed: true,
			Items: []string{
				"1 Minute",
				"5 Minutes",
				"15 Minutes",
				"1 Hour",
				"4 Hours",
				"8 Hours",
				"24 Hours",
				"Until Dawn",
				"Until Noon",
				"Until Dark",
				"Until Midnight",
			},
			Selected: func(s termui.TerminalDriver, i int) error {
				ret.Selected(durations[i])
				return termui.ErrorQuit
			},
		},
	}
	return ret
}

// HandleEvent implements the termui.Mode interface.
func (m *timeDialog) HandleEvent(s termui.TerminalDriver, e any) error {
	if err := m.list.HandleEvent(s, e); err != nil {
		return err
	}
	switch e.(type) {
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return nil
}

// Draw implements the termui.Mode interface.
func (m *timeDialog) Draw(s termui.TerminalDriver) {
	sb := util.NewRectWH(s.Size())
	m.list.Bounds = sb.CenterRect(16, 13)
	m.list.Title = m.Title
	m.list.Draw(s)
}
