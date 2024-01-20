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
	// PeriodicUpdate is responsible for handling periodic updates. The passed
	// duration may be very long in the case of reloading a chunk from disk.
	// Periodic update functions are expected to execute in linear time no
	// matter how long the duration.
	PeriodicUpdate(*Actor, *CityMap, time.Duration)
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
	TemplateID string                  // Template ID
	Position   util.Point              // Current position on the map
	AIModel    AIModel                 // AIModel for the actor
	NextThink  time.Time               // Time of the next think
	BodyParts  [BodyPartCount]BodyPart // Status of all body parts
	Equipment  [BodyPartCount]*Item    // All items equipped to the body, if any
	Inventory  []*Item                 // All items held in inventory, if any
	Weapon     *Item                   // The item wielded as a weapon, if any
	// Reconstructed values
	AITemplate string       // AI template name
	Name       string       // Descriptive name
	Rune       string       // Display rune
	Fg         termui.Color // Display foreground color
	Bg         termui.Color // Display background color
	Speed      float64      // Number of seconds between steps at walking pace
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
	for i := range a.Equipment { // Equipped items
		if util.GetBool(r) {
			a.Equipment[i] = NewItemFromReader(r)
		}
	}
	a.Inventory = make([]*Item, util.GetUint16(r)) // Inventory
	for i := range a.Inventory {
		a.Inventory[i] = NewItemFromReader(r)
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
	for _, i := range a.Equipment { // Equipped items
		if i == nil {
			util.PutBool(w, false)
		} else {
			util.PutBool(w, true)
			i.Write(w)
		}
	}
	util.PutUint16(w, uint16(len(a.Inventory))) // Inventory item count
	for _, i := range a.Inventory {             // Inventory items
		i.Write(w)
	}
}

// TargetedDamage applies a random amount of damage in the range [min-max) to
// the indicated body part scaled as needed and makes updates as necessary.
// Returns the amount of damage done.
func (a *Actor) TargetedDamage(which BodyPartCode, min, max float64, t time.Time, from *Actor) float64 {
	p := a.BodyParts[which]
	d := util.RandomF(min, max) * BodyPartInfo[which].DamageMod
	bs := ""
	os := "the"
	p.Health -= d
	if p.Health < 0 {
		p.Health = 0
		if !p.Broken {
			p.Broken = true
			bs = " BREAKING IT"
			if which != BodyPartHead && which != BodyPartBody && which != BodyPartHand {
				bs = " BREAKING THEM"
			}
			if which == BodyPartHead || which == BodyPartBody {
				a.Dead = true
				bs += ". KILLING BLOW!"
			}
		} else {
			os = "their broken"
			if a.IsPlayer {
				os = "your broken"
			}
		}
		p.BrokenUntil = t.Add(time.Hour * 24 * 14) // Takes two weeks for broken limbs to mend or zombies to get up
	}
	a.BodyParts[which] = p
	if a.IsPlayer {
		Log.Log(
			termui.ColorRed,
			"%s hit YOU in %s %s%s %d%%",
			from.Name,
			os,
			BodyPartInfo[which].Name,
			bs,
			int(d*100),
		)
	} else if from.IsPlayer {
		Log.Log(
			termui.ColorLime,
			"YOU hit %s in %s %s%s %d%%",
			a.Name,
			os,
			BodyPartInfo[which].Name,
			bs,
			int(d*100),
		)
	} else {
		Log.Log(
			termui.ColorYellow,
			"%s hit %s in %s %s%s %d%%",
			from.Name,
			a.Name,
			os,
			BodyPartInfo[which].Name,
			bs,
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

// WalkSpeed returns the current walking speed of this mobile in seconds.
func (a *Actor) WalkSpeed() float64 {
	// Broken legs or feet mean we crawl
	if a.BodyParts[BodyPartLegs].Broken || a.BodyParts[BodyPartFeet].Broken {
		return a.Speed * 4
	}
	// Otherwise we walk
	return a.Speed
}

// ActSpeed returns the current action speed of this mobile in seconds.
func (a *Actor) ActSpeed() float64 {
	// Broken arms or hands mean it's very difficult to take actions
	if a.BodyParts[BodyPartArms].Broken || a.BodyParts[BodyPartHand].Broken {
		return 4
	}
	return 1
}

// DropCorpse drops a corpse item for this actor.
func (a *Actor) DropCorpse(m *CityMap) {
	i := NewItem("Corpse", m.Now)
	i.SArg = a.TemplateID
	i.TArg = m.Now.Add(time.Hour * 24 * 14) // Takes two weeks for a corpse to resurrect
	i.Position = a.Position
	m.PlaceItem(i)
}

// WearItem attempts to wear the item as clothing. On failure a string is
// returned describing why the action failed as a complete, punctuated sentence.
// On success an empty string is returned.
func (a *Actor) WearItem(i *Item) string {
	if !i.Wearable {
		return "That item is not wearable."
	}
	if a.Equipment[i.WornBodyPart] != nil {
		return "An item is already being worn there."
	}
	a.Equipment[i.WornBodyPart] = i
	return ""
}

// WieldItem attempts to wield the item as a weapon. On failure a string is
// returned describing why the action failed as a complete, punctuated sentence.
// On success an empty string is returned.
func (a *Actor) WieldItem(i *Item) string {
	if !i.Weapon {
		return "That item is not a weapon."
	}
	if a.Weapon != nil {
		return "An item is already being wielded as a weapon."
	}
	a.Weapon = i
	return ""
}

// AddItemToInventory adds the item to the actor's inventory.
func (a *Actor) AddItemToInventory(i *Item) {
	a.Inventory = append(a.Inventory, i)
}
