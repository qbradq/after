package termui

import (
	"strings"

	"github.com/mitchellh/go-wordwrap"
	"github.com/qbradq/after/lib/util"
)

// TextBox implements a word-wrapped text box that displays a single string.
type TextBox struct {
	Bounds util.Rect // Fixed bounds of the text box display, line length is inferred from this
	Boxed  bool      // If true a box will be rendered around the text box
	Title  string    // Title to display if any, only valid if Boxed is true
	lines  []string  // Cache of lines to display
}

// SetText sets the text that the text box will display and returns the number
// of lines the string was broken into.
func (m *TextBox) SetText(s string) int {
	w := m.Bounds.Width()
	if m.Boxed {
		w -= 2
	}
	m.lines = strings.Split(wordwrap.WrapString(s, uint(w)), "\n")
	return len(m.lines)
}

// HandleEvent implements the Mode interface.
func (m *TextBox) HandleEvent(s TerminalDriver, e any) error {
	return nil
}

// Draw implements the Mode interface.
func (m *TextBox) Draw(s TerminalDriver) {
	b := m.Bounds
	if m.Boxed {
		DrawBox(s, b, CurrentTheme.Normal)
		DrawStringCenter(s, b, m.Title, CurrentTheme.Normal)
		b.TL.X++
		b.TL.Y++
		b.BR.X--
		b.BR.Y--
	}
	// Drawing
	DrawFill(s, b, Glyph{Rune: ' ', Style: CurrentTheme.Normal})
	for _, l := range m.lines {
		if b.TL.Y > b.BR.Y {
			break
		}
		DrawStringLeft(s, b, l, CurrentTheme.Normal)
		b.TL.Y++
	}
}
