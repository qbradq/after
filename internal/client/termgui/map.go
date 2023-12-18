package termgui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/qbradq/after/lib/util"
)

func mapTopLeft(center util.Point, area util.Rect) util.Point {
	// Calculate top-left corner
	ret := util.NewPoint(center.X-area.Width()/2,
		center.Y-area.Height()/2)
	if ret.X < 0 {
		ret.X = 0
	}
	if ret.X >= cityMap.Bounds.Width()-area.Width() {
		ret.X = (cityMap.Bounds.Width() - area.Width()) - 1
	}
	if ret.Y < 0 {
		ret.Y = 0
	}
	if ret.Y >= cityMap.Bounds.Height()-area.Height() {
		ret.Y = (cityMap.Bounds.Height() - area.Height()) - 1
	}
	return ret
}

func mapPointToScreen(p util.Point, area util.Rect) util.Point {
	tl := mapTopLeft(p, area)
	return util.NewPoint(p.X-tl.X, p.Y-tl.Y)
}

func drawMap(center util.Point, area util.Rect, cursor util.Point, cursorStyle int) {
	mtl := mapTopLeft(center, area)
	cityMap.Load(util.NewRectXYWH(mtl.X, mtl.Y, area.Width(), area.Height()))
	var p util.Point
	for p.Y = mtl.Y; p.Y < mtl.Y+area.Height(); p.Y++ {
		for p.X = mtl.X; p.X < mtl.X+area.Width(); p.X++ {
			sp := util.NewPoint(p.X-mtl.X+area.TL.X, p.Y-mtl.Y+area.TL.Y)
			t := cityMap.GetTile(p)
			ns := tcell.StyleDefault.
				Background(tcell.Color(t.Bg)).
				Foreground(tcell.Color(t.Fg))
			screen.SetContent(sp.X, sp.Y, rune(t.Rune[0]), nil, ns)
		}
	}
	drawCursor(util.Point{
		X: (center.X - mtl.X) + area.TL.X,
		Y: (center.Y - mtl.Y) + area.TL.Y,
	}, area, cursorStyle)
}
