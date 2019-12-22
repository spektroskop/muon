package muon

import "yuki.no/muon/rect"

type Layout interface {
	String() string
	Layout(monitor *Monitor, windowCount int) []rect.Rect
}

type HorizontalLayout struct{}

func NewHorizontalLayout() HorizontalLayout {
	return HorizontalLayout{}
}

func (_ HorizontalLayout) String() string {
	return "horizontal"
}

func (_ HorizontalLayout) Layout(monitor *Monitor, windowCount int) []rect.Rect {
	var (
		windows   []rect.Rect
		geometry  = monitor.Geometry.Pad(monitor.Padding)
		rootCount = monitor.RootCount
		subCount  = windowCount - rootCount
		border    = monitor.BorderWidth * 2
	)

	if windowCount == 0 {
		return windows
	}

	if windowCount == 1 {
		return append(windows, geometry)
	}

	if rootCount > windowCount {
		rootCount = windowCount
	}

	var x, y, w, h, s, r int

	r = int(float64(geometry.H) * monitor.Ratio)
	s = geometry.W / monitor.RootCount
	x = geometry.X
	y = geometry.Y
	w = s - border
	h = geometry.H - border

	if monitor.Mirrored && subCount > 0 {
		y = geometry.Y + geometry.H - r
	}

	if subCount > 0 {
		h = r - border
	}

	for i := 1; i <= rootCount; i++ {
		if i == rootCount {
			w = geometry.W - (x - geometry.X) - border
		}

		windows = append(windows, rect.New(x, y, w, h))
		x += s + monitor.WindowGap
	}

	if subCount == 0 {
		return windows
	}

	s = geometry.W / subCount
	x = geometry.X
	y = geometry.Y + r + monitor.WindowGap
	w = s - border
	h = geometry.H - border - r - monitor.WindowGap

	if monitor.Mirrored {
		y = geometry.Y
	}

	for i := 1; i <= subCount; i++ {
		if i == subCount {
			w = geometry.W - (x - geometry.X) - border
		}

		windows = append(windows, rect.New(x, y, w, h))
		x += s + monitor.WindowGap
	}

	return windows
}

type VerticalLayout struct{}

func NewVerticalLayout() VerticalLayout {
	return VerticalLayout{}
}

func (_ VerticalLayout) String() string {
	return "vertical"
}

func (_ VerticalLayout) Layout(monitor *Monitor, windowCount int) []rect.Rect {
	var (
		windows   []rect.Rect
		geometry  = monitor.Geometry.Pad(monitor.Padding)
		rootCount = monitor.RootCount
		subCount  = windowCount - rootCount
		border    = monitor.BorderWidth * 2
	)

	if windowCount == 0 {
		return windows
	}

	if windowCount == 1 {
		return append(windows, geometry)
	}

	if rootCount > windowCount {
		rootCount = windowCount
	}

	var x, y, w, h, s, r int

	r = int(float64(geometry.W) * monitor.Ratio)
	s = geometry.H / monitor.RootCount
	x = geometry.X
	y = geometry.Y
	w = geometry.W - border
	h = s - border

	if monitor.Mirrored && subCount > 0 {
		x = geometry.X + geometry.W - r
	}

	if subCount > 0 {
		w = r - border
	}

	for i := 1; i <= rootCount; i++ {
		if i == rootCount {
			h = geometry.H - (y - geometry.Y) - border
		}

		windows = append(windows, rect.New(x, y, w, h))
		y += s + monitor.WindowGap
	}

	if subCount == 0 {
		return windows
	}

	s = geometry.H / subCount
	x = geometry.X + r + monitor.WindowGap
	y = geometry.Y
	w = geometry.W - border - r - monitor.WindowGap
	h = s - border

	if monitor.Mirrored {
		x = geometry.X
	}

	for i := 1; i <= subCount; i++ {
		if i == subCount {
			h = geometry.H - (y - geometry.Y) - border
		}

		windows = append(windows, rect.New(x, y, w, h))
		y += s + monitor.WindowGap
	}

	return windows
}
