package termgui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/qbradq/after/lib/util"
)

func drawCursor(sp util.Point, area util.Rect, cursorStyle int) {
	// Render the cursor
	fn := func(p util.Point) {
		if !area.Contains(p) {
			return
		}
		r, _, s, _ := screen.GetContent(p.X, p.Y)
		fg, bg, _ := s.Decompose()
		ns := tcell.StyleDefault.Background(fg).Foreground(bg)
		screen.SetContent(p.X, p.Y, r, nil, ns)
	}
	switch cursorStyle {
	case 1:
		fn(sp)
	case 2:
		sp.Y--
		fn(sp)
		sp.Y += 2
		fn(sp)
		sp.Y--
		sp.X--
		fn(sp)
		sp.X += 2
		fn(sp)
	}
}
