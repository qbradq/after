package termgui

import (
	"github.com/gdamore/tcell/v2"
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
	runLoop(func(e tcell.Event) bool {
		// State update
		switch ev := e.(type) {
		case *tcell.EventKey:
			switch ev.Rune() {
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
		screen.Clear()
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
		termui.DrawBox(screen, sp.X, sp.Y, len(c.Name)+2, 3, termui.CurrentTheme.Normal)
		termui.DrawStringLeft(screen, sp.X+1, sp.Y+1, len(c.Name), c.Name, termui.CurrentTheme.Normal)
		return false
	})
}

func drawMinimap(center util.Point, area util.Rect, cursorStyle int) {
	mmtl := minimapTopLeft(center, area)
	for iy := 0; iy < area.Height(); iy++ {
		for ix := 0; ix < area.Width(); ix++ {
			c := cityMap.GetChunkFromMapPoint(util.Point{X: ix + mmtl.X, Y: iy + mmtl.Y})
			termui.DrawRune(screen, ix+area.TL.X, iy+area.TL.Y, rune(c.MinimapRune[0]),
				tcell.StyleDefault.
					Background(tcell.Color(c.MinimapBackground)).
					Foreground(tcell.Color(c.MinimapForeground)))
		}
	}
	drawCursor(util.Point{
		X: (center.X - mmtl.X) + area.TL.X,
		Y: (center.Y - mmtl.Y) + area.TL.Y,
	}, area, cursorStyle)
}
