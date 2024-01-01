package game

import (
	"fmt"
	"io"

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
	// Reconstructed values
	Events     map[string]string // Map of event names to event handler names
	Name       string            // Descriptive name
	Rune       string            // Display rune
	Fg         termui.Color      // Display foreground color
	Bg         termui.Color      // Display background color
	BlocksVis  bool              // If true this item blocks visibility
	BlocksWalk bool              // If true this item blocks walking
}

// NewItem creates a new item from the named template.
func NewItem(template string) *Item {
	i, found := ItemDefs[template]
	if !found {
		panic(fmt.Errorf("reference to non-existent item template %s", template))
	}
	ret := *i
	return &ret
}

// NewItemFromReader reads the item information from r and returns a new Item
// with this information.
func NewItemFromReader(r io.Reader) *Item {
	_ = util.GetUint32(r)         // Version
	tid := util.GetString(r)      // Template ID
	i := NewItem(tid)             // Create new object
	i.Position = util.GetPoint(r) // Map position
	return i
}

// Write writes the actor to the writer.
func (i *Item) Write(w io.Writer) {
	util.PutUint32(w, 0)            // Version
	util.PutString(w, i.TemplateID) // Template ID
	util.PutPoint(w, i.Position)    // Map position
}
