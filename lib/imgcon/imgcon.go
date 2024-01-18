// Package imgcon implements a console renderer that renders its buffer to an
// [image.Image].
package imgcon

import (
	"bytes"
	_ "embed"
	"image"
	"sync"

	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

//go:embed font.png
var fontData []byte

// Console implements a console renderer that renders its buffer to an
// [image.Image].
type Console struct {
	FBDirty bool            // If true, fb needs to be re-drawn
	lock    sync.RWMutex    // Mutex for fb
	b       util.Rect       // Bounds of the screen
	fb      []termui.Glyph  // Front buffer
	bb      []termui.Glyph  // Back buffer
	font    *image.Paletted // Font backing image
	screen  *image.Paletted // Current screen image updated with every call to [ImageConsole.Draw]
}

// New returns a new Console ready for use.
func New(b util.Rect) *Console {
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
	// Create the image console
	ret := &Console{
		b:    b,
		font: font,
	}
	ret.Resize(b)
	return ret
}

// Resize resizes the console buffers to the given size.
func (c *Console) Resize(b util.Rect) {
	if c.fb != nil && c.b == b {
		return
	}
	c.b = b
	c.lock.Lock()
	c.bb = make([]termui.Glyph, b.Width()*b.Height())
	c.fb = make([]termui.Glyph, b.Width()*b.Height())
	c.screen = image.NewPaletted(image.Rect(0, 0, b.Width()*8, b.Height()*8), termui.Palette)
	c.lock.Unlock()
}

// Draw returns a pointer to the newly updated screen image.
func (c *Console) Draw() image.Image {
	var p util.Point
	c.lock.RLock()
	for p.Y = c.b.TL.Y; p.Y <= c.b.BR.Y; p.Y++ {
		for p.X = c.b.TL.X; p.X <= c.b.BR.X; p.X++ {
			g := c.fb[p.Y*c.b.Width()+p.X]
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
					s := c.font.Pix[(sy+iy)*128+sx+ix]
					if s == 0 {
						c.screen.Pix[dy*c.screen.Stride+dx] = uint8(bg)
					} else {
						c.screen.Pix[dy*c.screen.Stride+dx] = uint8(fg)
					}
				}
			}
		}
	}
	c.lock.RUnlock()
	return c.screen
}

// SetCell implements the termui.TerminalDriver interface.
func (c *Console) SetCell(p util.Point, g termui.Glyph) {
	if !c.b.Contains(p) {
		return
	}
	c.lock.Lock()
	c.bb[p.Y*c.b.Width()+p.X] = g
	c.lock.Unlock()
}

// GetCell implements the termui.TerminalDriver interface.
func (c *Console) GetCell(p util.Point) termui.Glyph {
	c.lock.RLock()
	g := c.bb[p.Y*c.b.Width()+p.X]
	c.lock.RUnlock()
	return g
}

// Size implements the termui.TerminalDriver interface.
func (c *Console) Size() (int, int) {
	return c.b.Width(), c.b.Height()
}

// Sync implements the termui.TerminalDriver interface.
func (c *Console) Sync() {
	c.lock.Lock()
	copy(c.fb, c.bb)
	c.lock.Unlock()
	c.FBDirty = true
}

// Show implements the termui.TerminalDriver interface.
func (d *Console) Show() {
	d.Sync()
}
