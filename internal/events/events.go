package events

import (
	"fmt"

	"github.com/qbradq/after/internal/game"
)

// itemEvent is the signature of item event handlers
type itemEvent func(*game.Item, *game.Actor, *game.CityMap) error

// Global registry of item events
var itemEvents = map[string]itemEvent{}

// rie registers an item event by name.
func rie(name string, fn itemEvent) {
	if _, found := itemEvents[name]; found {
		panic(fmt.Errorf("duplicate event name %s", name))
	}
	itemEvents[name] = fn
}

// ExecuteItemEvent executes the named item event for the item if any.
func ExecuteItemEvent(name string, i *game.Item, src *game.Actor, m *game.CityMap) error {
	// Sanity checks
	if i == nil {
		return nil
	}
	hn := i.Events[name]
	// No handler named for this event
	if len(hn) == 0 {
		return nil
	}
	h := itemEvents[hn]
	if h == nil {
		panic(fmt.Errorf("item template %s, event %s references non-existent item event %s",
			i.TemplateID, name, hn))
	}
	return h(i, src, m)
}
