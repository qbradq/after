package termui

import (
	"github.com/qbradq/after/lib/util"
)

// List implements a mode that presents a scrollable list to the user.
type List struct {
	Bounds     util.Rect                       // Bounds of the rect on screen
	CursorPos  int                             // Current cursor position
	Boxed      bool                            // If true a box is drawn around the list
	Title      string                          // Title for the box, if any
	Items      []string                        // Items of the list
	Selected   func(TerminalDriver, int) error // The function that is called if the user selects an item
	HideCursor bool                            // If true we will not highlight the cursor position
}

// CurrentSelection returns the string under the current cursor position.
func (m *List) CurrentSelection() string {
	return m.Items[m.CursorPos]
}

// HandleEvent implements the Mode interface.
func (m *List) HandleEvent(s TerminalDriver, e any) error {
	// State sanitation
	if len(m.Items) < 1 {
		return ErrorQuit
	}
	if m.CursorPos < 0 {
		m.CursorPos = 0
	}
	if m.CursorPos >= len(m.Items) {
		m.CursorPos = len(m.Items) - 1
	}
	// State change in response to input
	switch ev := e.(type) {
	case *EventKey:
		switch ev.Key {
		case 'k':
			for {
				m.CursorPos--
				if m.CursorPos < 0 {
					m.CursorPos += len(m.Items)
				}
				if m.Items[m.CursorPos] != "_hbar_" {
					break
				}
			}
		case 'j':
			for {
				m.CursorPos++
				if m.CursorPos >= len(m.Items) {
					m.CursorPos -= len(m.Items)
				}
				if m.Items[m.CursorPos] != "_hbar_" {
					break
				}
			}
		case ' ':
			fallthrough
		case '\n':
			return m.Selected(s, m.CursorPos)
		case '\033':
			return ErrorQuit
		}
	}
	return nil
}

// Draw implements the Mode interface.
func (m *List) Draw(s TerminalDriver) {
	// Sanitation
	if m.CursorPos < 0 {
		m.CursorPos = 0
	}
	if m.CursorPos >= len(m.Items) {
		m.CursorPos = len(m.Items) - 1
	}
	// Box drawing
	b := m.Bounds
	if m.Boxed {
		DrawBox(s, b, CurrentTheme.Normal)
		DrawStringCenter(s, b, m.Title, CurrentTheme.Normal)
		b.TL.X++
		b.TL.Y++
		b.BR.X--
		b.BR.Y--
	}
	// List drawing
	DrawFill(s, b, Glyph{Rune: ' ', Style: CurrentTheme.Normal})
	si := 0
	if m.CursorPos > b.Height()/2 {
		si += m.CursorPos - b.Height()/2
	}
	if si+b.Height() > len(m.Items) {
		si = len(m.Items) - b.Height()
	}
	if si < 0 {
		si = 0
	}
	for i := si; i < si+b.Height() && i < len(m.Items); i++ {
		text := m.Items[i]
		if i == m.CursorPos && !m.HideCursor {
			DrawFill(s, util.NewRectXYWH(b.TL.X, b.TL.Y+i-si, b.Width(), 1), Glyph{
				Rune:  ' ',
				Style: CurrentTheme.Highlight,
			})
			if text == "_hbar_" {
				DrawHLine(s, util.NewPoint(b.TL.X, b.TL.Y+i-si), b.Width(),
					CurrentTheme.Highlight)
			} else {
				DrawStringLeft(s, util.NewRectXYWH(b.TL.X, b.TL.Y+i-si, b.Width(), 1),
					text, CurrentTheme.Highlight)
			}
		} else {
			if text == "_hbar_" {
				DrawHLine(s, util.NewPoint(b.TL.X, b.TL.Y+i-si), b.Width(),
					CurrentTheme.Normal)
			} else {
				DrawStringLeft(s, util.NewRectXYWH(b.TL.X, b.TL.Y+i-si, b.Width(), 1),
					text, CurrentTheme.Normal)
			}
		}
	}
}
