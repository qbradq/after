package termgui

import (
	"time"

	_ "github.com/qbradq/after/internal/ai"
	"github.com/qbradq/after/internal/citygen"
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/internal/mods"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// MainMenu implements the main program menu.
type MainMenu struct {
	list termui.List // Main menu item list
}

var debugMods = []string{"Base"}

// NewMainMenu returns a new MainMenu object.
func NewMainMenu(s termui.TerminalDriver) *MainMenu {
	return &MainMenu{
		list: termui.List{
			Boxed: true,
			Title: "Main Menu",
			Items: []string{
				"New Game",
				"Load Game",
				"Quit",
			},
			Selected: func(s termui.TerminalDriver, n int) error {
				switch n {
				case 0:
					if err := mods.LoadMods(debugMods); err != nil {
						panic(err)
					}
					if err := game.NewSave("debug-"+time.Now().Format(time.DateTime), debugMods); err != nil {
						panic(err)
					}
					m := citygen.CityGens["Interstate Town"]()
					m.Player = game.NewPlayer(m.Now)
					m.Player.Position = util.NewPoint(10*game.ChunkWidth+game.ChunkWidth/2, 15*game.ChunkHeight+game.ChunkHeight/2)
					m.SaveCityPlan()
					m.Update(m.Player.Position, 0)
					m.FullSave()
					game.SaveTileRefs()
					termui.RunMode(s, NewGameMode(m))
					game.CloseSave()
				case 1:
					game.LoadSaveInfo()
					termui.RunMode(s, NewLoadMenu(s))
				case 2:
					return termui.ErrorQuit
				}
				return nil
			},
		},
	}
}

// HandleEvent implements the termui.Mode interface.
func (m *MainMenu) HandleEvent(s termui.TerminalDriver, e any) error {
	switch ev := e.(type) {
	case *termui.EventKey:
		if ev.Key == '\033' {
			return termui.ErrorQuit
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return m.list.HandleInput(s, e)
}

// Draw implements the termui.Mode interface.
func (m *MainMenu) Draw(s termui.TerminalDriver) {
	w, h := s.Size()
	termui.DrawClear(s)
	termui.DrawStringCenter(s, util.NewRectXYWH(0, (h/2)-6, w, 1), "After", termui.CurrentTheme.Normal.Foreground(termui.ColorLime))
	termui.DrawStringCenter(s, util.NewRectXYWH(0, (h/2)-5, w, 1), "by Norman B. Lancaster qbradq@gmail.com", termui.CurrentTheme.Normal.Foreground(termui.ColorGreen))
	m.list.Bounds = util.NewRectWH(w, h).CenterRect(14, 5)
	m.list.Draw(s)
}
