package termui

import (
	"github.com/qbradq/after/lib/util"
)

// TextDialog implements a dialog that displays a multi-line string that may
// contain markup. No word wrapping is performed.
type TextDialog struct {
	Bounds  util.Rect // Fixed bounds of the text dialog
	Boxed   bool      // If true a box will be rendered around the text dialog
	Title   string    // Title to display if any, only valid if Boxed is true
	lines   []string  // Cache of lines to display
	vScroll int       // Number of lines scrolled vertically
	hScroll int       // Number of cells scrolled horizontally
}

// SetLines sets the lines that the text dialog will display.
func (m *TextDialog) SetLines(lines []string) {
	m.lines = lines
}

// HandleEvent implements the Mode interface.
func (m *TextDialog) HandleEvent(s TerminalDriver, e any) error {
	switch ev := e.(type) {
	case *EventKey:
		if ev.Key == '\033' {
			return ErrorQuit
		}
	case *EventQuit:
		return ErrorQuit
	}
	return nil
}

// Draw implements the Mode interface.
func (m *TextDialog) Draw(s TerminalDriver) {
	b := m.Bounds
	if m.Boxed {
		DrawBox(s, b, CurrentTheme.Normal)
		DrawStringCenter(s, b, m.Title, CurrentTheme.Normal)
		b.TL.X++
		b.TL.Y++
		b.BR.X--
		b.BR.Y--
		// Draw the progress bar
		ratio := float64(b.Height()) / float64(len(m.lines))
		top := int(float64(m.vScroll) * ratio)
		lastLine := m.vScroll + b.Height()
		if lastLine >= len(m.lines) {
			lastLine = len(m.lines) - 1
		}
		bottom := int(float64(lastLine) * ratio)
		DrawFill(s, util.NewRectXYWH(b.BR.X+1, b.TL.Y+top, 1, bottom-top), Glyph{
			Rune:  '+',
			Style: CurrentTheme.Normal,
		})
	}
	// Drawing
	DrawFill(s, b, Glyph{Rune: ' ', Style: CurrentTheme.Normal})
	// Draw lines withing vertical window
	for y := 0; y < b.Height(); y++ {
		iy := y + m.vScroll
		if iy >= len(m.lines) {
			break
		}
		// Escaped line drawing within horizontal window
		line := m.lines[iy]
		inEscape := false
		ns := CurrentTheme.Normal
		x := 0
		emit := true
		for _, c := range line {
			if inEscape {
				if c >= '0' && c <= 9 {
					ns = ns.Foreground(Color(c - '0'))
				} else if c >= 'A' && c <= 'F' {
					ns = ns.Foreground(Color((c - 'A') + 10))
				} else if c >= 'a' && c <= 'f' {
					ns = ns.Foreground(Color((c - 'a') + 10))
				} else if c == '%' {
					emit = true
				}
				inEscape = false
			} else {
				if c == '%' {
					inEscape = true
					emit = false
				} else if c >= ' ' {
					emit = true
				} else {
					emit = false
				}
			}
			if emit {
				if x >= m.hScroll {
					s.SetCell(util.NewPoint(x-m.hScroll+b.TL.X, y+b.TL.Y), Glyph{
						Rune:  c,
						Style: ns,
					})
				}
				x++
			}
		}
	}
}
