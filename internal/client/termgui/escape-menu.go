package termgui

import (
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// escapeMenu implements the system menu that appears when you press escape.
type escapeMenu struct {
	m       *gameMode   // Game mode back reference
	list    termui.List // Menu list
	debug   termui.List // Debug menu list
	inDebug bool        // If true, display the debug menu
}

// newEscapeMenu returns a new EscapeMenu ready for use.
func newEscapeMenu(m *gameMode) *escapeMenu {
	var ret *escapeMenu
	ret = &escapeMenu{
		m: m,
		list: termui.List{
			Boxed: true,
			Items: []string{
				"Resume",
				"Force Save",
				"Save and Quit",
				"Debug Menu",
			},
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
				case 3:
					ret.inDebug = true
					return nil
				}
				return termui.ErrorQuit
			},
		},
		debug: termui.List{
			Boxed: true,
			Items: []string{
				"Teleport to Chunk",
				"Log Chunk Info",
				"Generate Vehicle",
				"Toggle Debug Display",
			},
			Title: "Game Menu",
			Selected: func(td termui.TerminalDriver, i int) error {
				switch i {
				case 0:
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
								ws, cs := c.CanStep(&m.CityMap.Player.Actor, dp, m.CityMap)
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
				case 1:
					c := m.CityMap.GetChunk(m.CityMap.Player.Position)
					b, err := c.Facing.MarshalJSON()
					if err != nil {
						panic(err)
					}
					facing := string(b)
					m.logMode.Log(
						termui.ColorLime,
						"Chunk Info: Gen=%s, Var=%s, Facing=%s",
						c.Generator.GetGroup(),
						c.Generator.GetVariant(),
						facing,
					)
					return termui.ErrorQuit
				case 2:
					v := game.GenerateVehicle("Street", ret.m.CityMap.Now)
					if v == nil {
						game.Log.Log(termui.ColorRed, "Failed to generate vehicle.")
						return termui.ErrorQuit
					}
					v.Facing = util.FacingNorth
					v.Heading = v.Facing.Direction()
					v.Bounds = util.NewRectWH(v.Size.X, v.Size.Y).Move(ret.m.CityMap.Player.Position)
					if !ret.m.CityMap.PlaceVehicle(v) {
						game.Log.Log(termui.ColorRed, "Failed to place vehicle.")
						return termui.ErrorQuit
					}
					ret.m.CityMap.FlagBitmapsForVehicle(v)
					ret.m.logMode.Log(termui.ColorFuchsia, "New vehicle bounds: %v", v.Bounds)
					ret.m.CityMap.Update(ret.m.CityMap.Player.Position, 0, nil)
					return termui.ErrorQuit
				case 3:
					m.debug = !m.debug
					return termui.ErrorQuit
				case 4:
					if cpuProfile == nil {
						beginCPUProfile()
					} else {
						endCPUProfile()
					}
					return termui.ErrorQuit
				}
				return nil
			},
			Closed: func(td termui.TerminalDriver) error {
				ret.inDebug = false
				return termui.ErrorQuit
			},
		},
	}
	if cpuProfile == nil {
		ret.debug.Items = append(ret.debug.Items, "Begin CPU Profile")
	} else {
		ret.debug.Items = append(ret.debug.Items, "End CPU Profile")
	}
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
	if m.inDebug {
		return m.debug.HandleEvent(s, e)
	}
	return m.list.HandleEvent(s, e)
}

// Draw implements the termui.Mode interface.
func (m *escapeMenu) Draw(s termui.TerminalDriver) {
	// Main menu
	w, h := s.Size()
	maxW := 0
	for _, i := range m.list.Items {
		l := len(i)
		if l > maxW {
			maxW = l
		}
	}
	m.list.Bounds = util.NewRectWH(w, h).CenterRect(2+maxW, 2+len(m.list.Items))
	m.list.Draw(s)
	if !m.inDebug {
		return
	}
	// Debug menu
	maxW = 0
	for _, i := range m.debug.Items {
		l := len(i)
		if l > maxW {
			maxW = l
		}
	}
	m.debug.Bounds = util.NewRectWH(w, h).CenterRect(2+maxW, 2+len(m.debug.Items))
	m.debug.Draw(s)
}
