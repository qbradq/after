package termui

import (
	"github.com/gdamore/tcell/v2"
)

// DrawClear clears the screen.
func DrawClear(s tcell.Screen) {
	w, h := s.Size()
	DrawFill(s, 0, 0, w, h, ' ', CurrentTheme.Normal)
}

// DrawFill fills a region of the screen.
func DrawFill(s tcell.Screen, x, y, w, h int, r rune, style tcell.Style) {
	if x < 0 {
		x += w
		x = 0
	}
	if y < 0 {
		y += h
		y = 0
	}
	for iy := 0; iy < h; iy++ {
		for ix := 0; ix < w; ix++ {
			s.SetContent(x+ix, y+iy, r, nil, style)
		}
	}
}

// DrawRune draws a single rune.
func DrawRune(s tcell.Screen, x, y int, r rune, style tcell.Style) {
	s.SetContent(x, y, r, nil, style)
}

// DrawStringLeft draws a string left-justified.
func DrawStringLeft(s tcell.Screen, x, y, w int, t string, style tcell.Style) {
	for i, r := range t {
		if i >= w {
			break
		}
		s.SetContent(x+i, y, r, nil, style)
	}
}

// DrawStringRight draws a string right-justified.
func DrawStringRight(b tcell.Screen, x, y, w int, s string, style tcell.Style) {
	sx := x + (w - len(s))
	si := 0
	if sx < x {
		si += x - sx
		sx = x
	}
	ex := x + w
	for _, r := range s {
		if sx > ex || si >= len(s) {
			break
		}
		b.SetContent(sx, y, r, nil, style)
		sx++
		si++
	}
}

// DrawStringCenter draws a string centered.
func DrawStringCenter(b tcell.Screen, x, y, w int, s string, style tcell.Style) {
	sx := (x + (w / 2)) - (len(s) / 2)
	si := 0
	if sx < x {
		si += x - sx
		sx = x
	}
	ex := x + w
	for _, r := range s {
		if sx > ex || si >= len(s) {
			break
		}
		b.SetContent(sx, y, r, nil, style)
		sx++
		si++
	}
}

// DrawHLine draws a horizontal line.
func DrawHLine(b tcell.Screen, x, y, w int, style tcell.Style) {
	for i := 0; i < w; i++ {
		b.SetContent(x+i, y, '═', nil, style)
	}
}

// DrawVLine draws a vertical line.
func DrawVLine(b tcell.Screen, x, y, h int, style tcell.Style) {
	for i := 0; i < h; i++ {
		b.SetContent(x, y+i, '║', nil, style)
	}
}

// DrawBox draws a box.
func DrawBox(b tcell.Screen, x, y, w, h int, style tcell.Style) {
	DrawHLine(b, x+1, y, w-2, style)
	DrawHLine(b, x+1, y+h-1, w-2, style)
	DrawVLine(b, x, y+1, h-2, style)
	DrawVLine(b, x+w-1, y+1, h-2, style)
	b.SetContent(x, y, '╔', nil, style)
	b.SetContent(x+w-1, y, '╗', nil, style)
	b.SetContent(x, y+h-1, '╚', nil, style)
	b.SetContent(x+w-1, y+h-1, '╝', nil, style)
}
