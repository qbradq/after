// Package events implements late-bound function execution for items
// and actors.
package events

import (
	"fmt"
	"time"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/util"
)

func init() {
	game.ExecuteItemUpdateEvent = ExecuteItemUpdateEvent
}

// useEvent is the signature of item event handlers. These handlers are used for
// the player interacting with an item in the game.
type useEvent func(*game.Item, *game.Actor, *game.CityMap) error

// Global registry of item use events.
var useEvents = map[string]useEvent{}

// rue registers an item use event by name.
func rue(name string, fn useEvent) {
	if _, found := useEvents[name]; found {
		panic(fmt.Errorf("duplicate item use event name %s", name))
	}
	useEvents[name] = fn
}

// ExecuteItemUseEvent executes the named item use event for the item if any.
// The second return parameter is true if the event handler was called.
func ExecuteItemUseEvent(name string, i *game.Item, src *game.Actor, m *game.CityMap) (error, bool) {
	// Sanity checks
	if i == nil {
		return nil, false
	}
	hn := i.Events[name]
	// No handler named for this event
	if len(hn) == 0 {
		return nil, false
	}
	h := useEvents[hn]
	if h == nil {
		return fmt.Errorf("item template %s, event %s references non-existent item use function %s",
			i.TemplateID, name, hn), false
	}
	return h(i, src, m), true
}

// updateEvent is the signature of item update event handlers. These handlers
// are used for the periodic update of all items. The handler may be called with
// very high values for d in the event of a chunk reload. The handler is
// expected to execute in linear time no mater the value of d.
type updateEvent func(*game.Item, *game.CityMap, time.Duration) error

// Global registry of item update events.
var updateEvents = map[string]updateEvent{}

// rpue registers an item periodic update event by name.
func rpue(name string, fn updateEvent) {
	if _, found := updateEvents[name]; found {
		panic(fmt.Errorf("duplicate item update event name %s", name))
	}
	updateEvents[name] = fn
}

// ExecuteItemUpdateEvent executes the named item update event for the item if
// any.
func ExecuteItemUpdateEvent(name string, i *game.Item, m *game.CityMap, d time.Duration) error {
	// Sanity checks
	if i == nil {
		return nil
	}
	hn := i.Events[name]
	// No handler named for this event
	if len(hn) == 0 {
		return nil
	}
	h := updateEvents[hn]
	if h == nil {
		return fmt.Errorf("item template %s, event %s references non-existent item update function %s",
			i.TemplateID, name, hn)
	}
	return h(i, m, d)
}

// vehicleEvent is the signature of vehicle event handlers. These handlers are
// used for all vehicle interactions.
type vehicleEvent func(*game.Vehicle, *game.VehicleLocation, *game.Item, util.Point, *game.Actor, *game.CityMap) error

// Global registry of item vehicle events.
var vehicleEvents = map[string]vehicleEvent{}

// rve registers a vehicle event by name.
func rve(name string, fn vehicleEvent) {
	if _, found := vehicleEvents[name]; found {
		panic(fmt.Errorf("duplicate vehicle event name %s", name))
	}
	vehicleEvents[name] = fn
}

// ExecuteVehicleEvent executes the named vehicle event for the given item, if
// any. The second return parameter is true if the event handler was called.
func ExecuteVehicleEvent(name string, v *game.Vehicle, l *game.VehicleLocation, i *game.Item, p util.Point, src *game.Actor, m *game.CityMap) (error, bool) {
	// Sanity checks
	if i == nil {
		return nil, false
	}
	hn := i.Events[name]
	// No handler named for this event
	if len(hn) == 0 {
		return nil, false
	}
	h := vehicleEvents[hn]
	if h == nil {
		return fmt.Errorf("item template %s, event %s references non-existent vehicle function %s",
			i.TemplateID, name, hn), false
	}
	return h(v, l, i, p, src, m), true
}
