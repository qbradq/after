package game

import (
	"io"
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

// Vehicle contains all of the parts and functionality of a vehicle.
type Vehicle struct {
	Name      string            // Name of the vehicle
	Size      util.Point        // Width and height of the vehicle
	Bounds    util.Rect         // Current bounds in the city
	Facing    util.Facing       // Current facing
	Locations []VehicleLocation // All of the locations of the vehicle
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
	}
	return g.Get().Generate(now)
}

// Write writes the vehicle to the writer.
func (v *Vehicle) Write(w io.Writer) {
	util.PutUint32(w, 0)          // Version
	util.PutPoint(w, v.Bounds.TL) // Position
	if v.Facing == util.FacingEast || v.Facing == util.FacingWest {
		util.PutPoint(w, util.NewPoint(v.Size.Y, v.Size.X))
	} else {
		util.PutPoint(w, util.NewPoint(v.Size.X, v.Size.Y))
	}
	util.PutByte(w, byte(v.Facing)) // Facing
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
	v.Bounds = util.NewRectXYWH(p.X, p.Y, v.Size.X, v.Size.Y)
	v.Bounds = v.Bounds.RotateInPlace(v.Facing)
}
