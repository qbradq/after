package game

import (
	"os"
)

func init() {
	os.MkdirAll("saves", 0664)
}

// ChunkGen getter
var GetChunkGen func(string) ChunkGen
