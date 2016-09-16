package manager

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/Sirupsen/logrus"
	"github.com/spektroskop/muon/nd"
	"github.com/spektroskop/muon/util"
)

type Window struct {
	Id xproto.Window
}

func Manage(id xproto.Window, attr *xproto.GetWindowAttributesReply) *nd.Node {
	if attr == nil {
		var err error
		attr, err = xproto.GetWindowAttributes(Conn, id).Reply()
		if err != nil {
			logrus.Errorf("Could not get attributes for 0x%08x: %s",
				id, err,
			)
			return nil
		}
	}

	if attr.OverrideRedirect {
		logrus.Infof("Ignoring window 0x%08x: override redirect", id)
		return nil
	}

	reply, err := xproto.GetGeometry(Conn, xproto.Drawable(id)).Reply()
	if err != nil {
		logrus.Errorf("Could not get geometry for window 0x%08x: %s", id, err)
		return nil
	}

	monitor, _ := MonitorFromCoordinates(int(reply.X), int(reply.Y))
	if monitor == nil {
		logrus.Errorf("Could not find monitor for window 0x%08x", id)
		return nil
	}

	logrus.Debugf("Manage window 0x%08x", id)

	window := &Window{Id: id}
	node := nd.New(window)
	window.SetBorderColor(InactiveBorder)

	if monitor.Focus == nil {
		monitor.Root = node
		monitor.SetFocus(node)
	} else {
		node.Link(monitor.Focus)
	}

	return node
}

func (window *Window) SetBorderColor(color uint32) {
	xproto.ChangeWindowAttributes(Conn, window.Id, xproto.CwBorderPixel,
		[]uint32{color},
	)
}

func (window *Window) SetBorderWidth(width int) {
	xproto.ConfigureWindow(Conn, window.Id, xproto.ConfigWindowBorderWidth,
		[]uint32{uint32(width)},
	)
}

func (window *Window) SetGeometry(g util.Geometry) {
	values := []uint32{uint32(g.X), uint32(g.Y), uint32(g.Width), uint32(g.Height)}
	flags := xproto.ConfigWindowX | xproto.ConfigWindowY |
		xproto.ConfigWindowHeight | xproto.ConfigWindowWidth
	xproto.ConfigureWindow(Conn, window.Id, uint16(flags), values)
}

func (window *Window) Focus() {
	xproto.SetInputFocus(Conn, xproto.InputFocusPointerRoot, window.Id,
		xproto.TimeCurrentTime,
	)
}
