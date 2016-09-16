package manager

import (
	"fmt"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/Sirupsen/logrus"
	"github.com/spektroskop/muon/nd"
	"github.com/spektroskop/muon/util"
)

var (
	Conn     *xgb.Conn
	Screen   *xproto.ScreenInfo
	Colormap xproto.Colormap
	Root     xproto.Window
	Focus    *nd.Node

	ActiveBorder   uint32
	InactiveBorder uint32
	UrgentBorder   uint32
)

func New() error {
	return util.Do([]util.F{
		setup,
		register,
		monitor,
		manage,
		listen,
	})
}

func setup() error {
	conn, err := xgb.NewConn()
	if err != nil {
		return err
	}

	Conn = conn
	Screen = xproto.Setup(conn).DefaultScreen(conn)
	Colormap = Screen.DefaultColormap
	Root = Screen.Root

	ActiveBorder = ParseColor("#11809e")
	InactiveBorder = ParseColor("#3f3e3b")
	UrgentBorder = ParseColor("#cc3300")

	return nil
}

func register() error {
	mask := xproto.EventMaskSubstructureRedirect |
		xproto.EventMaskSubstructureNotify
	cookie := xproto.ChangeWindowAttributesChecked(
		Conn, Root, xproto.CwEventMask, []uint32{uint32(mask)},
	)

	return cookie.Check()
}

func monitor() error {
	randr.Init(Conn)

	reply, err := randr.GetScreenResources(Conn, Root).Reply()
	if err != nil {
		return err
	}

	for _, output := range reply.Outputs {
		outputInfo, err := randr.GetOutputInfo(Conn, output, xproto.TimeCurrentTime).Reply()
		if err != nil {
			return err
		}

		for _, crtc := range outputInfo.Crtcs {
			crtcInfo, err := randr.GetCrtcInfo(Conn, crtc, xproto.TimeCurrentTime).Reply()
			if err != nil {
				return err
			}

			geometry := util.Geometry{
				X:      int(crtcInfo.X),
				Y:      int(crtcInfo.Y),
				Width:  int(crtcInfo.Width),
				Height: int(crtcInfo.Height),
			}
			monitor := NewMonitor(string(outputInfo.Name), geometry)
			node := nd.New(monitor)

			if Focus == nil {
				Focus = node
			} else {
				node.Link(Focus)
			}

			logrus.Debugf("Monitor `%s' %s", monitor.Name, geometry)
		}
	}

	return randr.SelectInputChecked(Conn, Root, 0 /* TODO */).Check()
}

func manage() error {
	tree, err := xproto.QueryTree(Conn, Root).Reply()
	if err != nil {
		return err
	}

	for _, id := range tree.Children {
		attr, err := xproto.GetWindowAttributes(Conn, id).Reply()
		if err != nil {
			logrus.Debugf("Could not get attributes for 0x%08x: %s", id, err)
			continue
		} else if attr.MapState == xproto.MapStateUnmapped {
			logrus.Debugf("Ignoring unmapped window: 0x%08x", id)
		}

		logrus.Debugf("Existing window 0x%08x", id)

		Manage(id, attr)
	}

	for node := range Focus.Each() {
		monitor := node.Value.(*Monitor)
		monitor.Arrange()
	}

	return nil
}

func listen() error {
	go func() {
		for {
			event, err := Conn.WaitForEvent()
			if err != nil {
				logrus.Errorf("Event error: %s", err)
				continue
			}

			logrus.Debugf("Event: %s", event)

			switch actual := event.(type) {
			case xproto.MapRequestEvent:
				if err := MapRequest(actual); err != nil {
					logrus.Errorf("MapRequest: %s", err)
				}
			case xproto.DestroyNotifyEvent:
				if err := DestroyNotify(actual); err != nil {
					logrus.Errorf("DestroyNotify: %s", err)
				}
			}
		}
	}()

	return nil
}

func NodesFromPointer() (*nd.Node, *nd.Node) {
	reply, err := xproto.QueryPointer(Conn, Root).Reply()
	if err != nil {
		logrus.Errorf("Could not query pointer: %s", err)
		return nil, nil
	}

	return NodesFromId(reply.Child)
}

func NodesFromId(id xproto.Window) (*nd.Node, *nd.Node) {
	for monitorNode := range Focus.Each() {
		var windowNode *nd.Node
		WithMonitorNode(monitorNode, func(monitor *Monitor) {
			windowNode = monitor.NodeFromId(id)
		})

		return monitorNode, windowNode
	}

	return nil, nil
}

func ParseColor(color string) uint32 {
	var r, g, b uint16
	fmt.Sscanf(color, "#%02x%02x%02x", &r, &g, &b)
	r *= 0x101
	g *= 0x101
	b *= 0x101

	reply, err := xproto.AllocColor(Conn, Colormap, r, g, b).Reply()
	if err != nil {
		return 0
	}

	return reply.Pixel
}
