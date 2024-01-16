package termgui

import (
	"fmt"
	"strings"

	"github.com/mitchellh/go-wordwrap"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

const (
	maxLogLines int = 1024
)

type logLine struct {
	c termui.Color
	s string
}

// LogMode implements a log window.
type LogMode struct {
	Bounds util.Rect // Screen bounds of the log display
	lines  []logLine // Contents of the log
	lso    int       // Line scroll offset
}

// Log adds a line to the log.
func (m *LogMode) Log(c termui.Color, s string, args ...any) {
	s = fmt.Sprintf(s, args...)
	m.lines = append(m.lines, logLine{
		c: c,
		s: s,
	})
	if len(m.lines) > maxLogLines {
		m.lines = m.lines[len(m.lines)-maxLogLines:]
	}
}

// HandleEvent implements the termui.Mode interface.
func (m *LogMode) HandleEvent(s termui.TerminalDriver, e any) error {
	switch e.(type) {
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return nil
}

// Draw implements the termui.Mode interface.
func (m *LogMode) Draw(s termui.TerminalDriver) {
	ty := m.Bounds.BR.Y
	for i := (len(m.lines) - 1) - m.lso; i >= 0 && ty >= m.Bounds.TL.Y; i-- {
		l := m.lines[i]
		lines := strings.Split(wordwrap.WrapString(l.s, uint(m.Bounds.Width())), "\n")
		for j := len(lines) - 1; j >= 0 && ty >= m.Bounds.TL.Y; j-- {
			termui.DrawStringLeft(
				s,
				util.NewRectXYWH(m.Bounds.TL.X, ty, m.Bounds.Width(), 1),
				lines[j],
				termui.CurrentTheme.Normal.Foreground(l.c),
			)
			ty--
		}
	}
}
