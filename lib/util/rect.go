package util

// Rect represents integer 2D bounds.
type Rect struct {
	TL Point // Top-left point
	BR Point // Bottom-right point
}

// NewRect creates a new Bounds object from two points regardless of order.
func NewRect(a, b Point) Rect {
	top := a.Y
	if b.Y < top {
		top = b.Y
	}
	bottom := a.Y
	if b.Y > bottom {
		bottom = b.Y
	}
	left := a.X
	if b.X < left {
		left = b.X
	}
	right := a.X
	if b.X > right {
		right = b.X
	}
	return Rect{
		TL: Point{X: left, Y: top},
		BR: Point{X: right, Y: bottom},
	}
}

// NewRectWH creates a new Rect value with the given dimensions.
func NewRectWH(w, h int) Rect {
	return Rect{
		BR: Point{X: w - 1, Y: h - 1},
	}
}

// NewRectXYWH creates a new Rect value with the given dimensions and offset.
func NewRectXYWH(x, y, w, h int) Rect {
	return Rect{
		TL: Point{X: x, Y: y},
		BR: Point{X: x + w - 1, Y: y + h - 1},
	}
}

// NewRectFromRadius creates a new rect centered on point p with radius r.
func NewRectFromRadius(p Point, r int) Rect {
	return Rect{
		TL: Point{
			X: p.X - r,
			Y: p.Y - r,
		},
		BR: Point{
			X: p.X + r,
			Y: p.Y + r,
		},
	}
}

// NewRectFromExtents creates a new rect centered on point p with the given
// extents.
func NewRectFromExtents(cp Point, left, right, up, down int) Rect {
	return Rect{
		TL: Point{
			X: cp.X - left,
			Y: cp.Y - up,
		},
		BR: Point{
			X: cp.X + right,
			Y: cp.Y + down,
		},
	}
}

// Width returns the width of the rect.
func (r Rect) Width() int { return (r.BR.X - r.TL.X) + 1 }

// Height returns the height of the rect.
func (r Rect) Height() int { return (r.BR.Y - r.TL.Y) + 1 }

// Area returns the width of the rect multiplied by its height.
func (r Rect) Area() int {
	return r.Width() * r.Height()
}

// Extents returns the extents of the rect from the given centerpoint. If the
// point lies outside the rect, the result is undefined.
func (r Rect) Extents(cp Point) (left, right, top, bottom int) {
	left = cp.X - r.TL.X
	right = r.BR.X - cp.X
	top = cp.Y - r.TL.Y
	bottom = r.BR.Y - cp.Y
	return
}

// Divide divides all of the points of the rect by a.
func (r Rect) Divide(a int) Rect {
	return Rect{
		TL: r.TL.Divide(a),
		BR: r.BR.Divide(a),
	}
}

// Multiply multiplies all of the points of the rect by a.
func (r Rect) Multiply(a int) Rect {
	return Rect{
		TL: r.TL.Multiply(a),
		BR: r.BR.Multiply(a),
	}
}

// Add adds the point to both points of the rect.
func (r Rect) Add(p Point) Rect {
	return Rect{
		TL: r.TL.Add(p),
		BR: r.BR.Add(p),
	}
}

// Shrink removes n tiles from all sides of the rect.
func (r Rect) Shrink(n int) Rect {
	return Rect{
		TL: Point{
			X: r.TL.X + n,
			Y: r.TL.Y + n,
		},
		BR: Point{
			X: r.BR.X - n,
			Y: r.BR.Y - n,
		},
	}
}

// Grow removes n tiles from all sides of the rect.
func (r Rect) Grow(n int) Rect {
	return Rect{
		TL: Point{
			X: r.TL.X - n,
			Y: r.TL.Y - n,
		},
		BR: Point{
			X: r.BR.X + n,
			Y: r.BR.Y + n,
		},
	}
}

// Move moves the rect without modifying the width or height.
func (r Rect) Move(o Point) Rect {
	return Rect{
		TL: o,
		BR: Point{
			X: o.X + r.Width() - 1,
			Y: o.Y + r.Height() - 1,
		},
	}
}

// Contains returns true if the point is contained within the rect.
func (r Rect) Contains(p Point) bool {
	return p.X >= r.TL.X && p.X <= r.BR.X && p.Y >= r.TL.Y && p.Y <= r.BR.Y
}

// CenterRect returns the center a rect from the center of this rect with the
// given dimensions.
func (r Rect) CenterRect(w, h int) Rect {
	return NewRectXYWH(
		r.TL.X+(r.Width()-w)/2,
		r.TL.Y+(r.Height()-h)/2,
		w,
		h,
	)
}

