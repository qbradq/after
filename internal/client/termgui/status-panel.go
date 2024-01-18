package termgui

import (
	"time"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// StatusPanel implements a 16x21 window that displays information about the
// player and current state of the city.
type StatusPanel struct {
	Position util.Point    // Position of the top-left corner of the 16x21 panel
	CityMap  *game.CityMap // City map we are displaying information about
}

// HandleEvent implements the termui.Mode interface.
func (m *StatusPanel) HandleEvent(s termui.TerminalDriver, e any) error {
	switch e.(type) {
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return nil
}

// Draw implements the termui.Mode interface.
func (m *StatusPanel) Draw(s termui.TerminalDriver) {
	b := util.Rect{
		TL: m.Position,
		BR: m.Position.Add(util.Point{X: 15, Y: 20}),
	}
	termui.DrawStringCenter(s, b, m.CityMap.Player.Name, termui.CurrentTheme.Normal.Foreground(termui.ColorAqua))
	b.TL.Y++
	// Overall status display
	if m.CityMap.Player.Dead {
		termui.DrawStringCenter(s, b, "Dead", termui.CurrentTheme.Normal.Foreground(termui.ColorRed))
	} else {
		termui.DrawStringCenter(s, b, "Normal", termui.CurrentTheme.Normal.Foreground(termui.ColorSilver))
	}
	b.TL.Y++
	for _, part := range m.CityMap.Player.BodyParts {
		termui.DrawStringLeft(s, b, game.BodyPartInfo[part.Which].Name, termui.CurrentTheme.Normal)
		nb := b
		nb.TL.X += 5
		// Body part is broken, don't display internal HP until fully healed
		if part.Broken {
			termui.DrawStringCenter(s, nb, "--Broken--", termui.CurrentTheme.Normal.Foreground(termui.ColorRed))
			b.TL.Y++
			continue
		}
		// Else continue with health display
		nb.BR.Y = nb.TL.Y
		nb.BR.X--
		termui.DrawFill(s, nb, termui.Glyph{
			Rune:  '-',
			Style: termui.StyleDefault.Foreground(termui.ColorRed),
		})
		hp := int(part.Health * 10)
		if hp >= 10 {
			hp--
		}
		nb.BR.X = nb.TL.X + hp
		termui.DrawFill(s, nb, termui.Glyph{
			Rune:  '=',
			Style: termui.StyleDefault.Foreground(termui.ColorLime),
		})
		b.TL.Y++
	}
	b.TL.Y = m.Position.Y + 19
	termui.DrawStringLeft(s, b, m.CityMap.Now.Format("Mon Jan 02 2006"), termui.CurrentTheme.Normal)
	b.TL.Y++
	termui.DrawStringRight(s, b, m.CityMap.Now.Format(time.TimeOnly), termui.CurrentTheme.Normal)
}
