package main

import (
	"os"
	"os/signal"

	"github.com/qbradq/after/internal/client/termgui"
	"github.com/qbradq/after/lib/tcelldriver"
	"github.com/qbradq/after/lib/termui"
)

func main() {
	s := &tcelldriver.Driver{}
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
	termui.RunMode(s, termgui.NewMainMenu())
}
