package game

import (
	"io"
	"math"
	"time"

	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// VehicleLocation encapsulates all of the parts and functionality of one area
// of a vehicle.
type VehicleLocation struct {
	Parts []*Item      // Items at the location, from bottom to top
	Glyph termui.Glyph // Visual representation, if any
	Solid bool         // If true at least one part at this location is solid
}

// UpdateFlags updates the location's flags given the current contents of the
// parts list.
func (l *VehicleLocation) UpdateFlags() {
	l.Solid = false
	for _, p := range l.Parts {
		if p.VehicleSolid {
			l.Solid = true
		}
	}
}

// Add adds a part to this location.
func (l *VehicleLocation) Add(i *Item) bool {
	l.Parts = append(l.Parts, i)
	l.Glyph = termui.Glyph{
		Rune:  rune(i.Rune[0]),
		Style: termui.StyleDefault.Foreground(i.Fg).Background(i.Bg),
	}
	l.UpdateFlags()
	return true
}

// Remove removes a part from this location.
func (l *VehicleLocation) Remove(i *Item) bool {
	idx := -1
	for n, p := range l.Parts {
		if p == i {
			idx = n
			break
		}
	}
	if idx < 0 {
		return false
	}
	// Remove from slice while maintaining order
	copy(l.Parts[idx:], l.Parts[idx+1:])
	l.Parts[len(l.Parts)-1] = nil
	l.Parts = l.Parts[:len(l.Parts)-1]
	l.UpdateFlags()
	if len(l.Parts) > 0 {
		i := l.Parts[0]
		l.Glyph = termui.Glyph{
			Rune:  rune(i.Rune[0]),
			Style: termui.StyleDefault.Foreground(i.Fg).Background(i.Bg),
		}
	}
	return true
}

// AccelerationState represents the state of the vehicle's acceleration.
type AccelerationState uint8

const (
	AccelerationStateIdle         AccelerationState = 0 // Not accelerating or decelerating
	AccelerationStateAccelerating AccelerationState = 1 // Gaining speed
	AccelerationStateDecelerating AccelerationState = 2 // Slowing down / reversing
)

// TurningState represents the state of the vehicle's turning controls.
type TurningState uint8

const (
	TurningStateNone  TurningState = 0 // Not turning
	TurningStateRight TurningState = 1 // Turning to the right
	TurningStateLeft  TurningState = 2 // Turning to the left
)

// Vehicle contains all of the parts and functionality of a vehicle.
type Vehicle struct {
	Name         string            // Name of the vehicle
	Size         util.Point        // Width and height of the vehicle
	Bounds       util.Rect         // Current bounds in the city
	Facing       util.Facing       // Current facing
	Locations    []VehicleLocation // All of the locations of the vehicle
	Speed        float64           // Forward speed in scale miles per hour
	TopSpeed     float64           // Top speed in scale miles per hour
	Acceleration float64           // Forward acceleration in scale miles per hour per second
	Heading      util.Direction    // Direction of movement
	stp          float64           // Sub-tile position

	//
	// Non-persistent values
	//

	AccelerationState AccelerationState // Acceleration state
	TurningState      TurningState      // Turning state
}

// newVehicle returns a new vehicle with the given parameters.
func newVehicle(size util.Point) *Vehicle {
	ret := &Vehicle{
		Size:      size,
		Bounds:    util.NewRectWH(size.X, size.Y),
		Locations: make([]VehicleLocation, size.X*size.Y),
	}
	return ret
}

// NewVehicleFromReader reads a vehicle from a reader.
func NewVehicleFromReader(r io.Reader) *Vehicle {
	util.GetUint32(r)     // Version
	p := util.GetPoint(r) // Position
	s := util.GetPoint(r) // Size
	v := newVehicle(s)
	f := util.Facing(util.GetByte(r)) // Facing
	v.UpdateBoundsForPosition(p, f)
	v.Speed = util.GetFloat(r)                  // Forward speed
	v.Heading = util.Direction(util.GetByte(r)) // Movement heading
	v.stp = util.GetFloat(r)                    // Sub-tile position
	for idx := 0; idx < v.Size.X*v.Size.Y; idx++ {
		nParts := int(util.GetByte(r))            // Number of parts
		for iPart := 0; iPart < nParts; iPart++ { // Parts
			v.Locations[idx].Parts = append(v.Locations[idx].Parts, NewItemFromReader(r))
		}
	}
	return v
}

// GenerateVehicle generates a new vehicle from the named group.
func GenerateVehicle(gn string, now time.Time) *Vehicle {
	g, found := VehicleGenGroups[gn]
	if !found {
		Log.Log(termui.ColorRed, "Vehicle group %s not found.", gn)
		return nil
	}
	return g.Get().Generate(now)
}

// Write writes the vehicle to the writer.
func (v *Vehicle) Write(w io.Writer) {
	util.PutUint32(w, 0)             // Version
	util.PutPoint(w, v.Bounds.TL)    // Position
	util.PutPoint(w, v.Size)         // North-facing dimensions
	util.PutByte(w, byte(v.Facing))  // Facing
	util.PutFloat(w, v.Speed)        // Forward speed
	util.PutByte(w, byte(v.Heading)) // Movement heading
	util.PutFloat(w, v.stp)          // Sub-tile position
	for _, l := range v.Locations {
		util.PutByte(w, byte(len(l.Parts))) // Number of parts at this location
		for _, p := range l.Parts {         // Parts
			p.Write(w)
		}
	}
}

// Attach attached the given item as a part to the vehicle at the given relative
// offset.
func (v *Vehicle) Attach(i *Item, p util.Point) bool {
	if i == nil {
		return false
	}
	if !util.NewRectWH(v.Size.X, v.Size.Y).Contains(p) {
		return false
	}
	idx := p.Y*v.Size.X + p.X
	v.Locations[idx].Add(i)
	return true
}

// Remove removes the given item as a part from the vehicle.
func (v *Vehicle) Remove(i *Item) bool {
	if i == nil {
		return false
	}
	for _, l := range v.Locations {
		if l.Remove(i) {
			return true
		}
	}
	return false
}

// GetLocationRelative returns a pointer to the VehicleLocation for the given
// relative position and the current facing.
func (v *Vehicle) GetLocationRelative(rp util.Point) *VehicleLocation {
	if rp.X < 0 || rp.Y < 0 || rp.X >= v.Bounds.Width() || rp.Y >= v.Bounds.Height() {
		return nil
	}
	lp := v.Bounds.ReverseRotatePoint(rp, v.Facing)
	return &v.Locations[lp.Y*v.Size.X+lp.X]
}

// GetLocationAbsolute returns a pointer to the VehicleLocation for the given
// absolute map position and the current facing.
func (v *Vehicle) GetLocationAbsolute(ap util.Point) *VehicleLocation {
	rp := ap.Sub(v.Bounds.TL)
	if rp.X < 0 || rp.Y < 0 || rp.X >= v.Bounds.Width() || rp.Y >= v.Bounds.Height() {
		return nil
	}
	lp := v.Bounds.ReverseRotatePoint(rp, v.Facing)
	return &v.Locations[lp.Y*v.Size.X+lp.X]
}

// UpdateBoundsForPosition updates the vehicle based on the given position and
// facing.
func (v *Vehicle) UpdateBoundsForPosition(p util.Point, f util.Facing) {
	v.Facing = f
	v.Heading = f.Direction()
	v.Bounds = v.Bounds.Move(p)
	v.Bounds = v.boundsForFacing(f)
}

// boundsForFacing returns the correct vehicle bounds for the given facing.
func (v *Vehicle) boundsForFacing(f util.Facing) util.Rect {
	return util.NewRectXYWH(v.Bounds.TL.X, v.Bounds.TL.Y, v.Size.X, v.Size.Y).RotateInPlace(f)
}

// Update handles short term updates for vehicles.
func (v *Vehicle) Update(d time.Duration, cm *CityMap) {
	// Handle acceleration and deceleration
	switch v.AccelerationState {
	case AccelerationStateAccelerating:
		v.Speed += (float64(d) / float64(time.Second)) * v.Acceleration
		if v.Speed > v.TopSpeed {
			v.Speed = v.TopSpeed
		}
	case AccelerationStateDecelerating:
		m := v.Acceleration * 2
		if v.Speed <= 0 {
			m = v.Acceleration / 4
		}
		v.Speed -= (float64(d) / float64(time.Second)) * m
		if v.Speed < -v.TopSpeed/4 {
			v.Speed = -v.TopSpeed / 4
		}
	case AccelerationStateIdle:
		if v.Speed > 0 {
			v.Speed -= float64(d) / float64(time.Second*2)
			if v.Speed < 0 {
				v.Speed = 0
			}
		} else if v.Speed < 0 {
			v.Speed += float64(d) / float64(time.Second*2)
			if v.Speed > 0 {
				v.Speed = 0
			}
		}
	}
	// Handle turning
	doTurn := false
	oh := v.Heading
	switch v.TurningState {
	case TurningStateRight:
		v.Heading++
		doTurn = true
	case TurningStateLeft:
		v.Heading--
		doTurn = true
	}
	if doTurn {
		v.Heading = v.Heading.Bound()
		if !v.Heading.IsDiagonal() {
			nf := v.Heading.Facing()
			nb := v.boundsForFacing(nf)
			if cm.VehicleFits(v, nb) {
				v.UpdateBoundsForPosition(v.Bounds.TL, nf)
			} else {
				v.Heading = oh
			}
		}
	}
	// Handle movement
	mt := math.Abs(v.Speed) * (float64(d) / float64(time.Hour)) // Miles traveled
	v.stp += mt * 1760 / 4                                      // Tiles traveled
	ofs := util.DirectionOffsets[v.Heading.Bound()]
	if v.Speed < 0 {
		ofs = ofs.Multiply(-1)
	}
	for ; v.stp >= 1; v.stp -= 1 {
		if !cm.MoveVehicle(v, ofs) {
			v.stp = 0
			v.Speed = 0
			return
		}
	}
}
