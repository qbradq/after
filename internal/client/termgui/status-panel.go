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
	b := util.NewRectWH(17, 23).Add(m.Position)
	//
	// Player states
	//
	db := b
	termui.DrawBox(s, db, termui.CurrentTheme.Normal)
	db = db.Shrink(1)
	termui.DrawStringCenter(s, db, m.CityMap.Player.Name, termui.CurrentTheme.Normal.Foreground(termui.ColorAqua))
	db.TL.Y++
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
	termui.DrawStringCenter(s, db, ss, sss)
	db.TL.Y++
	// Running / Walking display
	ss = "Walking"
	sss = termui.CurrentTheme.Normal
	if m.CityMap.Player.Running {
		ss = "Running"
		sss = sss.Foreground(termui.ColorRed)
	}
	termui.DrawStringCenter(s, db, ss, sss)
	//
	// Body status display
	//
	db = b
	db.TL.Y += 4
	termui.DrawBox(s, db, termui.CurrentTheme.Normal)
	db = db.Shrink(1)
	// Stamina display
	termui.DrawStringLeft(s, db, "Stam", termui.CurrentTheme.Normal)
	nb := db
	nb.TL.X += 5
	if m.CityMap.Player.Stamina == 0 {
		termui.DrawStringCenter(s, nb, "Exhausted", termui.CurrentTheme.Normal.Foreground(termui.ColorRed))
	} else {
		nb.BR.Y = nb.TL.Y
		nb.BR.X--
		termui.DrawFill(s, nb, termui.Glyph{
			Rune:  '-',
			Style: termui.StyleDefault.Foreground(termui.ColorPurple),
		})
		hp := int(m.CityMap.Player.Stamina * 10)
		if hp >= 10 {
			hp--
		}
		nb.BR.X = nb.TL.X + hp
		termui.DrawFill(s, nb, termui.Glyph{
			Rune:  '=',
			Style: termui.StyleDefault.Foreground(termui.ColorFuchsia),
		})
	}
	db.TL.Y++
	// Hunger display
	termui.DrawStringLeft(s, db, "Food", termui.CurrentTheme.Normal)
	nb = db
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
	db.TL.Y++
	// Thirst display
	termui.DrawStringLeft(s, db, "Water", termui.CurrentTheme.Normal)
	nb = db
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
	//
	// Mind status display
	//
	db = b
	db.TL.Y += 8
	termui.DrawBox(s, db, termui.CurrentTheme.Normal)
	db = db.Shrink(1)
	// Happiness display
	termui.DrawStringLeft(s, db, "Joy", termui.CurrentTheme.Normal)
	nb = db
	nb.TL.X += 5
	if m.CityMap.Player.Joy == 0 {
		termui.DrawStringCenter(s, nb, "-Suicidal-", termui.CurrentTheme.Normal.Foreground(termui.ColorRed))
	} else {
		nb.BR.Y = nb.TL.Y
		nb.BR.X--
		termui.DrawFill(s, nb, termui.Glyph{
			Rune:  '-',
			Style: termui.StyleDefault.Foreground(termui.ColorOlive),
		})
		hp := int(m.CityMap.Player.Joy * 10)
		if hp >= 10 {
			hp--
		}
		nb.BR.X = nb.TL.X + hp
		termui.DrawFill(s, nb, termui.Glyph{
			Rune:  '=',
			Style: termui.StyleDefault.Foreground(termui.ColorYellow),
		})
	}
	db.TL.Y++
	// Sanity display
	termui.DrawStringLeft(s, db, "Sane", termui.CurrentTheme.Normal)
	nb = db
	nb.TL.X += 5
	if m.CityMap.Player.Mind == 0 {
		termui.DrawStringCenter(s, nb, "--Insane--", termui.CurrentTheme.Normal.Foreground(termui.ColorRed))
	} else {
		nb.BR.Y = nb.TL.Y
		nb.BR.X--
		termui.DrawFill(s, nb, termui.Glyph{
			Rune:  '-',
			Style: termui.StyleDefault.Foreground(termui.ColorGreen),
		})
		hp := int(m.CityMap.Player.Mind * 10)
		if hp >= 10 {
			hp--
		}
		nb.BR.X = nb.TL.X + hp
		termui.DrawFill(s, nb, termui.Glyph{
			Rune:  '=',
			Style: termui.StyleDefault.Foreground(termui.ColorLime),
		})
	}
	db.TL.Y++
	// Sleep display
	termui.DrawStringLeft(s, db, "Zzzz", termui.CurrentTheme.Normal)
	nb = db
	nb.TL.X += 5
	if m.CityMap.Player.Sleep == 0 {
		termui.DrawStringCenter(s, nb, "Drop  Dead", termui.CurrentTheme.Normal.Foreground(termui.ColorRed))
	} else {
		nb.BR.Y = nb.TL.Y
		nb.BR.X--
		termui.DrawFill(s, nb, termui.Glyph{
			Rune:  '-',
			Style: termui.StyleDefault.Foreground(termui.ColorNavy),
		})
		hp := int(m.CityMap.Player.Sleep * 10)
		if hp >= 10 {
			hp--
		}
		nb.BR.X = nb.TL.X + hp
		termui.DrawFill(s, nb, termui.Glyph{
			Rune:  '=',
			Style: termui.StyleDefault.Foreground(termui.ColorBlue),
		})
	}
	//
	// Body part display
	//
	db = b
	db.TL.Y += 12
	termui.DrawBox(s, db, termui.CurrentTheme.Normal)
	db = db.Shrink(1)
	for _, part := range m.CityMap.Player.BodyParts {
		termui.DrawStringLeft(s, db, game.BodyPartInfo[part.Which].Name, termui.CurrentTheme.Normal)
		nb = db
		nb.TL.X += 5
		// Body part is broken, don't display internal HP until fully healed
		if part.Broken {
			termui.DrawStringCenter(s, nb, "--Broken--", termui.CurrentTheme.Normal.Foreground(termui.ColorRed))
			db.TL.Y++
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
		db.TL.Y++
	}
	//
	// Time display
	//
	db = b
	db.TL.Y += 19
	termui.DrawBox(s, db, termui.CurrentTheme.Normal)
	db = db.Shrink(1)
	termui.DrawStringLeft(s, db, m.CityMap.Now.Format("Mon Jan 02 2006"), termui.CurrentTheme.Normal)
	db.TL.Y++
	termui.DrawStringRight(s, db, m.CityMap.Now.Format(time.TimeOnly), termui.CurrentTheme.Normal)
}
