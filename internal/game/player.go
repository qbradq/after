package game

import (
	"io"
	"time"
)

// Player implements the player's special actor.
type Player struct {
	Actor
}

// NewPlayer creates and returns a new Player struct.
func NewPlayer(now time.Time) *Player {
	a := NewActor("Player", now)
	return &Player{
		Actor: *a,
	}
}

// NewPlayerFromReader reads the player information from r and returns a new
// player with this information.
func NewPlayerFromReader(r io.Reader) *Player {
	a := NewActorFromReader(r)
	return &Player{
		Actor: *a,
	}
}

// Write writes the player to the writer.
func (a *Player) Write(w io.Writer) {
	a.Actor.Write(w)
}
