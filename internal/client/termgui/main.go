package termgui

import (
	"os"
	"os/signal"

	"github.com/gdamore/tcell/v2"
	"github.com/qbradq/after/internal/citygen"
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/internal/mods"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

var screen tcell.Screen
var cityMap *game.CityMap

func init() {
	var err error
	// Init tcell screen
	screen, err = tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	// Trap signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			gracefulShutdown()
			os.Exit(0)
		}
	}()
}

func gracefulShutdown() {
	screen.Fini()
}

func Main() {
	screen.Init()
	defer gracefulShutdown()
	mainMenu()
}

func runLoop(fn func(tcell.Event) bool) {
	var event tcell.Event
	fn(nil)
	screen.Show()
	quit := false
	for !quit {
		event = screen.PollEvent()
		switch event.(type) {
		case *tcell.EventResize:
			quit = fn(nil)
			screen.Sync()
			continue
		default:
		}
		quit = fn(event)
		screen.Show()
	}
}

func mainMenu() {
	pos := 0
	runLoop(func(e tcell.Event) bool {
		quit := false
		w, h := screen.Size()
		termui.DrawClear(screen)
		termui.DrawStringCenter(screen, 0, (h/2)-6, w, "After", termui.CurrentTheme.Normal.Foreground(tcell.ColorLime))
		termui.DrawStringCenter(screen, 0, (h/2)-5, w, "by Norman B. Lancaster qbradq@gmail.com", termui.CurrentTheme.Normal.Foreground(tcell.ColorGreen))
		termui.BoxList(screen, (w-14)/2, (h/2)-3, 14, 5, "Main Menu", []string{
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
		return quit
	})
}

func gameMode() {
	p := util.NewPoint(10*game.ChunkWidth+game.ChunkWidth/2, 15*game.ChunkHeight+game.ChunkHeight/2)
	runLoop(func(e tcell.Event) bool {
		// Respond to input
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
			case 'm':
				minimapMapMode(util.NewPoint(p.X/game.ChunkWidth, p.Y/game.ChunkHeight))
			}
		}
		// Draw the screen
		termui.DrawClear(screen)
		sw, sh := screen.Size()
		drawMap(p, util.NewRectWH(sw-56, sh), p, 0)
		termui.DrawVLine(screen, sw-56, 0, sh, termui.CurrentTheme.Normal)
		drawMinimap(util.NewPoint(p.X/game.ChunkWidth, p.Y/game.ChunkHeight),
			util.NewRectXYWH(sw-21, 0, 21, 21), 1)
		return false
	})
}
