package game

import (
	"fmt"
	"io"

	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// Map of all actor definitions from all mods.
var ActorDefs = map[string]*Actor{}

// Actor represents a moving, thinking actor on the map.
type Actor struct {
	TemplateID string       // Template ID
	Position   util.Point   // Current position on the map
	Name       string       // Descriptive name
	Rune       string       // Display rune
	Fg         termui.Color // Display foreground color
	Bg         termui.Color // Display background color
}

// NewActor creates a new actor from the named template.
func NewActor(template string) *Actor {
	a, found := ActorDefs[template]
	if !found {
		panic(fmt.Errorf("reference to non-existent actor template %s", template))
	}
	ret := *a
	return &ret
}

// NewActorFromReader reads the actor information from r and returns a new Actor
// with this information.
func NewActorFromReader(r io.Reader) *Actor {
	_ = util.GetUint32(r)         // Version
	tid := util.GetString(r)      // Template ID
	a := NewActor(tid)            // Create new object
	a.Position = util.GetPoint(r) // Map position
	return a
}

// Write writes the actor to the writer.
func (a *Actor) Write(w io.Writer) {
	util.PutUint32(w, 0)            // Version
	util.PutString(w, a.TemplateID) // Template ID
	util.PutPoint(w, a.Position)    // Map position
}
