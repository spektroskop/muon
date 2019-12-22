package muon

import (
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"go.uber.org/zap"

	"yuki.no/muon/rect"
	"yuki.no/muon/xlib"
)

func init() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
}

type Color struct {
	String string
	Value  uint32
}

type Manager struct {
	Monitors       *MonitorList
	SelectedBorder Color
	NormalBorder   Color
	FocusedBorder  Color
	WindowGap      int
	BorderWidth    int
}

func NewManager() (*Manager, error) {
	manager := new(Manager)
	manager.Reset()
	return manager, nil
}

func NewColor(hex string) Color {
	var color Color
	color.String = hex
	color.Value = xlib.HexColor(hex)
	return color
}

func (manager *Manager) Reset() {
	manager.SelectedBorder = NewColor("#c28ccf")
	manager.NormalBorder = NewColor("#3f3e3b")
	manager.FocusedBorder = NewColor("#11809e")
	manager.WindowGap = 3
	manager.BorderWidth = 4
}

func (manager *Manager) Focused() *Monitor {
	return manager.Monitors.Focused()
}

func (manager *Manager) Manage() error {
	return xproto.ChangeWindowAttributesChecked(xlib.Conn, xlib.RootWindow, xproto.CwEventMask,
		[]uint32{xproto.EventMaskSubstructureRedirect | xproto.EventMaskSubstructureNotify},
	).Check()
}

func (manager *Manager) SetupWindows() error {
	tree, err := xproto.QueryTree(xlib.Conn, xlib.RootWindow).Reply()
	if err != nil {
		return err
	}

	for _, id := range tree.Children {
		attr, err := xproto.GetWindowAttributes(xlib.Conn, id).Reply()
		if err != nil || attr.MapState == xproto.MapStateUnmapped || attr.OverrideRedirect {
			continue
		}

		geometry, err := xproto.GetGeometry(xlib.Conn, xproto.Drawable(id)).Reply()
		if err != nil {
			continue
		}

		window := NewWindow(manager, id)
		xlib.SetBorderColor(window.Id, manager.NormalBorder.Value)

		if mon := manager.FindMonitor(int(geometry.X), int(geometry.Y)); mon != nil {
			mon.Windows.Insert(window)
		} else if mon = manager.Monitors.Focused(); mon != nil {
			mon.Windows.Insert(window)
		}
	}

	if mon := manager.Focused(); mon != nil {
		if window := mon.Focused(); window != nil {
			xlib.SetBorderColor(window.Id, manager.FocusedBorder.Value)
		}
	}

	return nil
}

func (manager *Manager) SetupMonitors() error {
	var (
		screenResources *randr.GetScreenResourcesReply
		err             error
		monitors        []*Monitor
	)

	if err := randr.Init(xlib.Conn); err != nil {
		goto fallback
	}

	if screenResources, err = randr.GetScreenResources(xlib.Conn, xlib.RootWindow).Reply(); err != nil {
		goto fallback
	}

	for _, output := range screenResources.Outputs {
		outputInfo, err := randr.GetOutputInfo(xlib.Conn, output, xproto.TimeCurrentTime).Reply()
		if err != nil {
			goto fallback
		}

		if outputInfo.Connection != randr.ConnectionConnected {
			continue
		}

	loop:
		for _, crtc := range outputInfo.Crtcs {
			crtcInfo, err := randr.GetCrtcInfo(xlib.Conn, crtc, xproto.TimeCurrentTime).Reply()
			if err != nil {
				goto fallback
			}

			geometry := rect.New(int(crtcInfo.X), int(crtcInfo.Y),
				int(crtcInfo.Width), int(crtcInfo.Height),
			)

			for _, existing := range monitors {
				if existing.Geometry.X == geometry.X && existing.Geometry.Y == geometry.Y {
					continue loop
				}
			}

			monitors = append(monitors, NewMonitor(manager, geometry))
		}
	}

	sort.Sort(SortableMonitors(monitors))

	for _, mon := range monitors {
		manager.Monitors.Insert(mon)
	}

	return randr.SelectInputChecked(xlib.Conn, xlib.RootWindow, randr.NotifyMaskScreenChange).Check()

fallback:
	manager.Monitors.Insert(NewMonitor(manager, rect.New(0, 0,
		int(xlib.DefaultScreen.WidthInPixels),
		int(xlib.DefaultScreen.HeightInPixels),
	)))

	return randr.SelectInputChecked(xlib.Conn, xlib.RootWindow, randr.NotifyMaskScreenChange).Check()
}

func (manager *Manager) Setup() error {
	manager.Monitors = NewMonitorList()

	if err := manager.SetupMonitors(); err != nil {
		return err
	}

	if err := manager.SetupWindows(); err != nil {
		return err
	}

	for _, mon := range manager.Monitors.All() {
		mon.Arrange()
	}

	return nil
}

func (manager *Manager) FindMonitor(x, y int) *Monitor {
	for _, mon := range manager.Monitors.All() {
		if mon.Geometry.Contains(x, y) {
			return mon
		}
	}

	return nil
}

func (manager *Manager) FindWindow(id xproto.Window) (*Window, *Monitor) {
	for _, mon := range manager.Monitors.All() {
		for _, win := range mon.Windows.All() {
			if win.Id == id {
				return win, mon
			}
		}
	}

	return nil, nil
}

func (manager *Manager) FindWindowString(id string) (*Window, *Monitor) {
	base := strings.Replace(id, "0x", "", -1)
	if parsed, err := strconv.ParseUint(base, 16, 32); err == nil {
		return manager.FindWindow(xproto.Window(parsed))
	}

	return nil, nil
}

func (manager *Manager) FindWindowPointer() (*Window, *Monitor) {
	queryPointer, err := xproto.QueryPointer(xlib.Conn, xlib.RootWindow).Reply()
	if err != nil {
		return nil, nil
	}

	return manager.FindWindow(queryPointer.Child)
}
