package rect

import (
	"fmt"
)

type Rect struct {
	X, Y, W, H int
}

type Padding struct {
	L, R, T, B int
}

func New(x, y, w, h int) Rect {
	return Rect{X: x, Y: y, W: w, H: h}
}

func NewPadding(l, r, t, b int) Padding {
	return Padding{L: l, R: r, T: t, B: b}
}

func (rect Rect) Pad(pad Padding) Rect {
	rect.X += pad.L
	rect.Y += pad.T
	rect.W -= pad.L + pad.R
	rect.H -= pad.T + pad.B

	return rect
}

func (rect Rect) String() string {
	return fmt.Sprintf("%dx%d+%d+%d", rect.W, rect.H, rect.X, rect.Y)
}

func (rect Rect) Contains(x, y int) bool {
	return rect.X <= x && x < rect.X+rect.W && rect.Y <= y && y < rect.Y+rect.H
}

func (rect Rect) Less(other Rect) bool {
	return rect.X < other.X || (rect.X == other.X && rect.Y < other.Y)
}
