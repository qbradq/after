package termgui

import (
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// minimap implements a termui.Mode that displays the minimap
type minimap struct {
	CityMap     *game.CityMap      // The city we are displaying
	Bounds      util.Rect          // Bounds of the minimap display on screen
	Center      util.Point         // Current centerpoint of the display
	DrawInfo    bool               // If true draws a box containing info near the cursor
	CursorStyle int                // Cursor style
	Selected    func(p util.Point) // Function to execute on selection, if any
}

func (m *minimap) topLeft() util.Point {
	// Calculate top-left corner
	ret := util.NewPoint(m.Center.X-m.Bounds.Width()/2,
		m.Center.Y-m.Bounds.Height()/2)
	if ret.X < 0 {
		ret.X = 0
	}
	if ret.X >= game.CityMapWidth-m.Bounds.Width() {
		ret.X = (game.CityMapWidth - m.Bounds.Width()) - 1
	}
	if ret.Y < 0 {
		ret.Y = 0
	}
	if ret.Y >= game.CityMapHeight-m.Bounds.Height() {
		ret.Y = (game.CityMapHeight - m.Bounds.Height()) - 1
	}
	return ret
}

func (m *minimap) pointToScreen(p util.Point) util.Point {
	tl := m.topLeft()
	return util.NewPoint(p.X-tl.X, p.Y-tl.Y)
}

// HandleEvent implements the termui.Mode interface.
func (m *minimap) HandleEvent(s termui.TerminalDriver, e any) error {
	// State update
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
		case ' ':
			fallthrough
		case '\n':
			if m.Selected != nil {
				m.Selected(m.Center)
				return termui.ErrorQuit
			}
		case '\033':
			return termui.ErrorQuit
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	m.Center = m.CityMap.Bounds.Bound(m.Center)
	return nil
}

// Draw implements the termui.Mode interface.
func (m *minimap) Draw(s termui.TerminalDriver) {
	// Render the minimap and cursor
	mmTL := m.topLeft()
	for iy := 0; iy < m.Bounds.Height(); iy++ {
		for ix := 0; ix < m.Bounds.Width(); ix++ {
			c := m.CityMap.GetChunkFromMapPoint(util.Point{X: ix + mmTL.X, Y: iy + mmTL.Y})
			s.SetCell(util.NewPoint(ix+m.Bounds.TL.X, iy+m.Bounds.TL.Y), termui.Glyph{
				Rune: rune(c.MinimapRune[0]),
				Style: termui.StyleDefault.
					Background(c.MinimapBackground).
					Foreground(c.MinimapForeground),
			})
		}
	}
	drawCursor(s, util.Point{
		X: (m.Center.X - mmTL.X) + m.Bounds.TL.X,
		Y: (m.Center.Y - mmTL.Y) + m.Bounds.TL.Y,
	}, m.Bounds, m.CursorStyle)
	// Draw the nameplate if requested
	if m.DrawInfo {
		c := m.CityMap.GetChunkFromMapPoint(m.Center)
		sp := m.pointToScreen(m.Center)
		if sp.X > m.Bounds.Width()/2 {
			sp.X -= 2 + len(c.Name)
		} else {
			sp.X += 2
		}
		sp.Y--
		if sp.Y < 0 {
			sp.Y = 0
		}
		if sp.Y > m.Bounds.Height()-3 {
			sp.Y = m.Bounds.Height() - 3
		}
		termui.DrawBox(s, util.NewRectXYWH(sp.X, sp.Y, len(c.Name)+2, 3), termui.CurrentTheme.Normal)
		termui.DrawStringLeft(s, util.NewRectXYWH(sp.X+1, sp.Y+1, len(c.Name), 1), c.Name, termui.CurrentTheme.Normal)
	}
}
