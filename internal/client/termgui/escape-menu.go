package termgui

import (
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// escapeMenu implements the system menu that appears when you press escape.
type escapeMenu struct {
	m    *gameMode   // Game mode back reference
	list termui.List // Menu list
}

// newEscapeMenu returns a new EscapeMenu ready for use.
func newEscapeMenu(m *gameMode) *escapeMenu {
	ret := &escapeMenu{
		m: m,
		list: termui.List{
			Boxed: true,
			Items: []string{"Resume", "Force Save", "Save and Quit", "_hbar_"},
			Title: "Game Menu",
			Selected: func(td termui.TerminalDriver, i int) error {
				switch i {
				case 0:
					return termui.ErrorQuit
				case 1:
					m.CityMap.FullSave()
					return termui.ErrorQuit
				case 2:
					m.CityMap.FullSave()
					m.quit = true
					return termui.ErrorQuit
				case 4:
					m.debug = !m.debug
					return termui.ErrorQuit
				case 5:
					termui.RunMode(td, &minimap{
						CityMap:     m.CityMap,
						Bounds:      util.NewRectWH(td.Size()),
						Center:      util.NewPoint(m.CityMap.Player.Position.X/game.ChunkWidth, m.CityMap.Player.Position.Y/game.ChunkHeight),
						CursorStyle: 2,
						DrawInfo:    true,
						Selected: func(p util.Point) {
							c := m.CityMap.GetChunkFromMapPoint(p)
							m.CityMap.LoadChunk(c, m.CityMap.Now)
							p = p.Multiply(game.ChunkWidth)
							for i := 0; i < 512; i++ {
								dp := p
								dp.X += util.Random(0, game.ChunkWidth)
								dp.Y += util.Random(0, game.ChunkHeight)
								ws, cs := c.CanStep(&m.CityMap.Player.Actor, dp)
								if ws || cs {
									m.CityMap.Player.Position = dp
									m.logMode.Log(termui.ColorLime, "Teleported to %dx%d.", dp.X, dp.Y)
									m.CityMap.Update(dp, 0, nil)
									return
								}
							}
							m.logMode.Log(termui.ColorRed, "Teleport function ran out of stand attempts.")
						},
					})
					return termui.ErrorQuit
				}
				return nil
			},
		},
	}
	if m.debug {
		ret.list.Items = append(ret.list.Items, "Disable Debug Display")
	} else {
		ret.list.Items = append(ret.list.Items, "Enable Debug Display")
	}
	ret.list.Items = append(ret.list.Items, "Teleport Menu")
	return ret
}

// HandleEvent implements the termui.Mode interface.
func (m *escapeMenu) HandleEvent(s termui.TerminalDriver, e any) error {
	switch ev := e.(type) {
	case *termui.EventKey:
		if ev.Key == '\033' {
			return termui.ErrorQuit
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return m.list.HandleEvent(s, e)
}

// Draw implements the termui.Mode interface.
func (m *escapeMenu) Draw(s termui.TerminalDriver) {
	w, h := s.Size()
	m.list.Bounds = util.NewRectWH(w, h).CenterRect(27, 7)
	m.list.Draw(s)
}
