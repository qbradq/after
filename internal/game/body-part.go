package game

import (
	"fmt"
	"strings"
	"time"
)

// BodyPartCode is a code that indicates a player's body part.
type BodyPartCode uint8

const (
	BodyPartHead BodyPartCode = iota
	BodyPartBody
	BodyPartArms
	BodyPartLegs
	BodyPartHand
	BodyPartFeet
	BodyPartCount
)

func (c *BodyPartCode) UnmarshalJSON(in []byte) error {
	switch strings.ToLower(string(in[1 : len(in)-1])) {
	case "head":
		*c = BodyPartHead
	case "body":
		*c = BodyPartBody
	case "arms":
		*c = BodyPartBody
	case "hand":
		fallthrough
	case "hands":
		*c = BodyPartHand
	case "legs":
		*c = BodyPartLegs
	case "feet":
		*c = BodyPartFeet
	default:
		return fmt.Errorf("unsupported body part name %s", string(in))
	}
	return nil
}

// BodyPartInfo is a mapping of BodyPartCode to static information about a
// body part.
var BodyPartInfo = []struct {
	Name      string
	DamageMod float64
}{
	{"Head", 2.5},
	{"Body", 0.5},
	{"Arms", 1.0},
	{"Legs", 1.0},
	{"Hand", 1.5},
	{"Feet", 1.5},
}

// BodyPart encapsulates information about an actor's body part.
type BodyPart struct {
	// Persistent
	Health      float64   // Health between [0.0-1.0]
	BrokenUntil time.Time // When this body part will heal
	// Reconstituted values
	Which  BodyPartCode // Indicates which body part we describe
	Broken bool         // If true the body part is currently broken
}
