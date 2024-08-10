package util

// Direction represents one of the eight compass rose points.
type Direction byte

const (
	DirectionNorth Direction = iota
	DirectionNorthEast
	DirectionEast
	DirectionSouthEast
	DirectionSouth
	DirectionSouthWest
	DirectionWest
	DirectionNorthWest
	DirectionInvalid Direction = 0xFF
)

// Offsets for each direction
var DirectionOffsets = []Point{
	{0, -1},
	{1, -1},
	{1, 0},
	{1, 1},
	{0, 1},
	{-1, 1},
	{-1, 0},
	{-1, -1},
}

// IsDiagonal returns true if the direction is one of the diagonals.
func (d Direction) IsDiagonal() bool {
	return uint8(d)&0x01 != 0
}

// Bound returns a Direction value wrapped and bounded.
func (d Direction) Bound() Direction {
	return d & 0x07
}

// RotateClockwise rotates the direction clockwise by n steps.
func (d Direction) RotateClockwise(n int) Direction {
	return Direction((int(d) + n)) % 8
}

// RotateCounterclockwise rotates the direction counterclockwise by n steps.
func (d Direction) RotateCounterclockwise(n int) Direction {
	return Direction((int(d) - n)) % 8
}

// Facing returns the equivalent facing for this direction.
func (d Direction) Facing() Facing {
	switch d {
	case DirectionNorth:
		fallthrough
	case DirectionNorthEast:
		return FacingNorth
	case DirectionEast:
		fallthrough
	case DirectionSouthEast:
		return FacingEast
	case DirectionSouth:
		fallthrough
	case DirectionSouthWest:
		return FacingSouth
	default:
		return FacingWest
	}
}
