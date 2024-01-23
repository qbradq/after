package game

import (
	"fmt"
	"io"
	"time"

	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// Item definitions
var ItemDefs = map[string]*Item{}

// Item is any dynamic item within the world, anything that can be used, taken,
// destroyed or built.
type Item struct {
	// Persistent values
	TemplateID string     // Template ID
	Position   util.Point // Current position on the map
	LastUpdate time.Time  // Time of the last call to event update
	SArg       string     // Generic string argument
	TArg       time.Time  // Generic time argument
	// Reconstructed values
	Events       map[string]string // Map of event names to event handler names
	Name         string            // Descriptive name
	Rune         string            // Display rune
	Fg           termui.Color      // Display foreground color
	Bg           termui.Color      // Display background color
	BlocksVis    bool              // If true this item blocks visibility
	BlocksWalk   bool              // If true this item blocks walking
	Climbable    bool              // If true this item may be climbed over
	Destroyed    bool              // If true something has happened to this item to cause it to be destroyed, it will be removed from the world at the end of the next update cycle
	Fixed        bool              // If true the item cannot be moved at all
	Wearable     bool              // If true this item can be worn as a piece of clothing
	WornBodyPart BodyPartCode      // Code of the body part this item is worn on
	Weapon       bool              // If true this item can be wielded as a weapon
}

// NewItem creates a new item from the named template.
func NewItem(template string, now time.Time) *Item {
	i, found := ItemDefs[template]
	if !found {
		panic(fmt.Errorf("reference to non-existent item template %s", template))
	}
	ret := *i
	ret.LastUpdate = now
	return &ret
}

// NewItemFromReader reads the item information from r and returns a new Item
// with this information.
func NewItemFromReader(r io.Reader) *Item {
	_ = util.GetUint32(r)          // Version
	tid := util.GetString(r)       // Template ID
	i := NewItem(tid, time.Time{}) // Create new object
	i.Position = util.GetPoint(r)  // Map position
	i.LastUpdate = util.GetTime(r) // Time of last update
	i.SArg = util.GetString(r)     // Generic string argument
	i.TArg = util.GetTime(r)       // Generic time argument
	return i
}

// Write writes the actor to the writer.
func (i *Item) Write(w io.Writer) {
	util.PutUint32(w, 0)            // Version
	util.PutString(w, i.TemplateID) // Template ID
	util.PutPoint(w, i.Position)    // Map position
	util.PutTime(w, i.LastUpdate)   // Time of last update
	util.PutString(w, i.SArg)       // Generic string argument
	util.PutTime(w, i.TArg)         // Generic time argument
}
