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
type aiFn func(*AIModel, *game.Actor, time.Time, *game.CityMap) time.Duration

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
	tid string     // Template ID
	poi util.Point // Point of interest
	act string     // Act makes the actor take its next action and returns the delay until that actor's next Act() call.
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
	ai.poi = util.GetPoint(r)                      // Point of interest
	ai.act = util.GetString(r)                     // Act handler
	return ai
}

// Write writes out state information. See NewAIModelFromReader().
func (ai *AIModel) Write(w io.Writer) {
	util.PutUint32(w, 0)      // Version
	util.PutString(w, ai.tid) // Template ID
	util.PutPoint(w, ai.poi)  // Point of interest
	util.PutString(w, ai.act) // Current act handler
}

// Act is responsible for calling act().
func (ai *AIModel) Act(a *game.Actor, t time.Time, m *game.CityMap) time.Duration {
	return aiFns[ai.act](ai, a, t, m)
}