// Bound bounds a point to the rect, such that the point is forced inside the
// rect along the axis where necessary.
func (r Rect) Bound(p Point) Point {
	if p.X < r.TL.X {
		p.X = r.TL.X
	}
	if p.X > r.BR.X {
		p.X = r.BR.X
	}
	if p.Y < r.TL.Y {
		p.Y = r.TL.Y
	}
	if p.Y > r.BR.Y {
		p.Y = r.BR.Y
	}
	return p
}

// Contain returns the rect contained within this rect, that is moved along the
// axis so that b is contained within r. If any of the dimensions of b are
// larger than that dimension in r the results are undefined.
func (r Rect) Contain(b Rect) Rect {
	if b.TL.X < r.TL.X {
		b.BR.X += r.TL.X - b.TL.X
		b.TL.X = r.TL.X
	}
	if b.BR.X > r.BR.X {
		b.TL.X -= b.BR.X - r.BR.X
		b.BR.X = r.BR.X
	}
	if b.TL.Y < r.TL.Y {
		b.BR.Y += r.TL.Y - b.TL.Y
		b.TL.Y = r.TL.Y
	}
	if b.BR.Y > r.BR.Y {
		b.TL.Y -= b.BR.Y - r.BR.Y
		b.BR.Y = r.BR.Y
	}
	return b
}

// Overlaps returns true if r and a overlap.
func (r Rect) Overlaps(a Rect) bool {
	// return a.BR.X < r.TL.X && a.TL.X > r.BR.X && a.BR.Y < r.TL.Y && a.TL.Y > r.BR.Y
	if r.BR.X < a.TL.X ||
		r.TL.X > a.BR.X ||
		r.BR.Y < a.TL.Y ||
		r.TL.Y > a.BR.Y {
		return false
	}
	return true
}

// Overlap returns the overlapping rect between r and a. If there is no overlap
// the zero value is returned.
func (r Rect) Overlap(a Rect) Rect {
	if !r.Overlaps(a) {
		return Rect{}
	}
	if a.TL.X < r.TL.X {
		a.TL.X = r.TL.X
	}
	if a.BR.X > r.BR.X {
		a.BR.X = r.BR.X
	}
	if a.TL.Y < r.TL.Y {
		a.TL.Y = r.TL.Y
	}
	if a.BR.Y > r.BR.Y {
		a.BR.Y = r.BR.Y
	}
	return a
}

// Center returns the centerpoint of the rect.
func (r Rect) Center() Point {
	return Point{
		X: r.TL.X + r.Width()/2,
		Y: r.TL.Y + r.Height()/2,
	}
}

// Rotate rotates the rect to the given facing about the given point, assuming
// the rect is currently facing North.
func (r Rect) Rotate(cp Point, f Facing) Rect {
	left, right, up, down := r.Extents(cp)
	switch f {
	case FacingNorth:
		return r
	case FacingEast:
		return NewRectFromExtents(cp, down, up, left, right)
	case FacingSouth:
		return NewRectFromExtents(cp, right, left, down, up)
	default:
		return NewRectFromExtents(cp, up, down, right, left)
	}
}

// RotateInPlace rotates the rect to the given facing about the top-left corner.
func (r Rect) RotateInPlace(f Facing) Rect {
	x := r.TL.X
	y := r.TL.Y
	h := r.Height()
	w := r.Width()
	switch f {
	case FacingNorth:
		return r
	case FacingEast:
		return NewRectXYWH(x-(h-1), y, h, w)
	case FacingSouth:
		return NewRectXYWH(x+(w-1), y+(h-1), w, h)
	default:
		return NewRectXYWH(x, y-(h-1), h, w)
	}
}

// RotatePoint rotates the given relative point within the rect to translate
// from the relative position at facing North to the facing given.
func (r Rect) RotatePoint(p Point, f Facing) Point {
	w := r.Width() - 1
	h := r.Height() - 1
	switch f {
	case FacingNorth:
		return p
	case FacingEast:
		return Point{
			X: h - p.Y,
			Y: p.X,
		}
	case FacingSouth:
		return Point{
			X: w - p.X,
			Y: h - p.Y,
		}
	default:
		return Point{
			X: p.Y,
			Y: w - p.X,
		}
	}
}

// ReverseRotatePoint rotates the given relative point within the rect as if the
// rect is rotated to the given facing and returns the point to a North facing.
func (r Rect) ReverseRotatePoint(p Point, f Facing) Point {
	w := r.Width() - 1
	h := r.Height() - 1
	switch f {
	case FacingNorth:
		return p
	case FacingEast:
		return Point{
			X: p.Y,
			Y: w - p.X,
		}
	case FacingSouth:
		return Point{
			X: w - p.X,
			Y: h - p.Y,
		}
	default:
		return Point{
			X: h - p.Y,
			Y: p.X,
		}
	}
}
