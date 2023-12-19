package termui

import (
	"github.com/qbradq/after/lib/util"
)

// DrawClear clears the screen.
func DrawClear(s TerminalDriver) {
	w, h := s.Size()
	DrawFill(s, util.NewRectWH(w, h), Glyph{Rune: ' ', Style: CurrentTheme.Normal})
}

// DrawFill fills a region of the screen.
func DrawFill(s TerminalDriver, b util.Rect, g Glyph) {
	var p util.Point
	for p.Y = b.TL.Y; p.Y <= b.BR.Y; p.Y++ {
		for p.X = b.TL.X; p.X <= b.BR.X; p.X++ {
			s.SetCell(p, g)
		}
	}
}

// DrawStringLeft draws a string left-justified.
func DrawStringLeft(s TerminalDriver, b util.Rect, t string, style Style) {
	p := b.TL
	for i, r := range t {
		if i >= b.Width() {
			break
		}
		s.SetCell(p, Glyph{
			Rune:  r,
			Style: style,
		})
		p.X++
	}
}

// DrawStringRight draws a string right-justified.
func DrawStringRight(s TerminalDriver, b util.Rect, text string, style Style) {
	sx := b.TL.X + (b.Width() - len(text))
	si := 0
	if sx < b.TL.X {
		si += b.TL.X - sx
		sx = b.TL.X
	}
	for _, r := range text {
		if sx > b.BR.X || si >= len(text) {
			break
		}
		s.SetCell(util.Point{
			X: sx,
			Y: b.TL.Y,
		}, Glyph{
			Rune:  r,
			Style: style,
		})
		sx++
		si++
	}
}

// DrawStringCenter draws a string centered.
func DrawStringCenter(s TerminalDriver, b util.Rect, text string, style Style) {
	sx := (b.TL.X + (b.Width() / 2)) - (len(text) / 2)
	si := 0
	if sx < b.TL.X {
		si += b.TL.X - sx
		sx = b.TL.X
	}
	ex := b.BR.X
	for _, r := range text {
		if sx > ex || si >= len(text) {
			break
		}
		s.SetCell(util.Point{
			X: sx,
			Y: b.TL.Y,
		}, Glyph{
			Rune:  r,
			Style: style,
		})
		sx++
		si++
	}
}

// DrawHLine draws a horizontal line.
func DrawHLine(s TerminalDriver, p util.Point, w int, style Style) {
	g := Glyph{
		Rune:  '=',
		Style: style,
	}
	for i := 0; i < w; i++ {
		s.SetCell(p, g)
		p.X++
	}
}

// DrawVLine draws a vertical line.
func DrawVLine(s TerminalDriver, p util.Point, h int, style Style) {
	g := Glyph{
		Rune:  '|',
		Style: style,
	}
	for i := 0; i < h; i++ {
		s.SetCell(p, g)
		p.Y++
	}
}

// DrawBox draws a box.
func DrawBox(s TerminalDriver, b util.Rect, style Style) {
	DrawHLine(s, util.NewPoint(b.TL.X+1, b.TL.Y), b.Width()-2, style)
	DrawHLine(s, util.NewPoint(b.TL.X+1, b.TL.Y+b.Height()-1), b.Width()-2, style)
	DrawVLine(s, util.NewPoint(b.TL.X, b.TL.Y+1), b.Height()-2, style)
	DrawVLine(s, util.NewPoint(b.TL.X+b.Width()-1, b.TL.Y+1), b.Height()-2, style)
	s.SetCell(b.TL, Glyph{Rune: '+', Style: style})
	s.SetCell(util.NewPoint(b.BR.X, b.TL.Y), Glyph{Rune: '+', Style: style})
	s.SetCell(util.NewPoint(b.TL.X, b.BR.Y), Glyph{Rune: '+', Style: style})
	s.SetCell(b.BR, Glyph{Rune: '+', Style: style})
}
