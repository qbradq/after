package game

import (
	"fmt"
	"io"
	"strconv"
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
	Amount     int        // Stack amount
	FArg       float64    // Generic float argument
	SArg       string     // Generic string argument
	TArg       time.Time  // Generic time argument
	Inventory  []*Item    // Container contents if any
	// Reconstructed values
	Events          map[string]string // Map of event names to event handler names
	Name            string            // Descriptive name
	Rune            string            // Display rune
	Fg              termui.Color      // Display foreground color
	Bg              termui.Color      // Display background color
	BlocksVis       bool              // If true this item blocks visibility
	BlocksWalk      bool              // If true this item blocks walking
	Climbable       bool              // If true this item may be climbed over
	Destroyed       bool              // If true something has happened to this item to cause it to be destroyed, it will be removed from the world at the end of the next update cycle
	Stackable       bool              // If true this item can stack with others of the exact same template name
	Fixed           bool              // If true the item cannot be moved at all
	Wearable        bool              // If true this item can be worn as a piece of clothing
	WornBodyPart    BodyPartCode      // Code of the body part this item is worn on
	Weapon          bool              // If true this item can be wielded as a weapon
	WeaponMinDamage float64           // Minimum damage bonus when using this item as a weapon
	WeaponMaxDamage float64           // Maximum damage bonus when using this item as a weapon
	WeaponSwingStam float64           // Amount of stamina required to swing this weapon
	Container       bool              // If true this item contains other items
	Contents        []string          // Container content item statements if any
	// Cache values
	csCache []ItemStatement // Content statements cache
}

// NewItem creates a new item from the named template.
func NewItem(template string, now time.Time, genContents bool) *Item {
	i, found := ItemDefs[template]
	if !found {
		panic(fmt.Errorf("reference to non-existent item template %s", template))
	}
	ret := *i
	ret.LastUpdate = now
	if genContents {
		for _, s := range ret.csCache {
			for _, item := range s.Evaluate(now) {
				ret.AddItem(item)
			}
		}
	}
	return &ret
}

// NewItemFromReader reads the item information from r and returns a new Item
// with this information.
func NewItemFromReader(r io.Reader) *Item {
	_ = util.GetUint32(r)                 // Version
	tid := util.GetString(r)              // Template ID
	i := NewItem(tid, time.Time{}, false) // Create new object
	i.Position = util.GetPoint(r)         // Map position
	i.LastUpdate = util.GetTime(r)        // Time of last update
	i.Amount = int(util.GetUint32(r))     // Stack amount
	i.SArg = util.GetString(r)            // Generic string argument
	i.TArg = util.GetTime(r)              // Generic time argument
	n := int(util.GetUint16(r))           // Contents
	i.Inventory = make([]*Item, n)
	for idx := 0; idx < n; idx++ {
		i.Inventory[idx] = NewItemFromReader(r)
	}
	return i
}

// Write writes the actor to the writer.
func (i *Item) Write(w io.Writer) {
	util.PutUint32(w, 0)                        // Version
	util.PutString(w, i.TemplateID)             // Template ID
	util.PutPoint(w, i.Position)                // Map position
	util.PutTime(w, i.LastUpdate)               // Time of last update
	util.PutUint32(w, uint32(i.Amount))         // Stack amount
	util.PutString(w, i.SArg)                   // Generic string argument
	util.PutTime(w, i.TArg)                     // Generic time argument
	util.PutUint16(w, uint16(len(i.Inventory))) // Contents
	for _, i := range i.Inventory {
		i.Write(w)
	}
}

// DisplayName returns the string to display for this item in user-facing
// displays.
func (i *Item) DisplayName() string {
	ret := i.Name
	if i.Amount > 1 {
		ret += " x" + strconv.FormatInt(int64(i.Amount), 10)
	}
	return ret
}

// UIDisplayName returns the string to display in UIs like the inventory.
func (i *Item) UIDisplayName() string {
	ret := i.DisplayName()
	if i.Container {
		if len(i.Inventory) > 0 {
			return "+" + ret
		} else {
			return "-" + ret
		}
	}
	return " " + ret
}

// AddItem adds the item to this container's content if it is a container,
// returning true on success.
func (i *Item) AddItem(item *Item) bool {
	if !i.Container {
		return false
	}
	for _, o := range i.Inventory {
		if item == o {
			return false
		}
		// Try to stack
		if item.Stackable && item.TemplateID == o.TemplateID {
			o.Amount += item.Amount
			item.Destroyed = true
			return true
		}
	}
	i.Inventory = append(i.Inventory, item)
	return true
}

// RemoveItem removes the item from the container's content, returning true on
// success.
func (i *Item) RemoveItem(item *Item) bool {
	if !i.Container {
		return false
	}
	idx := -1
	for i, c := range i.Inventory {
		if c == item {
			idx = i
			break
		}
	}
	if idx < 0 {
		return false
	}
	i.Inventory = append(i.Inventory[:idx], i.Inventory[idx+1:]...)
	return true
}

// CacheContentStatements generates the cache of content statements. This must
// be called on all item prototypes after item and item gen loading is complete.
func (i *Item) CacheContentStatements() error {
	i.csCache = make([]ItemStatement, len(i.Contents))
	for idx, s := range i.Contents {
		is := ItemStatement{}
		if err := is.UnmarshalJSON([]byte("\"" + s + "\"")); err != nil {
			return err
		}
		i.csCache[idx] = is
	}
	return nil
}
