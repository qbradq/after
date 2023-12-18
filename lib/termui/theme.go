package termui

import "github.com/gdamore/tcell/v2"

// The default theme
var DefaultTheme = Theme{
	Normal: tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite),
	Highlight: tcell.StyleDefault.
		Background(tcell.ColorNavy).
		Foreground(tcell.ColorWhite),
	Error: tcell.StyleDefault.
		Background(tcell.ColorRed).
		Foreground(tcell.ColorWhite),
}

// The current theme
var CurrentTheme Theme = DefaultTheme

// Theme defines the colors used for various standard things.
type Theme struct {
	Normal    tcell.Style // Normal text
	Highlight tcell.Style // Highlighted text
	Error     tcell.Style // Error messages
}
