// Package pixeldriver implements a [termui.TerminalDriver] with
// a graphical terminal emulator using Pixel2 as the backing library.
package pixeldriver

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"time"

	"github.com/gopxl/pixel"
	"github.com/gopxl/pixel/pixelgl"
	commondriver "github.com/qbradq/after/lib/common-driver"
	"github.com/qbradq/after/lib/imgcon"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

//go:embed icon.png
var iconData []byte

// Driver implements a termui.TerminalDriver that uses the Pixel library
// to implement a graphical terminal emulator with a square font.
type Driver struct {
	imgcon.Console
	commondriver.Driver
	icon pixel.Picture        // Icon image
	td   *pixel.TrianglesData // TrianglesData cache to avoid constant allocation
}

// New returns a new Driver ready for use.
func New() *Driver {
	// Load icon
	icon, _, err := image.Decode(bytes.NewReader(iconData))
	if err != nil {
		panic(err)
	}
	// Driver
	b := util.NewRectWH(80, 50)
	d := &Driver{
		Console: *imgcon.New(b),
		Driver:  *commondriver.New(),
		icon:    pixel.PictureDataFromImage(icon),
		td:      &pixel.TrianglesData{},
	}
	d.resize(b)
	return d
}

// resize attempts to resize the driver's internal buffers if needed.
func (d *Driver) resize(b util.Rect) {
	d.Console.Resize(b)
	d.Events <- &termui.EventResize{
		Size: util.Point{
			X: b.Width(),
			Y: b.Height(),
		},
	}
}

// Run implements the pixel main loop.
func (d *Driver) Run() {
	cfg := pixelgl.WindowConfig{
		Title:     "After",
		Icon:      []pixel.Picture{d.icon},
		Bounds:    pixel.R(0, 0, 80*8*2, 50*8*2),
		Resizable: true,
		VSync:     true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	// Main program loop
	d.Console.FBDirty = true
	for !win.Closed() && !d.Driver.Done {
		// Refresh display
		if d.Console.FBDirty {
			d.draw(win)
			win.SwapBuffers()
			d.Console.FBDirty = false
		}
		// Wait for input
		win.UpdateInput()
		// Collect events
		for _, r := range win.Typed() {
			d.Driver.Events <- &termui.EventKey{Key: r}
			d.Console.FBDirty = true
		}
		if win.JustPressed(pixelgl.KeyEnter) {
			d.Driver.Events <- &termui.EventKey{Key: '\n'}
			d.Console.FBDirty = true
		}
		if win.JustPressed(pixelgl.KeyEscape) {
			d.Driver.Events <- &termui.EventKey{Key: '\033'}
			d.Console.FBDirty = true
		}
		// Give up time slice
		time.Sleep(time.Millisecond * 33)
	}
}

func (d *Driver) draw(t pixel.Target) {
	d.td.SetLen(0)
	screen := d.Console.Draw()
	pic := pixel.PictureDataFromImage(screen)
	s := pixel.NewSprite(pic, pic.Bounds())
	s.Draw(t, pixel.IM.
		Scaled(pixel.ZV, 2).
		Moved(pixel.V(pic.Bounds().W(), pic.Bounds().H())))
}

// Init implements the termui.TerminalDriver interface.
func (d *Driver) Init() error {
	// Nothing to do
	return nil
}

// Fini implements the termui.TerminalDriver interface.
func (d *Driver) Fini() {
	d.Quit()
}
