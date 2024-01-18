package game

import (
	"io"
	"time"

	"github.com/qbradq/after/lib/util"
)

// Player implements the player's special actor.
type Player struct {
	Actor
}

// NewPlayer creates and returns a new Player struct.
func NewPlayer(now time.Time) *Player {
	a := NewActor("Player", now)
	a.IsPlayer = true
	p := &Player{
		Actor: *a,
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
		Actor: *a,
	}
	return p
}

// Write writes the player to the writer.
func (a *Player) Write(w io.Writer) {
	a.Actor.Write(w)
	util.PutString(w, a.Name) // Persist the player's name
}
