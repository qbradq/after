package game

import (
	"io"
	"time"

	"github.com/qbradq/after/lib/util"
)

// Player implements the player's special actor.
type Player struct {
	Actor
	Hunger float64 // Hunger value from zero (starving) to one (stuffed)
	Thirst float64 // Thirst value from zero (dehydrated to death) to one (slaked)
}

// NewPlayer creates and returns a new Player struct.
func NewPlayer(now time.Time) *Player {
	a := NewActor("Player", now)
	a.IsPlayer = true
	p := &Player{
		Actor:  *a,
		Hunger: 0.2,
		Thirst: 0.1,
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
		Actor:  *a,
		Hunger: util.GetFloat(r),
		Thirst: util.GetFloat(r),
	}
	return p
}

// Write writes the player to the writer.
func (a *Player) Write(w io.Writer) {
	a.Actor.Write(w)
	util.PutString(w, a.Name)  // Persist the player's name
	util.PutFloat(w, a.Hunger) // Hunger
	util.PutFloat(w, a.Thirst) // Thirst
}
