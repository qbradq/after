package termgui

import (
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

func minimapTopLeft(center util.Point, area util.Rect) util.Point {
	// Calculate top-left corner
	ret := util.NewPoint(center.X-area.Width()/2,
		center.Y-area.Height()/2)
	if ret.X < 0 {
		ret.X = 0
	}
	if ret.X >= game.CityMapWidth-area.Width() {
		ret.X = (game.CityMapWidth - area.Width()) - 1
	}
	if ret.Y < 0 {
		ret.Y = 0
	}
	if ret.Y >= game.CityMapHeight-area.Height() {
		ret.Y = (game.CityMapHeight - area.Height()) - 1
	}
	return ret
}

func minimapPointToScreen(p util.Point, area util.Rect) util.Point {
	tl := minimapTopLeft(p, area)
	return util.NewPoint(p.X-tl.X, p.Y-tl.Y)
}

func minimapMapMode(center util.Point) {
	p := center
	termui.Run(screen, func(e any) error {
		// State update
		switch ev := e.(type) {
		case *termui.EventKey:
			switch ev.Key {
			case 'u':
				p.X++
				p.Y--
			case 'y':
				p.X--
				p.Y--
			case 'n':
				p.X++
				p.Y++
			case 'b':
				p.X--
				p.Y++
			case 'l':
				p.X++
			case 'h':
				p.X--
			case 'j':
				p.Y++
			case 'k':
				p.Y--
			case '\033':
				return termui.ErrorQuit
			}
		}
		// Bound focal point
		if p.X < 0 {
			p.X = 0
		}
		if p.X >= game.CityMapWidth {
			p.X = game.CityMapWidth - 1
		}
		if p.Y < 0 {
			p.Y = 0
		}
		if p.Y >= game.CityMapHeight {
			p.Y = game.CityMapHeight - 1
		}
		// Map sampling
		c := cityMap.GetChunkFromMapPoint(p)
		// Drawing
		sw, sh := screen.Size()
		termui.DrawClear(screen)
		drawMinimap(p, util.NewRectWH(sw, sh), 2)
		sp := minimapPointToScreen(p, util.NewRectWH(sw, sh))
		if sp.X > sw/2 {
			sp.X -= 2 + len(c.Name)
		} else {
			sp.X += 2
		}
		sp.Y--
		if sp.Y < 0 {
			sp.Y = 0
		}
		if sp.Y > sh-3 {
			sp.Y = sh - 3
		}
		termui.DrawBox(screen, util.NewRectXYWH(sp.X, sp.Y, len(c.Name)+2, 3), termui.CurrentTheme.Normal)
		termui.DrawStringLeft(screen, util.NewRectXYWH(sp.X+1, sp.Y+1, len(c.Name), 1), c.Name, termui.CurrentTheme.Normal)
		return nil
	})
}

func drawMinimap(center util.Point, area util.Rect, cursorStyle int) {
	mmtl := minimapTopLeft(center, area)
	for iy := 0; iy < area.Height(); iy++ {
		for ix := 0; ix < area.Width(); ix++ {
			c := cityMap.GetChunkFromMapPoint(util.Point{X: ix + mmtl.X, Y: iy + mmtl.Y})
			screen.SetCell(util.NewPoint(ix+area.TL.X, iy+area.TL.Y), termui.Glyph{
				Rune: rune(c.MinimapRune[0]),
				Style: termui.StyleDefault.
					Background(c.MinimapBackground).
					Foreground(c.MinimapForeground),
			})
		}
	}
	drawCursor(util.Point{
		X: (center.X - mmtl.X) + area.TL.X,
		Y: (center.Y - mmtl.Y) + area.TL.Y,
	}, area, cursorStyle)
}
