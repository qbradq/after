package main

import (
	"os"
	"os/signal"
	"runtime/pprof"

	"github.com/gopxl/pixel/pixelgl"
	"github.com/qbradq/after/internal/client/termgui"
	pixeldriver "github.com/qbradq/after/lib/pixel-driver"
	"github.com/qbradq/after/lib/termui"
)

func main() {
	s := pixeldriver.New()
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
		pf, err := os.Create("heap.pprof")
		if err != nil {
			panic(err)
		}
		pprof.WriteHeapProfile(pf)
		pf.Close()
		s.Fini()
		os.Exit(0)
	}()
	pixelgl.Run(s.Run)
	s.Fini()
	os.Exit(0)
}
