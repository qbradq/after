package termui

const (
	styleFGMask  Style = 0x00FF
	styleFGShift       = 0
	styleBGMask  Style = 0xFF00
	styleBGShift       = 8
)

// Style defines the appearance of a glyph.
type Style uint16

// Default terminal style
var StyleDefault = Style(0).Foreground(ColorWhite).Background(ColorBlack)

// Foreground returns the style with the foreground set.
func (s Style) Foreground(c Color) Style {
	return (s & Style(styleBGMask)) | (Style(c) << styleFGShift)
}

// Background returns the style with the background set.
func (s Style) Background(c Color) Style {
	return (s & Style(styleFGMask)) | (Style(c) << styleBGShift)
}

// Decompose returns the colors of the style.
func (s Style) Decompose() (fg, bg Color) {
	fg = Color((s & styleFGMask) >> styleFGShift)
	bg = Color((s & styleBGMask) >> styleBGShift)
	return fg, bg
}
