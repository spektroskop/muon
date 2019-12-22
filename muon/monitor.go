package muon

import (
	"fmt"

	"go.uber.org/zap"

	"yuki.no/muon/rect"
	"yuki.no/muon/xlib"
)

type Monitor struct {
	Name        string
	Windows     *WindowList
	Layouts     *LayoutList
	Geometry    rect.Rect
	Fullscreen  bool
	Mirrored    bool
	Ratio       float64
	RootCount   int
	WindowGap   int
	BorderWidth int
	Padding     rect.Padding
}

type SortableMonitors []*Monitor

func (mon SortableMonitors) Swap(i int, j int)      { mon[i], mon[j] = mon[j], mon[i] }
func (mon SortableMonitors) Len() int               { return len(mon) }
func (mon SortableMonitors) Less(i int, j int) bool { return mon[i].Geometry.Less(mon[j].Geometry) }

func (m *Monitor) String() string {
	if m.Name == "" {
		return fmt.Sprintf("%s", m.Geometry)
	} else {
		return fmt.Sprintf("%s %s", m.Name, m.Geometry)
	}
}

func NewMonitor(manager *Manager, geometry rect.Rect) *Monitor {
	mon := new(Monitor)

	mon.Windows = NewWindowList()
	mon.Fullscreen = false
	mon.Padding = rect.NewPadding(0, 0, 0, 0)
	mon.WindowGap = manager.WindowGap
	mon.BorderWidth = manager.BorderWidth
	mon.Geometry = geometry

	zap.S().Infow("monitor", "geometry", mon)

	mon.Reset()
	return mon
}

func (mon *Monitor) Focused() *Window {
	return mon.Windows.Focused()
}

func (mon *Monitor) Reset() {
	mon.Layouts = NewLayoutList()
	mon.Layouts.Insert(NewVerticalLayout())
	mon.Layouts.Insert(NewHorizontalLayout())

	mon.Mirrored = false
	mon.Ratio = 0.65
	mon.RootCount = 1
}

func (mon *Monitor) Arrange() {
	layout := mon.Layouts.Focused()

	zap.S().Infow("arrange", "layout", layout)

	if layout == nil || mon.Fullscreen {
		if win := mon.Focused(); win != nil {
			geometry := mon.Geometry.Pad(mon.Padding)
			win.Geometry = geometry

			xlib.SetGeometry(win.Id, geometry)
			xlib.ConfigureFromGeometry(win.Id, 0, geometry)
			xlib.SetBorderWidth(win.Id, 0)
			xlib.Raise(win.Id)
		}

		return
	}

	geometries := layout.Layout(mon, mon.Windows.Len())
	borderWidth := mon.BorderWidth
	if len(geometries) == 1 {
		borderWidth = 0
	}

	for i, win := range mon.Windows.All() {
		xlib.SetBorderWidth(win.Id, borderWidth)
		xlib.SetGeometry(win.Id, geometries[i])
		xlib.ConfigureFromGeometry(win.Id, mon.BorderWidth, geometries[i])

		win.Geometry = geometries[i]
		zap.S().Infow("arrange", "window", win, "geometry", win.Geometry, "root", i < mon.RootCount)
	}
}
