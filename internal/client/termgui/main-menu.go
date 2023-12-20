package termgui

import (
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

// NewMainMenu returns a new MainMenu object.
func NewMainMenu() *MainMenu {
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
					if err := mods.LoadModByID("Base"); err != nil {
						panic(err)
					}
					m := citygen.CityGens["Interstate Town"]()
					termui.RunMode(s, NewGameMode(m, util.NewPoint(10*game.ChunkWidth+game.ChunkWidth/2, 15*game.ChunkHeight+game.ChunkHeight/2)))
				case 1:
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
	return m.list.HandleInput(s, e)
}

// Draw implements the termui.Mode interface.
func (m *MainMenu) Draw(s termui.TerminalDriver) {
	w, h := s.Size()
	termui.DrawClear(s)
	termui.DrawStringCenter(s, util.NewRectXYWH(0, (h/2)-6, w, 1), "After", termui.CurrentTheme.Normal.Foreground(termui.ColorLime))
	termui.DrawStringCenter(s, util.NewRectXYWH(0, (h/2)-5, w, 1), "by Norman B. Lancaster qbradq@gmail.com", termui.CurrentTheme.Normal.Foreground(termui.ColorGreen))
	m.list.Bounds = util.NewRectXYWH((w-14)/2, (h/2)-3, 14, 5)
	m.list.Draw(s)
}
