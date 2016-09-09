package manager

import "github.com/spektroskop/muon/util"

func Laying(m *Monitor, count int) (geom []util.Geometry) {
	geometry := m.Geometry
	size := int(float64(geometry.Height) * m.Ratio)
	border := m.Border * 2
	subs := count - m.Roots

	x := geometry.X
	y := geometry.Y
	w := geometry.Width/m.Roots - border
	h := geometry.Height - border

	if subs > 0 {
		if m.Mirrored {
			y = geometry.Y + geometry.Height - size
		}

		h = size - border
	}

	for i := 1; i <= m.Roots; i++ {
		if i == m.Roots {
			w = geometry.Width - x - border
		}

		geom = append(geom, util.Geometry{X: x, Y: y, Width: w, Height: h})
		x += geometry.Width/m.Roots + m.Gap
	}

	if subs == 0 {
		return
	}

	x = geometry.X
	y = geometry.Y + size + m.Gap
	w = geometry.Width/subs - border
	h = geometry.Height - border - size - m.Gap

	if m.Mirrored {
		y = geometry.Y
	}

	for i := 1; i <= subs; i++ {
		if i == subs {
			w = geometry.Width - x - border
		}

		geom = append(geom, util.Geometry{X: x, Y: y, Width: w, Height: h})
		x += geometry.Width/subs + m.Gap
	}

	return
}

func Standing(m *Monitor, count int) (geom []util.Geometry) {
	geometry := m.Geometry
	size := int(float64(geometry.Width) * m.Ratio)
	border := m.Border * 2
	subs := count - m.Roots

	x := geometry.X
	y := geometry.Y
	w := geometry.Width - border
	h := geometry.Height/m.Roots - border

	if subs > 0 {
		if m.Mirrored {
			x = geometry.X + geometry.Width - size
		}

		w = size - border
	}

	for i := 1; i <= m.Roots; i++ {
		if i == m.Roots {
			h = geometry.Height - y - border
		}

		geom = append(geom, util.Geometry{X: x, Y: y, Width: w, Height: h})
		y += geometry.Height/m.Roots + m.Gap
	}

	if subs == 0 {
		return
	}

	x = geometry.X + size + m.Gap
	y = geometry.Y
	w = geometry.Width - border - size - m.Gap
	h = geometry.Height/subs - border

	if m.Mirrored {
		x = geometry.X
	}

	for i := 1; i <= subs; i++ {
		if i == subs {
			h = geometry.Height - y - border
		}

		geom = append(geom, util.Geometry{X: x, Y: y, Width: w, Height: h})
		y += geometry.Height/subs + m.Gap
	}

	return
}
