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
	Act(*Actor, time.Time, *CityMap) time.Duration
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

// Actor represents a moving, thinking actor on the map.
type Actor struct {
	// Persistent values
	TemplateID string     // Template ID
	Position   util.Point // Current position on the map
	AIModel    AIModel    // AIModel for the actor
	NextThink  time.Time  // Time of the next think
	// Reconstructed values
	AITemplate string       // AI template name
	Name       string       // Descriptive name
	Rune       string       // Display rune
	Fg         termui.Color // Display foreground color
	Bg         termui.Color // Display background color
	WalkSpeed  float64      // Number of seconds between steps at walking pace
	// Transient values
	pqIdx int // Priority queue index
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
	a.NextThink = util.GetTime(r)
	return a
}

// Write writes the actor to the writer.
func (a *Actor) Write(w io.Writer) {
	util.PutUint32(w, 0)            // Version
	util.PutString(w, a.TemplateID) // Template ID
	util.PutPoint(w, a.Position)    // Map position
	a.AIModel.Write(w)              // AI model
	util.PutTime(w, a.NextThink)
}
