package manager

import (
	"fmt"

	"github.com/BurntSushi/xgb/xproto"
)

func MapRequest(event xproto.MapRequestEvent) (err error) {
	node := Manage(event.Window, nil)

	WithWindowNode(node, func(window *Window) {
		WithMonitorNode(Focus, func(monitor *Monitor) {
			node.Link(monitor.Focus)
			monitor.Arrange()
			if err = xproto.MapWindowChecked(Conn, window.Id).Check(); err != nil {
				return
			}
			monitor.SetFocus(node)
		})
	})

	return err
}

func DestroyNotify(event xproto.DestroyNotifyEvent) error {
	monitor, window := NodesFromId(event.Window)
	if window == nil {
		return fmt.Errorf("DestroyNotify: Could not find window 0x%08x", event.Window)
	}

	WithMonitorNode(monitor, func(monitor *Monitor) {
		monitor.SetFocus(window.Prev())
		window.Unlink()
		monitor.Arrange()
	})

	return nil
}
