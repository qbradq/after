package termui

// The default theme
var DefaultTheme = Theme{
	Normal: StyleDefault.
		Background(ColorBlack).
		Foreground(ColorWhite),
	Highlight: StyleDefault.
		Background(ColorNavy).
		Foreground(ColorWhite),
	Error: StyleDefault.
		Background(ColorRed).
		Foreground(ColorWhite),
}

// The current theme
var CurrentTheme Theme = DefaultTheme

// Theme defines the colors used for various standard things.
type Theme struct {
	Normal    Style // Normal text
	Highlight Style // Highlighted text
	Error     Style // Error messages
}
