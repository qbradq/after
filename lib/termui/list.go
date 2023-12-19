package termui

import (
	"github.com/qbradq/after/lib/util"
)

// List implements a potentially scrolled vertical list of strings for selection
// The callback function is called on selection.
func List(s TerminalDriver, b util.Rect, list []string, state *int,
	e any, fn func(n int)) {
	// State sanitation
	if *state < 0 {
		*state = 0
	}
	if *state >= len(list) {
		*state = len(list) - 1
	}
	// State change in response to input
	switch ev := e.(type) {
	case *EventKey:
		switch ev.Key {
		case 'k':
			*state--
			if *state < 0 {
				*state += len(list)
			}
		case 'j':
			*state++
			if *state >= len(list) {
				*state -= len(list)
			}
		case ' ':
			fallthrough
		case '\n':
			fn(*state)
		}
	}
	// Drawing
	DrawFill(s, b, Glyph{Rune: ' ', Style: CurrentTheme.Normal})
	si := 0
	if *state > b.Height()/2 {
		si += *state - b.Height()/2
	}
	if si+b.Height() > len(list) {
		si = len(list) - b.Height()
	}
	if si < 0 {
		si = 0
	}
	for i := si; i < si+b.Height() && i < len(list); i++ {
		text := list[i]
		if i == *state {
			DrawFill(s, util.NewRectXYWH(b.TL.X, b.TL.Y+i-si, b.Width(), 1), Glyph{
				Rune:  ' ',
				Style: CurrentTheme.Highlight,
			})
			DrawStringLeft(s, util.NewRectXYWH(b.TL.X, b.TL.Y+i-si, b.Width(), 1),
				text, CurrentTheme.Highlight)
		} else {
			DrawStringLeft(s, util.NewRectXYWH(b.TL.X, b.TL.Y+i-si, b.Width(), 1),
				text, CurrentTheme.Normal)
		}
	}
}

// BoxList is like List but with a surrounding title box.
func BoxList(s TerminalDriver, b util.Rect, title string, list []string,
	state *int, e any, fn func(n int)) {
	DrawBox(s, b, CurrentTheme.Normal)
	DrawStringCenter(s, b, title, CurrentTheme.Normal)
	b.TL.X++
	b.TL.Y++
	b.BR.X--
	b.BR.Y--
	List(s, b, list, state, e, fn)
}
