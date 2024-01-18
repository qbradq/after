// Package commondriver implements part of the [termui.TerminalDriver] interface
// that is common between all drivers.
package commondriver

import "github.com/qbradq/after/lib/termui"

// Driver implements part of the [termui.TerminalDriver] interface that is
// common between all drives.
type Driver struct {
	Done   bool     // If true PollEvent will always return *EventQuit
	Events chan any // Events channel
}

// New returns a new CommonDriver ready for use.
func New() *Driver {
	return &Driver{
		Events: make(chan any, 128),
	}
}

// Quit implements the termui.TerminalDriver interface.
func (d *Driver) Quit() {
	d.Done = true
}

// PollEvent implements the termui.TerminalDriver interface.
func (d *Driver) PollEvent() any {
	if d.Done {
		return &termui.EventQuit{}
	}
	return <-d.Events
}

// FlushEvents implements the termui.TerminalDriver interface.
func (d *Driver) FlushEvents() {
	var keep []any
	var done bool
	if d.Done {
		d.Events <- &termui.EventQuit{}
		return
	}
	for !done {
		select {
		case ev := <-d.Events:
			switch ev.(type) {
			case *termui.EventResize:
				keep = append(keep, ev)
			case *termui.EventQuit:
				keep = append(keep, ev)
			case *termui.EventKey:
				// Discard event
			}
		default:
			// Out of events on the channel, re-populate with any system
			// messages we extracted and return
			for _, ev := range keep {
				d.Events <- ev
			}
			done = true
		}
	}
}
