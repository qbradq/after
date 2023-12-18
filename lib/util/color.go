package util

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Color is a wrapper value for tcell.Color.
type Color tcell.Color

const (
	ColorBlack   = Color(tcell.ColorBlack)
	ColorMaroon  = Color(tcell.ColorMaroon)
	ColorGreen   = Color(tcell.ColorGreen)
	ColorOlive   = Color(tcell.ColorOlive)
	ColorNavy    = Color(tcell.ColorNavy)
	ColorPurple  = Color(tcell.ColorPurple)
	ColorTeal    = Color(tcell.ColorTeal)
	ColorSilver  = Color(tcell.ColorSilver)
	ColorGray    = Color(tcell.ColorGray)
	ColorRed     = Color(tcell.ColorRed)
	ColorLime    = Color(tcell.ColorLime)
	ColorYellow  = Color(tcell.ColorYellow)
	ColorBlue    = Color(tcell.ColorBlue)
	ColorFuchsia = Color(tcell.ColorFuchsia)
	ColorAqua    = Color(tcell.ColorAqua)
	ColorWhite   = Color(tcell.ColorWhite)
)

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
