package termgui

import (
	"sort"

	"github.com/qbradq/after/internal/citygen"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// scenarioList implements a list the player uses to select which starting
// scenario to use.
type scenarioList struct {
	Selected func(string)    // Function called when a scenario name is selected by the player
	list     *termui.List    // Interactive list of scenario names
	tb       *termui.TextBox // Text box for the description text
}

// newScenarioList constructs a new scenarioList for use.
func newScenarioList() *scenarioList {
	var ret *scenarioList
	ret = &scenarioList{
		list: &termui.List{
			Boxed: true,
			Title: "Choose Scenario",
			Selected: func(td termui.TerminalDriver, i int) error {
				ret.Selected(ret.list.Items[i])
				return termui.ErrorQuit
			},
		},
		tb: &termui.TextBox{
			Boxed: true,
			Title: "Choose Scenario",
		},
	}
	for k := range citygen.Scenarios {
		ret.list.Items = append(ret.list.Items, k)
	}
	sort.Strings(ret.list.Items)
	return ret
}

// HandleEvent implements the termui.Mode interface.
func (m *scenarioList) HandleEvent(s termui.TerminalDriver, e any) error {
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
func (m *scenarioList) Draw(s termui.TerminalDriver) {
	sc := citygen.Scenarios[m.list.CurrentSelection()]
	w, h := s.Size()
	sb := util.NewRectWH(w, h)
	mb := sb.CenterRect(42, 23)
	lb := mb
	lb.BR.Y -= 11
	tb := mb
	tb.TL.Y += 11
	m.tb.SetText(sc.Description)
	m.list.Bounds = lb
	m.list.Draw(s)
	m.tb.Bounds = tb
	m.tb.Draw(s)
}
