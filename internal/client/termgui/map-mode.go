package termgui

import (
	"strconv"

	"github.com/qbradq/after/internal/ai"
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// mapModeCallback is a callback function for the map mode cursor select.
type mapModeCallback func(util.Point, bool) error

// mapMode implements the main play area of the client.
type mapMode struct {
	CityMap     *game.CityMap   // City we are running
	Bounds      util.Rect       // Area of the map display on the screen
	Center      util.Point      // Centerpoint of the map display in absolute map coordinates
	CursorPos   util.Point      // Position of the cursor, if any
	CursorRange int             // Maximum range of the cursor
	CursorStyle int             // Cursor style
	Callback    mapModeCallback // Callback function to execute when the user selects a tile or quits
	DrawInfo    bool            // If true full tile information will be displayed next to the cursor
	DrawPaths   bool            // If true, draw the paths of all actors on screen
}

func (m *mapMode) topLeft() util.Point {
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
func (m *mapMode) HandleEvent(s termui.TerminalDriver, e any) error {
	switch ev := e.(type) {
	case *termui.EventKey:
		switch ev.Key {
		case 'u':
			m.CursorPos.X++
			m.CursorPos.Y--
		case 'y':
			m.CursorPos.X--
			m.CursorPos.Y--
		case 'n':
			m.CursorPos.X++
			m.CursorPos.Y++
		case 'b':
			m.CursorPos.X--
			m.CursorPos.Y++
		case 'l':
			m.CursorPos.X++
		case 'h':
			m.CursorPos.X--
		case 'j':
			m.CursorPos.Y++
		case 'k':
			m.CursorPos.Y--
		case ' ':
			fallthrough
		case '\n':
			if m.Callback != nil {
				return m.Callback(m.CursorPos, true)
			}
			return nil
		case '\033':
			if m.Callback != nil {
				return m.Callback(m.CursorPos, false)
			}
			return termui.ErrorQuit
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	if m.CursorRange > 0 {
		m.CursorPos =
			util.NewRectFromRadius(m.Center, m.CursorRange).Bound(m.CursorPos)
	}
	m.CursorPos = m.CityMap.TileBounds.Bound(m.CursorPos)
	mtl := m.topLeft()
	mb := util.NewRectXYWH(mtl.X, mtl.Y, m.Bounds.Width(), m.Bounds.Height())
	m.CursorPos = mb.Bound(m.CursorPos)
	return nil
}

// Draw implements the termui.Mode interface.
func (m *mapMode) Draw(s termui.TerminalDriver) {
	mtl := m.topLeft()
	mb := util.NewRectXYWH(mtl.X, mtl.Y, m.Bounds.Width(), m.Bounds.Height())
	m.CityMap.EnsureLoaded(mb.Divide(game.ChunkWidth))
	m.drawMap(s, mtl, mb)
	if m.DrawPaths {
		m.drawPaths(s, mtl, mb)
	}
}

func (m *mapMode) drawPaths(s termui.TerminalDriver, mtl util.Point, mb util.Rect) {
	// Draws a single path step
	fn := func(p util.Point, r int) {
		sp := util.NewPoint(p.X-mtl.X+m.Bounds.TL.X, p.Y-mtl.Y+m.Bounds.TL.Y)
		if !mb.Contains(p) {
			return
		}
		code := r & 0x000F
		fg := termui.Color(15 - ((r & 0x00F0) >> 4))
		bg := termui.Color((r & 0x0F00) >> 8)
		rn := '0' + rune(code)
		if rn > '9' {
			rn += 7
		}
		ns := termui.StyleDefault.
			Background(bg).
			Foreground(fg)
		s.SetCell(sp, termui.Glyph{
			Rune:  rn,
			Style: ns,
		})
	}
	// Display tile bounds
	db := util.Rect{
		TL: mtl,
		BR: util.Point{
			X: mtl.X + mb.Width(),
			Y: mtl.Y + mb.Height(),
		},
	}
	// Draw paths for every actor on screen
	for _, a := range m.CityMap.ActorsWithin(db) {
		// Draw first step at actor
		p := a.Position
		r := 0
		fn(p, r)
		// Step along the path and draw steps along the way
		for _, d := range a.AIModel.(*ai.AIModel).Path {
			r++
			p = p.Step(d)
			fn(p, r)
		}
	}
}

func (m *mapMode) drawMap(s termui.TerminalDriver, mtl util.Point, mb util.Rect) {
	m.CityMap.MakeVisibilitySets(mb)
	var p util.Point
	var idx uint32
	// Draw the tile matrix
	for p.Y = mtl.Y; p.Y < mtl.Y+m.Bounds.Height(); p.Y++ {
		for p.X = mtl.X; p.X < mtl.X+m.Bounds.Width(); p.X++ {
			sp := util.NewPoint(p.X-mtl.X+m.Bounds.TL.X, p.Y-mtl.Y+m.Bounds.TL.Y)
			if m.CityMap.Visibility.Contains(idx) {
				t := m.CityMap.GetTile(p)
				ns := termui.StyleDefault.
					Background(t.Bg).
					Foreground(t.Fg)
				s.SetCell(sp, termui.Glyph{
					Rune:  rune(t.Rune[0]),
					Style: ns,
				})
			} else if m.CityMap.Remembered.Contains(idx) {
				t := m.CityMap.GetTile(p)
				ns := termui.StyleDefault.
					Foreground(termui.ColorGray)
				s.SetCell(sp, termui.Glyph{
					Rune:  rune(t.Rune[0]),
					Style: ns,
				})
			} else {
				ns := termui.StyleDefault.
					Foreground(termui.ColorGray)
				s.SetCell(sp, termui.Glyph{
					Rune:  '?',
					Style: ns,
				})
			}
			idx++
		}
	}
	// Draw items
	for _, i := range m.CityMap.ItemsWithin(mb) {
		p := i.Position
		idx = uint32((p.Y-mtl.Y)*m.Bounds.Width() + (p.X - mtl.X))
		if m.CityMap.Visibility.Contains(idx) {
			sp := util.NewPoint((p.X-mtl.X)+m.Bounds.TL.X, (p.Y-mtl.Y)+m.Bounds.TL.Y)
			ns := termui.StyleDefault.
				Background(i.Bg).
				Foreground(i.Fg)
			s.SetCell(sp, termui.Glyph{
				Rune:  rune(i.Rune[0]),
				Style: ns,
			})
		} else if m.CityMap.Remembered.Contains(idx) {
			sp := util.NewPoint((p.X-mtl.X)+m.Bounds.TL.X, (p.Y-mtl.Y)+m.Bounds.TL.Y)
			ns := termui.StyleDefault.
				Foreground(termui.ColorGray)
			s.SetCell(sp, termui.Glyph{
				Rune:  rune(i.Rune[0]),
				Style: ns,
			})
		}
	}
	// Draw actors
	for _, a := range m.CityMap.ActorsWithin(mb) {
		p := a.Position
		idx = uint32((p.Y-mtl.Y)*m.Bounds.Width() + (p.X - mtl.X))
		if m.CityMap.Visibility.Contains(idx) {
			sp := util.NewPoint((p.X-mtl.X)+m.Bounds.TL.X, (p.Y-mtl.Y)+m.Bounds.TL.Y)
			ns := termui.StyleDefault.
				Background(a.Bg).
				Foreground(a.Fg)
			r := rune(a.Rune[0])
			if a.Dead {
				r = '%'
				ns = termui.StyleDefault.
					Background(termui.ColorBlack).
					Foreground(termui.ColorOlive)
			}
			s.SetCell(sp, termui.Glyph{
				Rune:  r,
				Style: ns,
			})
		}
	}
	// Draw vehicles
	for _, v := range m.CityMap.VehiclesWithin(mb) {
		var p util.Point
		for p.Y = 0; p.Y < v.Bounds.Height(); p.Y++ {
			for p.X = 0; p.X < v.Bounds.Width(); p.X++ {
				l := v.Location(p)
				if l == nil || len(l.Parts) < 1 {
					continue
				}
				vp := v.Bounds.TL
				sp := util.NewPoint((vp.X-mtl.X)+m.Bounds.TL.X, (vp.Y-mtl.Y)+m.Bounds.TL.Y).Add(p)
				if !m.Bounds.Contains(sp) {
					continue
				}
				i := l.Parts[len(l.Parts)-1]
				vx := vp.X + p.X
				vy := vp.Y + p.Y
				idx = uint32((vy-mtl.Y)*m.Bounds.Width() + (vx - mtl.X))
				if m.CityMap.Visibility.Contains(idx) {
					ns := termui.StyleDefault.
						Background(i.Bg).
						Foreground(i.Fg)
					s.SetCell(sp, termui.Glyph{
						Rune:  rune(i.Rune[0]),
						Style: ns,
					})
				} else if m.CityMap.Remembered.Contains(idx) {
					ns := termui.StyleDefault.
						Foreground(termui.ColorGray)
					s.SetCell(sp, termui.Glyph{
						Rune:  rune(i.Rune[0]),
						Style: ns,
					})
				}
			}
		}
	}
	// Draw the player
	a := m.CityMap.Player
	p = a.Position
	sp := util.NewPoint((p.X-mtl.X)+m.Bounds.TL.X, (p.Y-mtl.Y)+m.Bounds.TL.Y)
	if m.Bounds.Contains(sp) {
		ns := termui.StyleDefault.
			Background(a.Bg).
			Foreground(a.Fg)
		s.SetCell(sp, termui.Glyph{
			Rune:  rune(a.Rune[0]),
			Style: ns,
		})
	}
	// Draw the cursor
	sp = util.Point{
		X: (m.CursorPos.X - mtl.X) + m.Bounds.TL.X,
		Y: (m.CursorPos.Y - mtl.Y) + m.Bounds.TL.Y,
	}
	if m.Bounds.Contains(sp) {
		drawCursor(s, sp, m.Bounds, m.CursorStyle)
	}
	// Draw info box if needed
	if m.DrawInfo {
		idx = uint32((m.CursorPos.Y-mtl.Y)*m.Bounds.Width() + (m.CursorPos.X - mtl.X))
		if m.CityMap.Visibility.Contains(idx) {
			t := m.CityMap.GetTile(m.CursorPos)
			a := m.CityMap.ActorAt(m.CursorPos)
			items := m.CityMap.ItemsAt(m.CursorPos)
			h := 1 + len(items)
			w := len(t.Name)
			if a != nil && len(a.Name) > w {
				w = len(a.Name)
			}
			if a != nil {
				h++
			}
			for _, i := range items {
				nl := len(i.Name)
				if i.Amount > 1 {
					nl += 2 + len(strconv.FormatInt(int64(i.Amount), 10))
				}
				if nl > w {
					w = nl
				}
			}
			if len(items) == 0 {
				h++
				if w < len("Nothing") {
					w = len("Nothing")
				}
			}
			dx := sp.X + 2
			if m.CursorPos.X > m.Center.X {
				dx = sp.X - (3 + w)
			}
			r := util.NewRectXYWH(dx, sp.Y-1, w+3, h+2)
			r = m.Bounds.Contain(r)
			termui.DrawBox(s, r, termui.CurrentTheme.Normal)
			r.TL.X++
			r.TL.Y++
			r.BR.X--
			r.BR.Y--
			termui.DrawFill(s, r, termui.Glyph{
				Rune:  ' ',
				Style: termui.CurrentTheme.Normal,
			})
			termui.DrawStringLeft(s, r, t.Name, termui.CurrentTheme.Normal)
			r.TL.Y++
			if a != nil {
				termui.DrawStringLeft(s, r, a.Name, termui.CurrentTheme.Normal.Foreground(termui.ColorLime))
				r.TL.Y++
			}
			for _, i := range items {
				n := " " + i.Name
				if i.Container {
					if len(i.Inventory) > 0 {
						n = "+" + i.Name
					} else {
						n = "-" + i.Name
					}
				}
				if i.Amount > 1 {
					n = n + " x" + strconv.FormatInt(int64(i.Amount), 10)
				}
				termui.DrawStringLeft(s, r, n, termui.CurrentTheme.Normal.Foreground(termui.ColorAqua))
				r.TL.Y++
			}
			if len(items) == 0 {
				termui.DrawStringCenter(s, r, "Nothing", termui.CurrentTheme.Normal.Foreground(termui.ColorGray))
			}
		} else if m.CityMap.Remembered.Contains(idx) {
			t := m.CityMap.GetTile(m.CursorPos)
			h := 2
			w := len("Remembered")
			if len(t.Name) > w {
				w = len(t.Name)
			}
			dx := sp.X + 2
			if m.CursorPos.X > m.Center.X {
				dx = sp.X - (3 + w)
			}
			r := util.NewRectXYWH(dx, sp.Y-1, w+2, h+2)
			r = m.Bounds.Contain(r)
			termui.DrawBox(s, r, termui.CurrentTheme.Normal)
			r.TL.X++
			r.TL.Y++
			r.BR.X--
			r.BR.Y--
			termui.DrawFill(s, r, termui.Glyph{
				Rune:  ' ',
				Style: termui.CurrentTheme.Normal,
			})
			termui.DrawStringLeft(s, r, t.Name, termui.CurrentTheme.Normal)
			r.TL.Y++
			termui.DrawStringCenter(s, r, "Remembered", termui.CurrentTheme.Normal.Foreground(termui.ColorGray))
		} else {
			h := 1
			w := len("Unseen")
			dx := sp.X + 2
			if m.CursorPos.X > m.Center.X {
				dx = sp.X - (3 + w)
			}
			r := util.NewRectXYWH(dx, sp.Y-1, w+2, h+2)
			r = m.Bounds.Contain(r)
			termui.DrawBox(s, r, termui.CurrentTheme.Normal)
			r.TL.X++
			r.TL.Y++
			r.BR.X--
			r.BR.Y--
			termui.DrawStringLeft(s, r, "Unseen", termui.CurrentTheme.Normal)
		}
	}
}
