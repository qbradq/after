package game

import (
	"io"
	"time"

	"github.com/qbradq/after/lib/util"
)

// BodyPartCode is a code that indicates a player's body part.
type BodyPartCode uint8

const (
	BodyPartHead BodyPartCode = iota
	BodyPartArms
	BodyPartBody
	BodyPartLegs
	BodyPartHand
	BodyPartFeet
	BodyPartCount
)

// BodyPartNames is a mapping of BodyPartCode to 4-character body part name.
var BodyPartNames = []string{
	"Head",
	"Arms",
	"Body",
	"Legs",
	"Hand",
	"Feet",
}

// BodyPart encapsulates information about a player's body part.
type BodyPart struct {
	Which  BodyPartCode // Indicates which body part we describe
	Health float64      // Health between [0.0-1.0]
}

// Player implements the player's special actor.
type Player struct {
	Actor
	BodyParts [BodyPartCount]BodyPart
}

// NewPlayer creates and returns a new Player struct.
func NewPlayer(now time.Time) *Player {
	a := NewActor("Player", now)
	p := &Player{
		Actor: *a,
	}
	for i := range p.BodyParts {
		p.BodyParts[i].Which = BodyPartCode(i)
		p.BodyParts[i].Health = 1
	}
	return p
}

// NewPlayerFromReader reads the player information from r and returns a new
// player with this information.
func NewPlayerFromReader(r io.Reader) *Player {
	a := NewActorFromReader(r)
	p := &Player{
		Actor: *a,
	}
	for i := range p.BodyParts {
		p.BodyParts[i].Health = util.GetFloat(r)
	}
	return p
}

// Write writes the player to the writer.
func (a *Player) Write(w io.Writer) {
	a.Actor.Write(w)
	for _, p := range a.BodyParts {
		util.PutFloat(w, p.Health)
	}
}
