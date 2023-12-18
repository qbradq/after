package util

import (
	"fmt"
	"strings"
)

// Direction represents one of the eight compass rose points.
type Direction uint8

const (
	DirectionNorth Direction = iota
	DirectionNorthEast
	DirectionEast
	DirectionSouthEast
	DirectionSouth
	DirectionSouthWest
	DirectionWest
	DirectionNorthWest
)

// Bound returns a Direction value wrapped and bounded.
func (d Direction) Bound() Direction {
	return d & 0x07
}

// Facing represents one of the four cardinal directions.
type Facing uint8

const (
	FacingNorth Facing = iota
	FacingEast
	FacingSouth
	FacingWest
)

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
