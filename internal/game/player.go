package game

import (
	"io"
	"time"

	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// Player implements the player's special actor.
type Player struct {
	Actor
	Stamina float64 // Stamina value from zero (exhausted) to one (well rested)
	Hunger  float64 // Hunger value from zero (starving) to one (stuffed)
	Thirst  float64 // Thirst value from zero (dehydrated to death) to one (slaked)
	Running bool    // If true the player is running and consuming stamina
}

// NewPlayer creates and returns a new Player struct.
func NewPlayer(now time.Time) *Player {
	a := NewActor("Player", now)
	a.IsPlayer = true
	p := &Player{
		Actor:   *a,
		Stamina: 0.3,
		Hunger:  0.4,
		Thirst:  0.25,
	}
	return p
}

// NewPlayerFromReader reads the player information from r and returns a new
// player with this information.
func NewPlayerFromReader(r io.Reader) *Player {
	a := NewActorFromReader(r)
	a.IsPlayer = true
	a.Name = util.GetString(r)
	p := &Player{
		Actor:   *a,
		Stamina: util.GetFloat(r),
		Hunger:  util.GetFloat(r),
		Thirst:  util.GetFloat(r),
		Running: util.GetBool(r),
	}
	return p
}

// Write writes the player to the writer.
func (a *Player) Write(w io.Writer) {
	a.Actor.Write(w)
	util.PutString(w, a.Name)   // Persist the player's name
	util.PutFloat(w, a.Stamina) // Stamina
	util.PutFloat(w, a.Hunger)  // Hunger
	util.PutFloat(w, a.Thirst)  // Thirst
	util.PutBool(w, a.Running)  // Running
}

// TookTurn is responsible for per-turn updates for the player.
func (a *Player) TookTurn(now time.Time, d time.Duration) {
	// Stamina regeneration
	a.Stamina += float64(d) / float64(time.Minute*5) // Takes 5 minutes to fully rest
	if a.Stamina > 1.0 {
		a.Stamina = 1.0
	}
	if a.Stamina < 0 {
		a.Stamina = 0
	}
	// Hunger decay
	a.Hunger -= float64(d) / float64(time.Hour*24*5) // Takes 5 days to start starving
	if a.Hunger < 0 {
		a.Hunger = 0
	}
	// Thirst decay
	a.Thirst -= float64(d) / float64(time.Hour*24*3) // Takes three days to start dying of dehydration
	if a.Thirst < 0 {
		a.Thirst = 0
	}
	// Process broken part timers
	days := float64(d) / float64(time.Hour*24)
	for i, p := range a.BodyParts {
		if !p.BrokenUntil.IsZero() && !now.Before(p.BrokenUntil) {
			p.Broken = false
			p.BrokenUntil = time.Time{}
		}
		a.BodyParts[i] = p
	}
	if a.Hunger > 0 && a.Thirst > 0 {
		// Not starving or dehydrated, heal body parts as normal
		for i, p := range a.BodyParts {
			p.Health += days / 2 // Body parts heal in two days
			if p.Health > 1 {
				p.Health = 1
			}
			a.BodyParts[i] = p
		}
	} else {
		// We are either starving or dehydrated or both so we wither
		for i, p := range a.BodyParts {
			p.Health -= days / 5 // Can last five days without food and water
			if p.Health < 0 {
				p.Health = 0 // Withering does not break body parts
			}
			a.BodyParts[i] = p
		}
		// Check if we're dead from withering
		if a.BodyParts[BodyPartHead].Health <= 0 ||
			a.BodyParts[BodyPartBody].Health <= 0 {
			a.Dead = true
			Log.Log(termui.ColorRed, "You have withered to death.")
		}
	}
}
