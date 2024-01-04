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

// Width returns the width of the rect.
func (r Rect) Width() int { return (r.BR.X - r.TL.X) + 1 }

// Height returns the height of the rect.
func (r Rect) Height() int { return (r.BR.Y - r.TL.Y) + 1 }

// Divide divides all of the points of the rect by a.
func (r Rect) Divide(a int) Rect {
	return Rect{
		TL: r.TL.Divide(a),
		BR: r.BR.Divide(a),
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

// Overlap returns the overlapping rect between r and a. If there is no overlap
// the zero value is returned.
func (r Rect) Overlap(a Rect) Rect {
	if a.BR.X < r.TL.X || a.TL.X > r.BR.X || a.BR.Y < r.TL.Y || a.TL.Y > r.BR.Y {
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
