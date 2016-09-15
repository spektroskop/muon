package manager

import (
	"github.com/Sirupsen/logrus"
	"github.com/spektroskop/muon/node"
)

func MonitorFromCoordinates(x, y int) (*Monitor, *node.Node) {
	for _, node := range Nodes.All() {
		monitor := node.Value.(*Monitor)
		if monitor.Geometry.Contains(x, y) {
			return monitor, node
		}
	}

	return nil, nil
}

func WithMonitorNode(node *node.Node, f func(monitor *Monitor)) {
	if monitor, ok := node.Value.(*Monitor); !ok {
		logrus.Error("Could not get monitor from node")
	} else {
		f(monitor)
	}
}

func WithFocus(f func(*Monitor)) {
	WithMonitorNode(Focus, f)
}

func WithWindowNode(node *node.Node, f func(window *Window)) {
	if window, ok := node.Value.(*Window); !ok {
		logrus.Error("Could not get window from node")
	} else {
		f(window)
	}
}
