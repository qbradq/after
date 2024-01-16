// Package game - in concert with packages events and ai - implements the
// game simulation.
package game

import (
	"os"

	"github.com/qbradq/after/lib/termui"
)

func init() {
	os.MkdirAll("saves", 0664)
}

// ChunkGen getter
var GetChunkGen func(string) ChunkGen

// Logger implementers can consume colored log messages.
type Logger interface {
	// Log adds a line to the log.
	Log(termui.Color, string, ...any)
}

// Global log consumer
var Log Logger
