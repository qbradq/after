package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/qbradq/after/internal/client/termgui"
	ebitendriver "github.com/qbradq/after/lib/ebiten-driver"
	"github.com/qbradq/after/lib/termui"
)

func main() {
	s := ebitendriver.New()
	s.Init()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			s.Fini()
			os.Exit(0)
		}
	}()
	go func() {
		// pf, err := os.Create("cpu.pprof")
		// if err != nil {
		// 	panic(err)
		// }
		// pprof.StartCPUProfile(pf)
		termui.RunMode(s, termgui.NewMainMenu(s))
		// pprof.StopCPUProfile()
		// pf.Close()
		s.Fini()
		os.Exit(0)
	}()
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
	s.Fini()
	s.Quit()
}
