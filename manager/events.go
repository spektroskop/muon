package manager

import (
	"fmt"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/spektroskop/muon/node"
)

func MapRequest(event xproto.MapRequestEvent) (err error) {
	window := Manage(event.Window, nil)
	node := node.New(window)

	WithMonitorNode(Focus, func(monitor *Monitor) {
		monitor.Nodes.Append(node)
		monitor.Arrange()
		if err = xproto.MapWindowChecked(Conn, window.Id).Check(); err != nil {
			return
		}
		monitor.SetFocus(node)
	})

	return err
}

func DestroyNotify(event xproto.DestroyNotifyEvent) error {
	monitor, window := NodesFromId(event.Window)
	if window == nil {
		return fmt.Errorf("DestroyNotify: Could not find window 0x%08x", event.Window)
	}

	WithMonitorNode(monitor, func(monitor *Monitor) {
		monitor.SetFocus(window.Prev(monitor.Nodes))
		window.Unlink()
		monitor.Arrange()
	})

	return nil
}
