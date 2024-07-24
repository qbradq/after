// Package game - in concert with packages events and ai - implements the
// game simulation.
package game

import (
	"os"
	"time"

	"github.com/qbradq/after/lib/termui"
)

func init() {
	os.MkdirAll("saves", 0664)
}

// GetChunkGen is the ChunkGen getter.
var GetChunkGen func(string, string) ChunkGen

// Logger implementers can consume colored log messages.
type Logger interface {
	// Log adds a line to the log.
	Log(termui.Color, string, ...any)
}

// Log is the global log consumer.
var Log Logger

// ExecuteItemUpdateEvent is the Item update event executer.
var ExecuteItemUpdateEvent func(string, *Item, *CityMap, time.Duration) error
