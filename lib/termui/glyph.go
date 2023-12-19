package termui

// Glyph represents the content of one cell in the terminal.
type Glyph struct {
	Rune  rune  // Rune to display
	Style Style // Display style
}
