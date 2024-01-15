package pixeldriver

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"sync"
	"time"

	"github.com/gopxl/pixel"
	"github.com/gopxl/pixel/pixelgl"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

//go:embed font.png
var fontData []byte

//go:embed icon.png
var iconData []byte

// Driver implements a termui.TerminalDriver that uses the Pixel library
// to implement a graphical terminal emulator with a square font.
type Driver struct {
	quit    bool                 // If true PollEvent will always return *EventQuit
	lock    sync.RWMutex         // Mutex for fb
	b       util.Rect            // Bounds of the screen
	events  chan any             // Events channel
	fbDirty bool                 // If true, fb needs to be re-drawn
	fb      []termui.Glyph       // Front buffer
	bb      []termui.Glyph       // Back buffer
	font    *image.Paletted      // Font backing image
	screen  *image.Paletted      // Rendering surface
	icon    pixel.Picture        // Icon image
	td      *pixel.TrianglesData // TrianglesData cache to avoid constant allocation
}

// NewDriver returns a new Driver ready for use.
func NewDriver() *Driver {
	// Load icon
	icon, _, err := image.Decode(bytes.NewReader(iconData))
	if err != nil {
		panic(err)
	}
	// Load base font
	baseFont, _, err := image.Decode(bytes.NewReader(fontData))
	if err != nil {
		panic(err)
	}
	// Generate font backing image
	w := baseFont.Bounds().Dx()
	h := baseFont.Bounds().Dy()
	font := image.NewPaletted(image.Rect(0, 0, w, h), termui.Palette)
	for sy := 0; sy < h; sy++ {
		for sx := 0; sx < w; sx++ {
			sc := baseFont.At(sx, sy)
			if _, _, _, a := sc.RGBA(); a == 0 {
				font.SetColorIndex(sx, sy, 0)
			} else {
				font.SetColorIndex(sx, sy, 1)
			}
		}
	}
	// Driver
	d := &Driver{
		events: make(chan any, 128),
		icon:   pixel.PictureDataFromImage(icon),
		font:   font,
		td:     &pixel.TrianglesData{},
	}
	d.resize(util.NewRectWH(80, 50))
	return d
}

// resize attempts to resize the driver's internal buffers if needed.
func (d *Driver) resize(b util.Rect) {
	if d.fb != nil && d.b == b {
		return
	}
	d.b = b
	d.lock.Lock()
	d.bb = make([]termui.Glyph, b.Width()*b.Height())
	d.fb = make([]termui.Glyph, b.Width()*b.Height())
	d.screen = image.NewPaletted(image.Rect(0, 0, b.Width()*8, b.Height()*8), termui.Palette)
	d.lock.Unlock()
	d.events <- &termui.EventResize{
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
	d.fbDirty = true
	for !win.Closed() && !d.quit {
		// Refresh display
		// if d.fbDirty {
		start := time.Now()
		d.draw(win)
		win.SwapBuffers()
		util.Log("PixelDriver.draw: %dms", time.Since(start).Milliseconds())
		d.fbDirty = false
		// }
		// Wait for input
		win.UpdateInput()
		// Collect events
		for _, r := range win.Typed() {
			d.events <- &termui.EventKey{Key: r}
			d.fbDirty = true
		}
		if win.JustPressed(pixelgl.KeyEnter) {
			d.events <- &termui.EventKey{Key: '\n'}
			d.fbDirty = true
		}
		if win.JustPressed(pixelgl.KeyEscape) {
			d.events <- &termui.EventKey{Key: '\033'}
			d.fbDirty = true
		}
		// Give up time slice
		time.Sleep(time.Millisecond * 33)
	}
}

func (d *Driver) draw(t pixel.Target) {
	var p util.Point
	d.td.SetLen(0)
	d.lock.RLock()
	for p.Y = d.b.TL.Y; p.Y <= d.b.BR.Y; p.Y++ {
		for p.X = d.b.TL.X; p.X <= d.b.BR.X; p.X++ {
			g := d.fb[p.Y*d.b.Width()+p.X]
			fg, bg := g.Style.Decompose()
			n := int(g.Rune)
			if n > 126 {
				n = 126
			}
			sx := (n % 16) * 8
			sy := (n / 16) * 8
			for iy := 0; iy < 8; iy++ {
				dy := (p.Y * 8) + iy
				for ix := 0; ix < 8; ix++ {
					dx := (p.X * 8) + ix
					s := d.font.Pix[(sy+iy)*128+sx+ix]
					if s == 0 {
						d.screen.Pix[dy*d.screen.Stride+dx] = uint8(bg)
					} else {
						d.screen.Pix[dy*d.screen.Stride+dx] = uint8(fg)
					}
				}
			}
		}
	}
	d.lock.RUnlock()
	pic := pixel.PictureDataFromImage(d.screen)
	s := pixel.NewSprite(pic, pic.Bounds())
	s.Draw(t, pixel.IM.
		Scaled(pixel.ZV, 2).
		Moved(pixel.V(pic.Bounds().W(), pic.Bounds().H())))
}

// Quit sets the quit flag so the application can exit gracefully.
func (d *Driver) Quit() {
	d.quit = true
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

// SetCell implements the termui.TerminalDriver interface.
func (d *Driver) SetCell(p util.Point, g termui.Glyph) {
	if !d.b.Contains(p) {
		return
	}
	d.lock.Lock()
	d.bb[p.Y*d.b.Width()+p.X] = g
	d.lock.Unlock()
}

// GetCell implements the termui.TerminalDriver interface.
func (d *Driver) GetCell(p util.Point) termui.Glyph {
	d.lock.RLock()
	g := d.bb[p.Y*d.b.Width()+p.X]
	d.lock.RUnlock()
	return g
}

// PollEvent implements the termui.TerminalDriver interface.
func (d *Driver) PollEvent() any {
	if d.quit {
		return &termui.EventQuit{}
	}
	return <-d.events
}

// FlushEvents implements the termui.TerminalDriver interface.
func (d *Driver) FlushEvents() {
	var keep []any
	var done bool
	if d.quit {
		d.events <- &termui.EventQuit{}
		return
	}
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

// Size implements the termui.TerminalDriver interface.
func (d *Driver) Size() (int, int) {
	return d.b.Width(), d.b.Height()
}

// Sync implements the termui.TerminalDriver interface.
func (d *Driver) Sync() {
	d.lock.Lock()
	copy(d.fb, d.bb)
	d.lock.Unlock()
	d.fbDirty = true
}

// Show implements the termui.TerminalDriver interface.
func (d *Driver) Show() {
	d.Sync()
}
