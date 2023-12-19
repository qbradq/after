package termui

import (
	"errors"

	"github.com/qbradq/after/lib/util"
)

// Error that indicates we just want to quit the Mode function
var ErrorQuit = errors.New("quit")

// TerminalDriver is the interface terminal front-end implementations must
// implement to use termui.
type TerminalDriver interface {
	// Init is responsible for all initialization operations.
	Init() error
	// Fini is responsible for all cleanup operations.
	Fini()
	// SetCell sets a single cell of the screen.
	SetCell(util.Point, Glyph)
	// GetCell returns the glyph at the given point on the screen.
	GetCell(util.Point) Glyph
	// Size returns the size of the screen.
	Size() (int, int)
	// PollEvent returns the next input or system event.
	PollEvent() any
	// Sync must redraw the entire terminal display.
	Sync()
	// Show must redraw the dirty areas of the terminal display.
	Show()
}

// Mode is a function that accepts the next input event and then re-draws the
// screen. If error is non-null the runner will exit the mode. See Run().
type Mode func(any) error

// Run runs the given mode function.
func Run(s TerminalDriver, fn Mode) {
	var event any
	if err := fn(nil); err != nil {
		return
	}
	s.Show()
	for {
		event = s.PollEvent()
		switch event.(type) {
		case *EventResize:
			if err := fn(event); err != nil {
				return
			}
			s.Sync()
			continue
		default:
		}
		if err := fn(event); err != nil {
			return
		}
		s.Show()
	}
}
