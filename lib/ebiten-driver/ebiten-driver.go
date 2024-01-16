// Package ebitendriver implements a [termui.TerminalDriver] with
// a graphical terminal emulator using Ebitengine as the backing library.
package ebitendriver

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

//go:embed font.png
var fontData []byte

// Driver implements a termui.TerminalDriver that uses the Pixel library
// to implement a graphical terminal emulator with a square font.
type Driver struct {
	quit   bool            // If true PollEvent will always return *EventQuit
	lock   sync.RWMutex    // Mutex for fb
	b      util.Rect       // Bounds of the screen
	events chan any        // Events channel
	fb     []termui.Glyph  // Front buffer
	bb     []termui.Glyph  // Back buffer
	font   *ebiten.Image   // Font backing image
	glyphs []*ebiten.Image // Cache of glyph images
}

// NewDriver returns a new Driver ready for use.
func NewDriver() *Driver {
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
		font:   ebiten.NewImageFromImage(font),
		glyphs: make([]*ebiten.Image, 128*16),
	}
	for iy := 0; iy < 8*16; iy++ {
		for ix := 0; ix < 16; ix++ {
			sx := ix * 8
			sy := iy * 8
			d.glyphs[iy*16+ix] = d.font.SubImage(image.Rect(sx, sy, sx+8, sy+8)).(*ebiten.Image)
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

// Update implements the ebiten.Game interface.
func (d *Driver) Update() error {
	in := make([]rune, 0, 32)
	in = ebiten.AppendInputChars(in)
	for _, r := range in {
		d.events <- &termui.EventKey{Key: r}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) {
		d.events <- &termui.EventKey{Key: '\n'}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		d.events <- &termui.EventKey{Key: '\033'}
	}
	return nil
}

// Draw implements the ebiten.Game interface.
func (d *Driver) Draw(s *ebiten.Image) {
	var op ebiten.DrawImageOptions
	d.lock.RLock()
	var p util.Point
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
			op.GeoM.Reset()
			op.GeoM.Translate(float64(p.X*8), float64(p.Y*8))
			s.DrawImage(block, &op)
			s.DrawImage(char, &op)
		}
	}
	d.lock.RUnlock()
}

// Layout implements the ebiten.Game interface.
func (d *Driver) Layout(w, h int) (int, int) {
	z := 2
	w /= z
	h /= z
	d.resize(util.NewRectWH(w/8, h/8))
	return w, h
}

// Quit sets the quit flag so the application can exit gracefully.
func (d *Driver) Quit() {
	d.quit = true
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
}

// Show implements the termui.TerminalDriver interface.
func (d *Driver) Show() {
	d.Sync()
}
