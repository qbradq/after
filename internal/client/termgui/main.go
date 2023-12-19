package termgui

import (
	"log"

	"github.com/qbradq/after/internal/citygen"
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/internal/mods"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

var screen termui.TerminalDriver
var cityMap *game.CityMap

func Main(s termui.TerminalDriver) {
	screen = s
	if err := screen.Init(); err != nil {
		log.Fatal(err)
	}
	defer screen.Fini()
	mainMenu()
}

func mainMenu() {
	pos := 0
	termui.Run(screen, func(e any) error {
		quit := false
		w, h := screen.Size()
		termui.DrawClear(screen)
		termui.DrawStringCenter(screen, util.NewRectXYWH(0, (h/2)-6, w, 1), "After", termui.CurrentTheme.Normal.Foreground(termui.ColorLime))
		termui.DrawStringCenter(screen, util.NewRectXYWH(0, (h/2)-5, w, 1), "by Norman B. Lancaster qbradq@gmail.com", termui.CurrentTheme.Normal.Foreground(termui.ColorGreen))
		termui.BoxList(screen, util.NewRectXYWH((w-14)/2, (h/2)-3, 14, 5), "Main Menu", []string{
			"New Game",
			"Continue",
			"Quit",
		}, &pos, e, func(n int) {
			switch n {
			case 0:
				if err := mods.LoadModByID("Base"); err != nil {
					panic(err)
				}
				cityMap = citygen.CityGens["Interstate Town"]()
				gameMode()
			case 1:
			case 2:
				quit = true
			}
		})
		if quit {
			return termui.ErrorQuit
		}
		return nil
	})
}

func gameMode() {
	p := util.NewPoint(10*game.ChunkWidth+game.ChunkWidth/2, 15*game.ChunkHeight+game.ChunkHeight/2)
	termui.Run(screen, func(e any) error {
		// Respond to input
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
			case 'm':
				minimapMapMode(util.NewPoint(p.X/game.ChunkWidth, p.Y/game.ChunkHeight))
			}
		}
		// Draw the screen
		termui.DrawClear(screen)
		sw, sh := screen.Size()
		drawMap(p, util.NewRectWH(sw-56, sh), p, 0)
		termui.DrawVLine(screen, util.NewPoint(sw-56, 0), sh, termui.CurrentTheme.Normal)
		drawMinimap(util.NewPoint(p.X/game.ChunkWidth, p.Y/game.ChunkHeight),
			util.NewRectXYWH(sw-21, 0, 21, 21), 1)
		return nil
	})
}
