package main

import (
	"os"
	"os/signal"

	"github.com/qbradq/after/internal/client/termgui"
	tcelldriver "github.com/qbradq/after/lib/tcell-driver"
	"github.com/qbradq/after/lib/termui"
)

func main() {
	s := tcelldriver.New()
	s.Init()
	defer s.Fini()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			s.Fini()
			os.Exit(0)
		}
	}()
	// pf, err := os.Create("cpu.pprof")
	// if err != nil {
	// 	panic(err)
	// }
	// pprof.StartCPUProfile(pf)
	termui.RunMode(s, termgui.NewMainMenu(s))
	// pprof.StopCPUProfile()
	// pf.Close()
}
