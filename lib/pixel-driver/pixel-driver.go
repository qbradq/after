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
	quit    bool            // If true PollEvent will always return *EventQuit
	lock    sync.RWMutex    // Mutex for fb
	b       util.Rect       // Bounds of the screen
	events  chan any        // Events channel
	fbDirty bool            // If true, fb needs to be re-drawn
	fb      []termui.Glyph  // Front buffer
	bb      []termui.Glyph  // Back buffer
	icon    pixel.Picture   // Icon image
	font    pixel.Picture   // Font backing image
	glyphs  []*pixel.Sprite // Cache of glyph images
}

// NewDriver returns a new Driver ready for use.
func NewDriver() *Driver {
	icon, _, err := image.Decode(bytes.NewReader(iconData))
	if err != nil {
		return nil
	}
	baseFont, _, err := image.Decode(bytes.NewReader(fontData))
	if err != nil {
		return nil
	}
	w := baseFont.Bounds().Dx()
	h := baseFont.Bounds().Dy()
	font := image.NewRGBA(image.Rect(0, 0, w, h*16))
	for sy := 0; sy < h; sy++ {
		for sx := 0; sx < w; sx++ {
			sc := baseFont.At(sx, sy)
			if _, _, _, a := sc.RGBA(); a == 0 {
				continue
			}
			for i, c := range termui.Palette {
				font.Set(sx, i*h+sy, c)
			}
		}
	}
	d := &Driver{
		events: make(chan any, 128),
		icon:   pixel.PictureDataFromImage(icon),
		font:   pixel.PictureDataFromImage(font),
		glyphs: make([]*pixel.Sprite, 128*16),
	}
	for iy := 0; iy < 8*16; iy++ {
		for ix := 0; ix < 16; ix++ {
			sx := ix * 8
			sy := int(d.font.Bounds().H()) - (iy*8 + 8)
			d.glyphs[iy*16+ix] = pixel.NewSprite(d.font,
				pixel.R(float64(sx), float64(sy), float64(sx+8), float64(sy+8)))
		}
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
	d.bb = make([]termui.Glyph, b.Width()*b.Height())
	d.lock.Lock()
	d.fb = make([]termui.Glyph, b.Width()*b.Height())
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
	var m pixel.Matrix
	_, bh := d.Size()
	sh := bh * 8 * 2
	b := pixel.NewBatch(&pixel.TrianglesData{}, d.font)
	d.lock.RLock()
	for p.Y = d.b.TL.Y; p.Y <= d.b.BR.Y; p.Y++ {
		for p.X = d.b.TL.X; p.X <= d.b.BR.X; p.X++ {
			g := d.fb[p.Y*d.b.Width()+p.X]
			fg, bg := g.Style.Decompose()
			n := int(g.Rune)
			if n > 126 {
				n = 126
			}
			n += int(fg) * 128
			char := d.glyphs[n]
			block := d.glyphs[int(bg)*128+127]
			m = pixel.IM.
				Scaled(pixel.ZV, 2).
				Moved(pixel.V(float64(p.X*8*2+8), float64((sh-p.Y*8*2)-8)))
			block.Draw(b, m)
			char.Draw(b, m)
		}
	}
	d.lock.RUnlock()
	b.Draw(t)
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
	d.bb[p.Y*d.b.Width()+p.X] = g
}

// GetCell implements the termui.TerminalDriver interface.
func (d *Driver) GetCell(p util.Point) termui.Glyph {
	return d.bb[p.Y*d.b.Width()+p.X]
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
