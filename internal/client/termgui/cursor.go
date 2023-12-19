package termgui

import (
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

func drawCursor(sp util.Point, area util.Rect, cursorStyle int) {
	// Render the cursor
	fn := func(p util.Point) {
		if !area.Contains(p) {
			return
		}
		g := screen.GetCell(p)
		fg, bg := g.Style.Decompose()
		g.Style = termui.StyleDefault.Background(fg).Foreground(bg)
		screen.SetCell(p, g)
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
