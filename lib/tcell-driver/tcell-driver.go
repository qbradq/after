// Package tcelldriver implements a [termui.TerminalDriver] for a tty
// using tcell as the backing library.
package tcelldriver

import (
	"github.com/gdamore/tcell/v2"
	commondriver "github.com/qbradq/after/lib/common-driver"
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
	commondriver.Driver
	s    tcell.Screen  // TCell backing screen
	quit chan struct{} // Quit event channel
}

// New returns a new Driver ready for use.
func New() *Driver {
	return &Driver{
		Driver: *commondriver.New(),
		quit:   make(chan struct{}),
	}
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
	go pumpEvents(d)
	return nil
}

// Fini must be called to close the driver.
func (d *Driver) Fini() {
	d.Quit()
	d.s.Fini()
}

// Quit implements the termui.Driver interface.
func (d *Driver) Quit() {
	d.Driver.Quit()
	d.quit <- struct{}{}
}

func pumpEvents(d *Driver) {
	for {
		select {
		case <-d.quit:
			d.Driver.Events <- &termui.EventQuit{}
			return
		default:
			e := d.s.PollEvent()
			switch ev := e.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyRune:
					d.Driver.Events <- &termui.EventKey{Key: ev.Rune()}
				case tcell.KeyEnter:
					d.Driver.Events <- &termui.EventKey{Key: '\n'}
				case tcell.KeyBackspace:
					d.Driver.Events <- &termui.EventKey{Key: '\010'}
				case tcell.KeyEsc:
					d.Driver.Events <- &termui.EventKey{Key: '\033'}
				}
			case *tcell.EventResize:
				w, h := ev.Size()
				d.Driver.Events <- &termui.EventResize{
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

// Sync implements the termui.TerminalDriver interface.
func (d *Driver) Sync() {
	d.s.Sync()
}

// Show implements the termui.TerminalDriver interface.
func (d *Driver) Show() {
	d.s.Show()
}
