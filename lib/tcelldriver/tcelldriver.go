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
	s tcell.Screen
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
	return nil
}

// Fini must be called to close the driver.
func (d *Driver) Fini() {
	d.s.Fini()
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
	for {
		e := d.s.PollEvent()
		switch ev := e.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyRune:
				return &termui.EventKey{Key: ev.Rune()}
			case tcell.KeyEnter:
				return &termui.EventKey{Key: '\n'}
			case tcell.KeyEsc:
				return &termui.EventKey{Key: '\033'}
			}
		case *tcell.EventResize:
			w, h := ev.Size()
			return &termui.EventResize{
				Size: util.Point{
					X: w,
					Y: h,
				},
			}
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
