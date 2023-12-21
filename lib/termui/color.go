package termui

import (
	"fmt"
	"image/color"
	"strings"
)

// Color is a wrapper value for tcell.Color.
type Color uint8

const (
	ColorBlack Color = iota
	ColorMaroon
	ColorGreen
	ColorOlive
	ColorNavy
	ColorPurple
	ColorTeal
	ColorSilver
	ColorGray
	ColorRed
	ColorLime
	ColorYellow
	ColorBlue
	ColorFuchsia
	ColorAqua
	ColorWhite
)

// Built-in palette
var Palette = []color.Color{
	color.RGBA{R: 0, G: 0, B: 0, A: 255},
	color.RGBA{R: 170, G: 0, B: 0, A: 255},
	color.RGBA{R: 0, G: 170, B: 0, A: 255},
	color.RGBA{R: 170, G: 85, B: 0, A: 255},
	color.RGBA{R: 0, G: 0, B: 170, A: 255},
	color.RGBA{R: 170, G: 0, B: 170, A: 255},
	color.RGBA{R: 0, G: 170, B: 170, A: 255},
	color.RGBA{R: 170, G: 170, B: 170, A: 255},
	color.RGBA{R: 85, G: 85, B: 85, A: 255},
	color.RGBA{R: 255, G: 85, B: 85, A: 255},
	color.RGBA{R: 85, G: 255, B: 85, A: 255},
	color.RGBA{R: 255, G: 255, B: 85, A: 255},
	color.RGBA{R: 85, G: 85, B: 255, A: 255},
	color.RGBA{R: 255, G: 85, B: 255, A: 255},
	color.RGBA{R: 85, G: 255, B: 255, A: 255},
	color.RGBA{R: 255, G: 255, B: 255, A: 255},
}

func (c Color) MarshalJSON() ([]byte, error) {
	switch c {
	case ColorBlack:
		return []byte([]byte("Black")), nil
	case ColorMaroon:
		return []byte([]byte("Maroon")), nil
	case ColorGreen:
		return []byte([]byte("Green")), nil
	case ColorOlive:
		return []byte([]byte("Olive")), nil
	case ColorNavy:
		return []byte([]byte("Navy")), nil
	case ColorPurple:
		return []byte([]byte("Purple")), nil
	case ColorTeal:
		return []byte([]byte("Teal")), nil
	case ColorSilver:
		return []byte([]byte("Silver")), nil
	case ColorGray:
		return []byte([]byte("Gray")), nil
	case ColorRed:
		return []byte([]byte("Red")), nil
	case ColorLime:
		return []byte([]byte("Lime")), nil
	case ColorYellow:
		return []byte([]byte("Yellow")), nil
	case ColorBlue:
		return []byte([]byte("Blue")), nil
	case ColorFuchsia:
		return []byte([]byte("Fuchsia")), nil
	case ColorAqua:
		return []byte([]byte("Aqua")), nil
	case ColorWhite:
		return []byte([]byte("White")), nil
	}
	return nil, fmt.Errorf("unknown color code %d", c)
}

func (c *Color) UnmarshalJSON(in []byte) error {
	switch strings.ToLower(string(in[1 : len(in)-1])) {
	case "black":
		*c = ColorBlack
	case "maroon":
		*c = ColorMaroon
	case "green":
		*c = ColorGreen
	case "olive":
		*c = ColorOlive
	case "navy":
		*c = ColorNavy
	case "purple":
		*c = ColorPurple
	case "teal":
		*c = ColorTeal
	case "silver":
		*c = ColorSilver
	case "gray":
		*c = ColorGray
	case "red":
		*c = ColorRed
	case "lime":
		*c = ColorLime
	case "yellow":
		*c = ColorYellow
	case "blue":
		*c = ColorBlue
	case "fuchsia":
		*c = ColorFuchsia
	case "aqua":
		*c = ColorAqua
	case "white":
		*c = ColorWhite
	default:
		return fmt.Errorf("unsupported color name %s", string(in))
	}
	return nil
}
