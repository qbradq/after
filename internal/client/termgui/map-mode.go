package termgui

import (
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// MapMode implements the main play area of the client.
type MapMode struct {
	CityMap     *game.CityMap // City we are running
	Bounds      util.Rect     // Area of the map display on the screen
	Center      util.Point    // Centerpoint of the map display in absolute map coordinates
	CursorStyle int           // Cursor style
}

func (m *MapMode) topLeft() util.Point {
	// Calculate top-left corner
	ret := util.NewPoint(m.Center.X-m.Bounds.Width()/2,
		m.Center.Y-m.Bounds.Height()/2)
	if ret.X < 0 {
		ret.X = 0
	}
	if ret.X >= m.CityMap.Bounds.Width()*game.ChunkWidth-m.Bounds.Width() {
		ret.X = (m.CityMap.Bounds.Width()*game.ChunkWidth - m.Bounds.Width()) - 1
	}
	if ret.Y < 0 {
		ret.Y = 0
	}
	if ret.Y >= m.CityMap.Bounds.Height()*game.ChunkHeight-m.Bounds.Height() {
		ret.Y = (m.CityMap.Bounds.Height()*game.ChunkHeight - m.Bounds.Height()) - 1
	}
	return ret
}

// HandleEvent implements the termui.Mode interface.
func (m *MapMode) HandleEvent(s termui.TerminalDriver, e any) error {
	switch ev := e.(type) {
	case *termui.EventKey:
		switch ev.Key {
		case 'u':
			m.Center.X++
			m.Center.Y--
		case 'y':
			m.Center.X--
			m.Center.Y--
		case 'n':
			m.Center.X++
			m.Center.Y++
		case 'b':
			m.Center.X--
			m.Center.Y++
		case 'l':
			m.Center.X++
		case 'h':
			m.Center.X--
		case 'j':
			m.Center.Y++
		case 'k':
			m.Center.Y--
		case 'm':
			termui.RunMode(s, &Minimap{
				CityMap:     m.CityMap,
				Bounds:      util.NewRectWH(s.Size()),
				Center:      util.NewPoint(m.Center.X/game.ChunkWidth, m.Center.Y/game.ChunkHeight),
				CursorStyle: 2,
				DrawInfo:    true,
			})
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return nil
}

// Draw implements the termui.Mode interface.
func (m *MapMode) Draw(s termui.TerminalDriver) {
	mtl := m.topLeft()
	m.CityMap.Load(util.NewRectXYWH(mtl.X, mtl.Y, m.Bounds.Width(), m.Bounds.Height()))
	var p util.Point
	for p.Y = mtl.Y; p.Y < mtl.Y+m.Bounds.Height(); p.Y++ {
		for p.X = mtl.X; p.X < mtl.X+m.Bounds.Width(); p.X++ {
			sp := util.NewPoint(p.X-mtl.X+m.Bounds.TL.X, p.Y-mtl.Y+m.Bounds.TL.Y)
			t := m.CityMap.GetTile(p)
			ns := termui.StyleDefault.
				Background(t.Bg).
				Foreground(t.Fg)
			s.SetCell(sp, termui.Glyph{
				Rune:  rune(t.Rune[0]),
				Style: ns,
			})
		}
	}
	drawCursor(s, util.Point{
		X: (m.Center.X - mtl.X) + m.Bounds.TL.X,
		Y: (m.Center.Y - mtl.Y) + m.Bounds.TL.Y,
	}, m.Bounds, m.CursorStyle)
}
