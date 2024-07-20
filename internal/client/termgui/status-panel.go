package termgui

import (
	"time"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// statusPanel implements a 16x21 window that displays information about the
// player and current state of the city.
type statusPanel struct {
	Position util.Point    // Position of the top-left corner of the 16x21 panel
	CityMap  *game.CityMap // City map we are displaying information about
}

// HandleEvent implements the termui.Mode interface.
func (m *statusPanel) HandleEvent(s termui.TerminalDriver, e any) error {
	switch e.(type) {
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return nil
}

// Draw implements the termui.Mode interface.
func (m *statusPanel) Draw(s termui.TerminalDriver) {
	b := util.Rect{
		TL: m.Position,
		BR: m.Position.Add(util.Point{X: 15, Y: 20}),
	}
	termui.DrawStringCenter(s, b, m.CityMap.Player.Name, termui.CurrentTheme.Normal.Foreground(termui.ColorAqua))
	b.TL.Y++
	// Overall status display
	ss := "Normal"
	sss := termui.CurrentTheme.Normal.Foreground(termui.ColorSilver)
	if m.CityMap.Player.BodyParts[game.BodyPartArms].Broken ||
		m.CityMap.Player.BodyParts[game.BodyPartHand].Broken {
		ss = "Mangled"
		sss = sss.Foreground(termui.ColorYellow)
	}
	if m.CityMap.Player.BodyParts[game.BodyPartLegs].Broken ||
		m.CityMap.Player.BodyParts[game.BodyPartFeet].Broken {
		ss = "Crippled"
		sss = sss.Foreground(termui.ColorYellow)
	}
	if m.CityMap.Player.Dead {
		ss = "Dead"
		sss = sss.Foreground(termui.ColorRed)
	}
	termui.DrawStringCenter(s, b, ss, sss)
	b.TL.Y++
	// Hunger display
	termui.DrawStringLeft(s, b, "Food", termui.CurrentTheme.Normal)
	nb := b
	nb.TL.X += 5
	if m.CityMap.Player.Hunger == 0 {
		termui.DrawStringCenter(s, nb, "-Starving-", termui.CurrentTheme.Normal.Foreground(termui.ColorRed))
	} else {
		nb.BR.Y = nb.TL.Y
		nb.BR.X--
		termui.DrawFill(s, nb, termui.Glyph{
			Rune:  '-',
			Style: termui.StyleDefault.Foreground(termui.ColorOlive),
		})
		hp := int(m.CityMap.Player.Hunger * 10)
		if hp >= 10 {
			hp--
		}
		nb.BR.X = nb.TL.X + hp
		termui.DrawFill(s, nb, termui.Glyph{
			Rune:  '=',
			Style: termui.StyleDefault.Foreground(termui.ColorYellow),
		})
	}
	b.TL.Y++
	// Thirst display
	termui.DrawStringLeft(s, b, "Water", termui.CurrentTheme.Normal)
	nb = b
	nb.TL.X += 5
	if m.CityMap.Player.Thirst == 0 {
		termui.DrawStringCenter(s, nb, "Dehydrated", termui.CurrentTheme.Normal.Foreground(termui.ColorRed))
	} else {
		nb.BR.Y = nb.TL.Y
		nb.BR.X--
		termui.DrawFill(s, nb, termui.Glyph{
			Rune:  '-',
			Style: termui.StyleDefault.Foreground(termui.ColorNavy),
		})
		hp := int(m.CityMap.Player.Thirst * 10)
		if hp >= 10 {
			hp--
		}
		nb.BR.X = nb.TL.X + hp
		termui.DrawFill(s, nb, termui.Glyph{
			Rune:  '=',
			Style: termui.StyleDefault.Foreground(termui.ColorAqua),
		})
	}
	b.TL.Y++
	// Body part display
	for _, part := range m.CityMap.Player.BodyParts {
		termui.DrawStringLeft(s, b, game.BodyPartInfo[part.Which].Name, termui.CurrentTheme.Normal)
		nb = b
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
