package util

import (
	"fmt"
	"strings"
)

// Facing represents one of the four cardinal directions.
type Facing uint8

const (
	FacingNorth Facing = iota
	FacingEast
	FacingSouth
	FacingWest
)

// Offsets for each facing
var FacingOffsets = []Point{
	{0, -1},
	{1, 0},
	{0, 1},
	{-1, 0},
}

// Bound returns a Facing value wrapped and bounded.
func (f Facing) Bound() Facing {
	return f & 0x03
}

func (f Facing) MarshalJSON() ([]byte, error) {
	switch f.Bound() {
	case FacingNorth:
		return []byte("North"), nil
	case FacingEast:
		return []byte("East"), nil
	case FacingSouth:
		return []byte("South"), nil
	default:
		return []byte("West"), nil
	}
}

func (f *Facing) UnmarshalJSON(in []byte) error {
	switch strings.ToLower(string(in)) {
	case "north":
		*f = FacingNorth
	case "east":
		*f = FacingEast
	case "south":
		*f = FacingSouth
	case "west":
		*f = FacingWest
	default:
		return fmt.Errorf("unsupported facing name %s", string(in))
	}
	return nil
}
