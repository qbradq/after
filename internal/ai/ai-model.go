package ai

import (
	"fmt"
	"io"
	"time"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/util"
)

func init() {
	// Inject ourselves into the game module
	game.NewAIModel = NewAIModel
	game.NewAIModelFromReader = NewAIModelFromReader
}

// aiFn is the function signature all AI functions take.
type aiFn func(*AIModel, *game.Actor, *game.CityMap) time.Duration

// Global registry of AI functions.
var aiFns = map[string]aiFn{}

// regFn registers an AI function by name.
func regFn(name string, fn aiFn) {
	if _, found := aiFns[name]; found {
		panic(fmt.Errorf("duplicate AI function %s", name))
	}
	aiFns[name] = fn
}

// AIModel implements the thinking AI of CPU-controlled actors.
type AIModel struct {
	POI  util.Point    // Point of interest
	Path game.Path     // Path from current position to poi
	tid  string        // Template ID
	act  string        // Act makes the actor take its next action and returns the delay until that actor's next Act() call.
	cd   time.Duration // General-purpose cool-down counter
}

// aiModelConstructor functions construct AIModel objects pre-configured for a
// defined set of behaviors.
type aiModelConstructor func() *AIModel

// Global registry of AI model constructors
var ctors = map[string]aiModelConstructor{}

// reg registers the given ctor by name.
func reg(name string, ctor aiModelConstructor) {
	if _, found := ctors[name]; found {
		panic(fmt.Errorf("duplicate AI model %s", name))
	}
	ctors[name] = ctor
}

// NewAIModel returns a newly constructed AIModel object with the named
// configuration.
func NewAIModel(name string) game.AIModel {
	ctor := ctors[name]
	if ctor == nil {
		panic(fmt.Errorf("undefined AIModel constructor %s", name))
	}
	ai := ctor()
	ai.tid = name
	return ai
}

// NewAIModelFromReader constructs a new AIModel object from the information in
// the reader.
func NewAIModelFromReader(r io.Reader) game.AIModel {
	_ = util.GetUint32(r)                          // version
	ai := NewAIModel(util.GetString(r)).(*AIModel) // Template ID
	ai.POI = util.GetPoint(r)                      // Point of interest
	ai.act = util.GetString(r)                     // Act handler
	var b = []byte{0}                              // Path to PoI
	r.Read(b)
	ai.Path = make(game.Path, b[0])
	b = make([]byte, b[0])
	r.Read(b)
	for i, d := range b {
		ai.Path[i] = util.Direction(d)
	}
	return ai
}

// Write writes out state information. See NewAIModelFromReader().
func (ai *AIModel) Write(w io.Writer) {
	util.PutUint32(w, 0)              // Version
	util.PutString(w, ai.tid)         // Template ID
	util.PutPoint(w, ai.POI)          // Point of interest
	util.PutString(w, ai.act)         // Current act handler
	b := make([]byte, len(ai.Path)+1) // Path to PoI
	b[0] = byte(len(ai.Path))
	for i, d := range ai.Path {
		b[i+1] = byte(d)
	}
	w.Write(b)
}

// Act is responsible for calling act().
func (ai *AIModel) Act(a *game.Actor, m *game.CityMap) time.Duration {
	return aiFns[ai.act](ai, a, m)
}

func (ai *AIModel) setPOI(p util.Point, a *game.Actor, m *game.CityMap) {
	ai.POI = p
	ai.Path = ai.Path[:0]
	game.NewPath(a.Position, p, m, &ai.Path)
}

func (ai *AIModel) targetPlayer(a *game.Actor, m *game.CityMap) bool {
	// If we are too far away from the player to see them we bail
	if a.Position.Distance(m.Player.Position) > a.SightRange {
		return false
	}
	// If we can't see the player we can't target them
	if !m.CanSeePlayerFrom(a.Position) {
		return false
	}
	// If the player is already standing at our POI we don't need to re-path
	if m.Player.Position == ai.POI {
		return true
	}
	// Re-path
	ai.setPOI(m.Player.Position, a, m)
	return true
}
