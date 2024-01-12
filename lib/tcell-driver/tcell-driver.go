package tcelldriver

import (
	"github.com/gdamore/tcell/v2"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// Mapping of termui.Color values to tcell.Color values
var colors = []tcell.Color{
	tcell.ColorBlack,
	tcell.ColorMaroon,
	tcell.ColorGreen,
	tcell.ColorOlive,
	tcell.ColorNavy,
	tcell.ColorPurple,
	tcell.ColorTeal,
	tcell.ColorSilver,
	tcell.ColorGray,
	tcell.ColorRed,
	tcell.ColorLime,
	tcell.ColorYellow,
	tcell.ColorBlue,
	tcell.ColorFuchsia,
	tcell.ColorAqua,
	tcell.ColorWhite,
}

// Mapping of tcell.Color values to termui.Color values
var colorBackRef = map[tcell.Color]termui.Color{
	tcell.ColorBlack:   termui.ColorBlack,
	tcell.ColorMaroon:  termui.ColorMaroon,
	tcell.ColorGreen:   termui.ColorGreen,
	tcell.ColorOlive:   termui.ColorOlive,
	tcell.ColorNavy:    termui.ColorNavy,
	tcell.ColorPurple:  termui.ColorPurple,
	tcell.ColorTeal:    termui.ColorTeal,
	tcell.ColorSilver:  termui.ColorSilver,
	tcell.ColorGray:    termui.ColorGray,
	tcell.ColorRed:     termui.ColorRed,
	tcell.ColorLime:    termui.ColorLime,
	tcell.ColorYellow:  termui.ColorYellow,
	tcell.ColorBlue:    termui.ColorBlue,
	tcell.ColorFuchsia: termui.ColorFuchsia,
	tcell.ColorAqua:    termui.ColorAqua,
	tcell.ColorWhite:   termui.ColorWhite,
}

// Driver is the termui.Driver implementation over tcell.
type Driver struct {
	s      tcell.Screen
	quit   chan struct{}
	events chan any
}

// Init must be called to initialize the driver.
func (d *Driver) Init() error {
	var err error
	d.s, err = tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := d.s.Init(); err != nil {
		return err
	}
	d.s.Clear()
	d.quit = make(chan struct{})
	d.events = make(chan any, 1024)
	go pumpEvents(d)
	return nil
}

// Fini must be called to close the driver.
func (d *Driver) Fini() {
	d.quit <- struct{}{}
	d.s.Fini()
}

func pumpEvents(d *Driver) {
	for {
		select {
		case <-d.quit:
			d.events <- &termui.EventQuit{}
			return
		default:
			e := d.s.PollEvent()
			switch ev := e.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyRune:
					d.events <- &termui.EventKey{Key: ev.Rune()}
				case tcell.KeyEnter:
					d.events <- &termui.EventKey{Key: '\n'}
				case tcell.KeyEsc:
					d.events <- &termui.EventKey{Key: '\033'}
				}
			case *tcell.EventResize:
				w, h := ev.Size()
				d.events <- &termui.EventResize{
					Size: util.Point{
						X: w,
						Y: h,
					},
				}
			}
		}
	}
}

// Size implements the termui.TerminalDriver interface.
func (d *Driver) Size() (w, h int) {
	return d.s.Size()
}

// SetCell implements the termui.TerminalDriver interface.
func (d *Driver) SetCell(p util.Point, g termui.Glyph) {
	fg, bg := g.Style.Decompose()
	ns := tcell.StyleDefault.Foreground(colors[fg]).Background(colors[bg])
	d.s.SetContent(p.X, p.Y, g.Rune, nil, ns)
}

// GetCell implements the termui.TerminalDriver interface.
func (d *Driver) GetCell(p util.Point) termui.Glyph {
	r, _, s, _ := d.s.GetContent(p.X, p.Y)
	fg, bg, _ := s.Decompose()
	return termui.Glyph{
		Rune: r,
		Style: termui.StyleDefault.
			Background(colorBackRef[bg]).
			Foreground(colorBackRef[fg]),
	}
}

// PollEvent implements the termui.TerminalDriver interface.
func (d *Driver) PollEvent() any {
	return <-d.events
}

// FlushEvents implements the termui.TerminalDriver interface.
func (d *Driver) FlushEvents() {
	var keep []any
	var done bool
	for !done {
		select {
		case ev := <-d.events:
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
				d.events <- ev
			}
			done = true
		}
	}
}

// Sync implements the termui.TerminalDriver interface.
func (d *Driver) Sync() {
	d.s.Sync()
}

// Show implements the termui.TerminalDriver interface.
func (d *Driver) Show() {
	d.s.Show()
}
