package termui

import (
	"github.com/gdamore/tcell/v2"
)

// List implements a potentially scrolled vertical list of strings for selection
// The callback function is called on selection.
func List(b tcell.Screen, x, y, w, h int, list []string, state *int,
	e tcell.Event, fn func(n int)) {
	// State sanitation
	if *state < 0 {
		*state = 0
	}
	if *state >= len(list) {
		*state = len(list) - 1
	}
	// State change in response to input
	switch ev := e.(type) {
	case *tcell.EventKey:
		switch ev.Rune() {
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
			fn(*state)
		default:
			switch ev.Key() {
			case tcell.KeyCR:
				fn(*state)
			}
		}
	}
	// Drawing
	DrawFill(b, x, y, w, h, ' ', CurrentTheme.Normal)
	si := 0
	if *state > h/2 {
		si += *state - h/2
	}
	if si+h > len(list) {
		si = len(list) - h
	}
	if si < 0 {
		si = 0
	}
	for i := si; i < si+h && i < len(list); i++ {
		s := list[i]
		if i == *state {
			DrawFill(b, x, y+i-si, w, 1, ' ', CurrentTheme.Highlight)
			DrawStringLeft(b, x, y+i-si, w, s, CurrentTheme.Highlight)
		} else {
			DrawStringLeft(b, x, y+i-si, w, s, CurrentTheme.Normal)
		}
	}
}

// BoxList is like List but with a surrounding title box.
func BoxList(b tcell.Screen, x, y, w, h int, title string, list []string,
	state *int, e tcell.Event, fn func(n int)) {
	DrawBox(b, x, y, w, h, CurrentTheme.Normal)
	DrawStringCenter(b, x, y, w, title, CurrentTheme.Normal)
	List(b, x+1, y+1, w-2, h-2, list, state, e, fn)
}
