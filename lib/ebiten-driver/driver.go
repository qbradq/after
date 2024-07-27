// Package ebitendriver implements a [termui.TerminalDriver] with
// a graphical terminal emulator using Ebitengine as the backing library.
package ebitendriver

import (
	_ "embed"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	commondriver "github.com/qbradq/after/lib/common-driver"
	"github.com/qbradq/after/lib/imgcon"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// Driver implements a termui.TerminalDriver that uses the Pixel library
// to implement a graphical terminal emulator with a square font.
type Driver struct {
	imgcon.Console
	commondriver.Driver
}

// New returns a new Driver ready for use.
func New() *Driver {
	b := util.NewRectWH(80, 50)
	d := &Driver{
		Console: *imgcon.New(b),
		Driver:  *commondriver.New(),
	}
	d.resize(b)
	return d
}

// resize attempts to resize the driver's internal buffers if needed.
func (d *Driver) resize(b util.Rect) {
	d.Console.Resize(b)
	d.Driver.Events <- &termui.EventResize{
		Size: util.Point{
			X: b.Width(),
			Y: b.Height(),
		},
	}
}

// Update implements the ebiten.Game interface.
func (d *Driver) Update() error {
	in := make([]rune, 0, 32)
	in = ebiten.AppendInputChars(in)
	for _, r := range in {
		d.Driver.Events <- &termui.EventKey{Key: r}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) {
		d.Driver.Events <- &termui.EventKey{Key: '\n'}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		d.Driver.Events <- &termui.EventKey{Key: '\010'}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		d.Driver.Events <- &termui.EventKey{Key: '\033'}
	}
	return nil
}

// Draw implements the ebiten.Game interface.
func (d *Driver) Draw(s *ebiten.Image) {
	sImg := d.Console.Draw()
	screen := ebiten.NewImageFromImage(sImg)
	s.DrawImage(screen, &ebiten.DrawImageOptions{})
}

// Layout implements the ebiten.Game interface.
func (d *Driver) Layout(w, h int) (int, int) {
	z := 2
	w /= z
	h /= z
	d.resize(util.NewRectWH(w/8, h/8))
	return w, h
}

// Init implements the termui.TerminalDriver interface.
func (d *Driver) Init() error {
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetRunnableOnUnfocused(false)
	ebiten.SetWindowSize(80*8*2, 50*8*2)
	ebiten.SetWindowTitle("After")
	return nil
}

// Fini implements the termui.TerminalDriver interface.
func (d *Driver) Fini() {
	d.Quit()
}
