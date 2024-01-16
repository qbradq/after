package game

import (
	"fmt"
	"io"
	"time"

	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// AIModel is the interface the actor AI models must implement.
type AIModel interface {
	// Act is responsible for taking the next action for the actor and returning
	// the duration until the next call to Act().
	Act(*Actor, *CityMap) time.Duration
	// Write writes the internal state of the model to the writer.
	Write(io.Writer)
}

// NewAIModelFromReader reads AI model state information from r and constructs
// a new AIModel ready for use.
var NewAIModelFromReader func(io.Reader) AIModel

// NewAIModel should return a new AI model by template name.
var NewAIModel func(string) AIModel

// Map of all actor definitions from all mods.
var ActorDefs = map[string]*Actor{}

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

// BodyPartNames is a mapping of BodyPartCode to 4-character body part name.
var BodyPartNames = []string{
	"Head",
	"Body",
	"Arms",
	"Legs",
	"Hand",
	"Feet",
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

// Actor represents a moving, thinking actor on the map.
type Actor struct {
	// Persistent values
	TemplateID string                  // Template ID
	Position   util.Point              // Current position on the map
	AIModel    AIModel                 // AIModel for the actor
	NextThink  time.Time               // Time of the next think
	BodyParts  [BodyPartCount]BodyPart // Status of all body parts
	// Reconstructed values
	AITemplate string       // AI template name
	Name       string       // Descriptive name
	Rune       string       // Display rune
	Fg         termui.Color // Display foreground color
	Bg         termui.Color // Display background color
	WalkSpeed  float64      // Number of seconds between steps at walking pace
	SightRange int          // Distance this actor can see
	MinDamage  float64      // Minimum damage done by normal attacks
	MaxDamage  float64      // Maximum damage done by normal attacks
	IsPlayer   bool         // Only true for the player's actor
	// Transient values
	Dead  bool // If true something has happened to this actor to cause death
	pqIdx int  // Priority queue index
}

// NewActor creates a new actor from the named template.
func NewActor(template string, now time.Time) *Actor {
	a, found := ActorDefs[template]
	if !found {
		panic(fmt.Errorf("reference to non-existent actor template %s", template))
	}
	ret := *a
	ret.AIModel = NewAIModel(ret.AITemplate)
	ret.NextThink = now.Add(time.Second * time.Duration(util.RandomF(0, 1)))
	for i := range ret.BodyParts {
		ret.BodyParts[i].Which = BodyPartCode(i)
		ret.BodyParts[i].Health = 1
	}
	return &ret
}

// NewActorFromReader reads the actor information from r and returns a new Actor
// with this information.
func NewActorFromReader(r io.Reader) *Actor {
	_ = util.GetUint32(r)               // Version
	tid := util.GetString(r)            // Template ID
	a := NewActor(tid, time.Time{})     // Create new object
	a.Position = util.GetPoint(r)       // Map position
	a.AIModel = NewAIModelFromReader(r) // AI model
	a.NextThink = util.GetTime(r)       // Next think time
	for i := range a.BodyParts {        // Body part status
		p := BodyPart{
			Which:       BodyPartCode(i),
			Health:      util.GetFloat(r),
			BrokenUntil: util.GetTime(r),
		}
		if !p.BrokenUntil.IsZero() {
			p.Broken = true
		}
		a.BodyParts[i] = p
	}
	return a
}

// Write writes the actor to the writer.
func (a *Actor) Write(w io.Writer) {
	util.PutUint32(w, 0)            // Version
	util.PutString(w, a.TemplateID) // Template ID
	util.PutPoint(w, a.Position)    // Map position
	a.AIModel.Write(w)              // AI model
	util.PutTime(w, a.NextThink)    // Next think time
	for _, p := range a.BodyParts { // Body part status
		util.PutFloat(w, p.Health)
		util.PutTime(w, p.BrokenUntil)
	}
}

// TargetedDamage applies a random amount of damage in the range [min-max) to
// the indicated body part scaled as needed and makes updates as necessary.
// Returns the amount of damage done.
func (a *Actor) TargetedDamage(which BodyPartCode, min, max float64, t time.Time, from *Actor) float64 {
	p := a.BodyParts[which]
	d := util.RandomF(min, max)
	p.Health -= d
	if p.Health < 0 {
		p.Health = 0
		p.Broken = true
		p.BrokenUntil = t.Add(time.Hour * 24 * 14)
	}
	a.BodyParts[which] = p
	if a.IsPlayer {
		Log.Log(
			termui.ColorRed,
			"%s hit YOU for %d points of damage.",
			from.Name,
			int(d*100),
		)
	} else if from.IsPlayer {
		Log.Log(
			termui.ColorLime,
			"YOU hit %s for %d points of damage.",
			a.Name,
			int(d*100),
		)
	} else {
		Log.Log(
			termui.ColorYellow,
			"%s hit %s for %d points of damage.",
			from.Name,
			a.Name,
			int(d*100),
		)
	}
	return d
}

// Damage calls TargetedDamage with a random body part weighted to hit
// probabilities. Returns the amount of damage done.
func (a *Actor) Damage(min, max float64, t time.Time, from *Actor) float64 {
	var which BodyPartCode
	r := util.Random(0, 99)
	if r < 5 {
		which = BodyPartHead
	} else if r < 15 {
		which = BodyPartFeet
	} else if r < 25 {
		which = BodyPartHand
	} else if r < 45 {
		which = BodyPartLegs
	} else if r < 65 {
		which = BodyPartArms
	} else {
		which = BodyPartBody
	}
	return a.TargetedDamage(which, min, max, t, from)
}
