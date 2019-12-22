package main

import (
	"fmt"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"go.uber.org/zap"

	"yuki.no/muon/muon"
	"yuki.no/muon/xlib"
)

func handleEvent(manager *muon.Manager, base xgb.Event, previousWindow **muon.Window) {
	switch event := base.(type) {
	case randr.ScreenChangeNotifyEvent:
		zap.S().Infow("event", "type", "ScreenChangeNotify")
		// manager.Setup()

	case xproto.ConfigureNotifyEvent:
		zap.S().Infow("event", "type", "ConfigureNotify", "window", fmt.Sprintf("0x%08x", event.Window))

		if event.Window == xlib.RootWindow {
			// manager.Setup()
			return
		}

		if win, _ := manager.FindWindow(event.Window); win != nil {
			win.Name = xlib.GetString(event.Window, xlib.NameAtom)
		}

	case xproto.ConfigureRequestEvent:
		zap.S().Infow("event", "type", "ConfigureRequest", "window", fmt.Sprintf("0x%08x", event.Window))

		if win, mon := manager.FindWindow(event.Window); mon != nil && win != nil {
			xlib.ConfigureFromGeometry(event.Window, mon.BorderWidth, win.Geometry)
		} else {
			xlib.ConfigureFromRequest(event)
		}

	case xproto.UnmapNotifyEvent:
		zap.S().Infow("event", "type", "UnmapNotify", "window", fmt.Sprintf("0x%08x", event.Window))

		if win, mon := manager.FindWindow(event.Window); mon != nil || win != nil {
			mon.Windows.RemoveMatch(win)
			mon.Arrange()
			if *previousWindow == win {
				*previousWindow = nil
			}
		}

	case xproto.MapRequestEvent:
		zap.S().Infow("event", "type", "MapRequest", "window", fmt.Sprintf("0x%08x", event.Window))

		attr, err := xproto.GetWindowAttributes(xlib.Conn, event.Window).Reply()
		if err != nil || attr.OverrideRedirect {
			return
		}

		if win, _ := manager.FindWindow(event.Window); win != nil {
			return
		}

		window := muon.NewWindow(manager, event.Window)
		xlib.SetBorderColor(window.Id, manager.NormalBorder.Value)
		defer xlib.MapWindow(window.Id)

		if transient := xlib.GetWindow(event.Window, xlib.TransientForAtom); transient != 0 {
			if win, mon := manager.FindWindow(transient); mon != nil && win != nil {
				mon.Windows.InsertAfterFocus(window)
				if !mon.Fullscreen && mon.Focused() == win {
					mon.Windows.FocusMatch(window)
				}
				mon.Arrange()
				break
			}
		}

		if mon := manager.Focused(); mon != nil {
			mon.Windows.Insert(window)
			mon.Arrange()
		}
	}

	return
}
