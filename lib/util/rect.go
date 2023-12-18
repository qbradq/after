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

// Width returns the width of the rect.
func (r Rect) Width() int { return (r.BR.X - r.TL.X) + 1 }

// Height returns the height of the rect.
func (r Rect) Height() int { return (r.BR.Y - r.TL.Y) + 1 }

// Contains returns true if the point is contained within the rect.
func (r Rect) Contains(p Point) bool {
	return p.X >= r.TL.X && p.X <= r.BR.X && p.Y >= r.TL.Y && p.Y <= r.BR.Y
}
