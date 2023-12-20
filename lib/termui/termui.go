package termui

import (
	"errors"
	"log"

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
	// PollEvent returns the next event.
	PollEvent() any
	// Size returns the size of the screen.
	Size() (int, int)
	// Sync must redraw the entire terminal display.
	Sync()
	// Show must redraw the dirty areas of the terminal display.
	Show()
}

// Mode is the interface all objects implementing a game mode must implement.
type Mode interface {
	// HandleEvent is responsible for updating the mode's state in response to
	// events. If ErrorQuit is returned the mode is exited. Any other error is
	// treated as fatal.
	HandleEvent(TerminalDriver, any) error
	// Draw is responsible for drawing the mode to the terminal driver.
	Draw(TerminalDriver)
}

// RunMode runs the given mode.
func RunMode(s TerminalDriver, m Mode) {
	m.Draw(s)
	s.Show()
	for {
		e := s.PollEvent()
		switch e.(type) {
		case *EventQuit:
			return
		case *EventResize:
			m.Draw(s)
			s.Sync()
		default:
			if err := m.HandleEvent(s, e); err != nil {
				if errors.Is(err, ErrorQuit) {
					return
				}
				log.Fatal(err)
			}
			m.Draw(s)
			s.Show()
		}
	}
}
