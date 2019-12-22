package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/alexflint/go-arg"
	"go.uber.org/zap"

	"yuki.no/muon/muon"
	"yuki.no/muon/xlib"
)

func init() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
}

type Request struct {
	net.Conn
	Command string
	Args    []string
}

type Command interface {
	Run(Request, *muon.Manager, *muon.Monitor, *muon.Window, *muon.Window) *muon.Window
}

func isCount(arg string) bool {
	return arg != "" && strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "+")
}

func isHex(arg string) bool {
	return arg != "" && strings.HasPrefix(arg, "0x")
}

func prepareCommand(req Request) (Command, error) {
	var cmd Command

	switch req.Command {
	case "selected-border":
		cmd = &SelectedBorderCmd{}
	case "normal-border":
		cmd = &NormalBorderCmd{}
	case "focused-border":
		cmd = &FocusedBorderCmd{}
	case "border-width":
		cmd = &BorderWidthCmd{}
	case "window-gap":
		cmd = &WindowGapCmd{}
	case "root-count":
		cmd = &RootCountCmd{}
	case "ratio":
		cmd = &RatioCmd{}
	case "padding":
		cmd = &PaddingCmd{}
	case "fullscreen":
		cmd = &FullscreenCmd{}
	case "select-layout":
		cmd = &SelectLayoutCmd{}
	case "reset-layout":
		cmd = &ResetLayoutCmd{}
	case "mirror-layout":
		cmd = &MirrorLayoutCmd{}
	case "focus-monitor":
		cmd = &FocusMonitorCmd{}
	case "focus-window":
		cmd = &FocusWindowCmd{}
	case "select-window":
		cmd = &SelectWindowCmd{}
	case "root-window":
		cmd = &RootWindowCmd{}
	case "move-window":
		cmd = &MoveWindowCmd{}
	case "close-window":
		cmd = &CloseWindowCmd{}
	default:
		return nil, fmt.Errorf("command not found: %s", req.Command)
	}

	parser, err := arg.NewParser(arg.Config{Program: req.Command}, cmd)
	if err != nil {
		return nil, err
	}

	return cmd, parser.Parse(req.Args)
}

func runCommand(manager *muon.Manager, req Request, focusedMonitor *muon.Monitor, focusedWindow, selectedWindow *muon.Window) *muon.Window {
	defer req.Close()

	if focusedMonitor == nil {
		return selectedWindow
	}

	zap.S().Infow("request", "command", req.Command, "args", req.Args)

	cmd, err := prepareCommand(req)
	if err != nil {
		fmt.Fprintln(req, err.Error())
		return selectedWindow
	}

	return cmd.Run(req, manager, focusedMonitor, focusedWindow, selectedWindow)
}

// selected-border [color]

type SelectedBorderCmd struct {
	Color string `arg:"positional"`
}

func (cmd SelectedBorderCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case cmd.Color == "":
		fmt.Fprintln(req, manager.SelectedBorder.String)

	case cmd.Color != "" && selectedWindow != nil:
		manager.SelectedBorder = muon.NewColor(cmd.Color)
		xlib.SetBorderColor(selectedWindow.Id, manager.SelectedBorder.Value)

	case cmd.Color != "":
		manager.SelectedBorder = muon.NewColor(cmd.Color)
	}

	return selectedWindow
}

// normal-border [color]

type NormalBorderCmd struct {
	Color string `arg:"positional"`
}

func (cmd NormalBorderCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case cmd.Color == "":
		fmt.Fprintln(req, manager.NormalBorder.String)

	case cmd.Color != "":
		manager.NormalBorder = muon.NewColor(cmd.Color)

		for _, mon := range manager.Monitors.All() {
			for _, win := range mon.Windows.All() {
				if win != mon.Focused() && win != selectedWindow {
					xlib.SetBorderColor(win.Id, manager.NormalBorder.Value)
				}
			}
		}
	}

	return selectedWindow
}

// focused-border [color]

type FocusedBorderCmd struct {
	Color string `arg:"positional"`
}

func (cmd FocusedBorderCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case cmd.Color == "":
		fmt.Fprintln(req, manager.FocusedBorder.String)

	case cmd.Color != "":
		manager.FocusedBorder = muon.NewColor(cmd.Color)

		if mon := manager.Focused(); mon != nil {
			if win := mon.Focused(); win != nil {
				xlib.SetBorderColor(win.Id, manager.FocusedBorder.Value)
			}
		}
	}

	return selectedWindow
}

// border-width [-default] [size]

type BorderWidthCmd struct {
	Default bool
	Size    string `arg:"positional"`
}

func (cmd BorderWidthCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case cmd.Size == "" && cmd.Default:
		fmt.Fprintln(req, manager.BorderWidth)

	case cmd.Size == "":
		fmt.Fprintln(req, focusedMonitor.BorderWidth)

	case cmd.Size != "" && cmd.Default:
		if value, err := strconv.Atoi(cmd.Size); err == nil {
			manager.BorderWidth = value
		}

	case cmd.Size != "":
		if value, err := strconv.Atoi(cmd.Size); err == nil {
			focusedMonitor.BorderWidth = value
			focusedMonitor.Arrange()
		}
	}

	return selectedWindow
}

// window-gap [-default] [size]

type WindowGapCmd struct {
	Default bool
	Size    string `arg:"positional"`
}

func (cmd WindowGapCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case cmd.Size == "" && cmd.Default:
		fmt.Fprintln(req, manager.WindowGap)

	case cmd.Size == "":
		fmt.Fprintln(req, focusedMonitor.WindowGap)

	case cmd.Size != "" && cmd.Default:
		if value, err := strconv.Atoi(cmd.Size); err == nil {
			manager.WindowGap = value
		}

	case cmd.Size != "":
		if value, err := strconv.Atoi(cmd.Size); err == nil {
			focusedMonitor.WindowGap = value
			focusedMonitor.Arrange()
		}

	}

	return selectedWindow
}

// root-count [-N|+N|count]

type RootCountCmd struct {
	Count string `arg:"positional"`
}

func (cmd RootCountCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case cmd.Count == "":
		fmt.Fprintln(req, focusedMonitor.RootCount)

	case isCount(cmd.Count):
		if count, err := strconv.Atoi(cmd.Count); err == nil {
			if newCount := focusedMonitor.RootCount + count; newCount >= 1 && newCount <= focusedMonitor.Windows.Len() {
				focusedMonitor.RootCount = newCount
				focusedMonitor.Arrange()
			}
		}

	case cmd.Count != "":
		if count, err := strconv.Atoi(cmd.Count); err == nil {
			if count >= 1 && count <= focusedMonitor.Windows.Len() {
				focusedMonitor.RootCount = count
				focusedMonitor.Arrange()
			}
		}

	}

	return selectedWindow
}

// ratio [size]

type RatioCmd struct {
	Ratio string `arg:"positional"`
}

func (cmd RatioCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case cmd.Ratio == "":
		fmt.Fprintln(req, focusedMonitor.Ratio)

	case isCount(cmd.Ratio):
		if size, err := strconv.ParseFloat(cmd.Ratio, 64); err == nil {
			if newSize := focusedMonitor.Ratio + size; newSize >= 0.2 && newSize <= 0.8 {
				focusedMonitor.Ratio = newSize
				focusedMonitor.Arrange()
			}
		}

	case cmd.Ratio != "":
		if size, err := strconv.ParseFloat(req.Args[0], 64); err == nil {
			if size >= 0.2 && size <= 0.8 {
				focusedMonitor.Ratio = size
				focusedMonitor.Arrange()
			}
		}

	}

	return selectedWindow
}

// padding left|right|top|bottom [size]

type PaddingCmd struct {
	Direction string `arg:"positional"`
	Size      string `arg:"positional"`
}

func (cmd PaddingCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case cmd.Direction == "left" && cmd.Size == "":
		fmt.Fprintln(req, focusedMonitor.Padding.L)

	case cmd.Direction == "left" && cmd.Size != "":
		if padding, err := strconv.Atoi(cmd.Size); err == nil {
			focusedMonitor.Padding.L = padding
			focusedMonitor.Arrange()
		}

	case cmd.Direction == "right" && cmd.Size == "":
		fmt.Fprintln(req, focusedMonitor.Padding.R)

	case cmd.Direction == "right" && cmd.Size != "":
		if padding, err := strconv.Atoi(cmd.Size); err == nil {
			focusedMonitor.Padding.R = padding
			focusedMonitor.Arrange()
		}

	case cmd.Direction == "top" && cmd.Size == "":
		fmt.Fprintln(req, focusedMonitor.Padding.T)

	case cmd.Direction == "top" && cmd.Size != "":
		if padding, err := strconv.Atoi(cmd.Size); err == nil {
			focusedMonitor.Padding.T = padding
			focusedMonitor.Arrange()
		}

	case cmd.Direction == "bottom" && cmd.Size == "":
		fmt.Fprintln(req, focusedMonitor.Padding.B)

	case cmd.Direction == "bottom" && cmd.Size != "":
		if padding, err := strconv.Atoi(cmd.Size); err == nil {
			focusedMonitor.Padding.B = padding
			focusedMonitor.Arrange()
		}
	}

	return selectedWindow
}

// fullscreen [ false|true|toggle ]

type FullscreenCmd struct {
	State string `arg:"positional"`
}

func (cmd FullscreenCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case cmd.State == "" && focusedMonitor.Fullscreen:
		fmt.Fprintln(req, "true")

	case cmd.State == "":
		fmt.Fprintln(req, "false")

	case cmd.State == "false":
		if focusedMonitor.Fullscreen {
			focusedMonitor.Fullscreen = false
			focusedMonitor.Arrange()
		}

	case cmd.State == "true":
		if !focusedMonitor.Fullscreen {
			focusedMonitor.Fullscreen = true
			focusedMonitor.Arrange()
		}

	case cmd.State == "toggle":
		focusedMonitor.Fullscreen = !focusedMonitor.Fullscreen
		focusedMonitor.Arrange()
	}

	return selectedWindow
}

// select-layout [-N|+N|name]

type SelectLayoutCmd struct {
	Layout string `arg:"positional"`
}

func (cmd SelectLayoutCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case cmd.Layout == "":
		if layout := focusedMonitor.Layouts.Focused(); layout != nil {
			fmt.Fprintln(req, layout.String())
		}

	case isCount(cmd.Layout):
		if count, err := strconv.Atoi(cmd.Layout); err == nil {
			focusedMonitor.Layouts.Focus(count)
			focusedMonitor.Arrange()
		}

	case cmd.Layout != "":
		previousLayout := focusedMonitor.Layouts.Focused()
		focusedMonitor.Layouts.FocusFunc(func(layout muon.Layout) bool {
			return layout.String() == cmd.Layout
		})

		if focusedMonitor.Layouts.Focused() != previousLayout {
			focusedMonitor.Arrange()
		}
	}

	return selectedWindow
}

// reset-layout

type ResetLayoutCmd struct {
}

func (cmd ResetLayoutCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	focusedMonitor.Reset()
	focusedMonitor.Arrange()

	return selectedWindow
}

// mirror-layout [false|true|toggle]

type MirrorLayoutCmd struct {
	State string `arg:"positional"`
}

func (cmd MirrorLayoutCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case cmd.State == "" && focusedMonitor.Mirrored:
		fmt.Fprintln(req, "true")

	case cmd.State == "":
		fmt.Fprintln(req, "false")

	case cmd.State == "false":
		if focusedMonitor.Mirrored {
			focusedMonitor.Mirrored = false
			focusedMonitor.Arrange()
		}

	case cmd.State == "true":
		if !focusedMonitor.Mirrored {
			focusedMonitor.Mirrored = true
			focusedMonitor.Arrange()
		}

	case cmd.State == "toggle":
		focusedMonitor.Mirrored = !focusedMonitor.Mirrored
		focusedMonitor.Arrange()

	}

	return selectedWindow
}

// focus-monitor [-N|+N]

type FocusMonitorCmd struct {
	Selector string `arg:"positional"`
}

func (cmd FocusMonitorCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case isCount(cmd.Selector):
		if count, err := strconv.Atoi(cmd.Selector); err == nil {
			manager.Monitors.Focus(count)
		}
	}

	return selectedWindow
}

// focus-window [pointer|id|-N+|+N]

type FocusWindowCmd struct {
	Selector string `arg:"positional"`
}

func (cmd FocusWindowCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case cmd.Selector == "pointer":
		if win, mon := manager.FindWindowPointer(); mon != nil && win != nil {
			manager.Monitors.FocusMatch(mon)
			mon.Windows.FocusMatch(win)

			if mon.Fullscreen {
				mon.Arrange()
			}
		}

	case isCount(cmd.Selector) && selectedWindow != nil:
		focusedMonitor.Windows.FocusMatch(selectedWindow)

		if focusedMonitor.Fullscreen {
			focusedMonitor.Arrange()
		}

	case isCount(cmd.Selector):
		if count, err := strconv.Atoi(cmd.Selector); err == nil {
			focusedMonitor.Windows.Focus(count)

			if focusedMonitor.Fullscreen {
				focusedMonitor.Arrange()
			}
		}

	case isHex(cmd.Selector):
		if win, mon := manager.FindWindowString(cmd.Selector); mon != nil && win != nil {
			mon.Windows.FocusMatch(win)

			if mon.Fullscreen {
				mon.Arrange()
			}
		}
	}

	return nil
}

// select-window [id|-N+|+N]

type SelectWindowCmd struct {
	Selector string `arg:"positional"`
}

func (cmd SelectWindowCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case isCount(cmd.Selector):
		if count, err := strconv.Atoi(cmd.Selector); err == nil {
			return focusedMonitor.Windows.Select(count)
		}

	case isHex(cmd.Selector):
		selectedWindow, _ = manager.FindWindowString(cmd.Selector)
		return selectedWindow
	}

	return selectedWindow
}

// move-window [pointer|id|-N|+N|(selected)]

type MoveWindowCmd struct {
	Selector string `arg:"positional"`
}

func (cmd MoveWindowCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case cmd.Selector == "pointer":
		if win, mon := manager.FindWindowPointer(); mon != nil && win != nil {
			mon.Windows.MoveFocusMatch(win)
			mon.Arrange()
		}

	case isCount(cmd.Selector):
		if count, err := strconv.Atoi(cmd.Selector); err == nil {
			focusedMonitor.Windows.MoveFocus(count)
			focusedMonitor.Arrange()
		}

	case isHex(cmd.Selector):
		if win, mon := manager.FindWindowString(cmd.Selector); mon != nil && win != nil {
			mon.Windows.MoveFocusMatch(win)
			mon.Arrange()
		}

	case cmd.Selector == "" && selectedWindow != nil:
		focusedMonitor.Windows.MoveFocusMatch(selectedWindow)
		focusedMonitor.Arrange()
	}

	return nil
}

// root-window [-focus] [pointer|id|-N|+N|(selected)]

type RootWindowCmd struct {
	Focus    bool
	Selector string `arg:"positional"`
}

func (cmd RootWindowCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case cmd.Selector == "":
		focusedMonitor.Windows.NodeMatch(focusedWindow, focusedMonitor.Windows.SwapFront)
		focusedMonitor.Arrange()

	case selectedWindow != nil:
		focusedMonitor.Windows.NodeMatch(selectedWindow, focusedMonitor.Windows.SwapFront)

		if cmd.Focus {
			focusedMonitor.Windows.FocusMatch(selectedWindow)
		}

		focusedMonitor.Arrange()

	case cmd.Selector == "pointer":
		if win, mon := manager.FindWindowPointer(); mon != nil && win != nil {
			mon.Windows.NodeMatch(win, mon.Windows.SwapFront)

			if cmd.Focus {
				manager.Monitors.FocusMatch(mon)
				mon.Windows.FocusMatch(win)
			}

			mon.Arrange()
		}

	case isCount(cmd.Selector):
		if count, err := strconv.Atoi(cmd.Selector); err == nil {
			if win := focusedMonitor.Windows.Select(count); win != nil {
				focusedMonitor.Windows.NodeMatch(win, focusedMonitor.Windows.SwapFront)

				if cmd.Focus {
					focusedMonitor.Windows.FocusMatch(win)
				}

				focusedMonitor.Arrange()
			}
		}

	case isHex(cmd.Selector):
		if win, mon := manager.FindWindowString(cmd.Selector); mon != nil && win != nil {
			mon.Windows.NodeMatch(win, mon.Windows.SwapFront)

			if cmd.Focus {
				manager.Monitors.FocusMatch(mon)
				mon.Windows.FocusMatch(win)
			}

			mon.Arrange()
		}
	}

	return nil
}

// close-window [pointer|(selected)]

type CloseWindowCmd struct {
	Selector string `arg:"positional"`
}

func (cmd CloseWindowCmd) Run(req Request,
	manager *muon.Manager, focusedMonitor *muon.Monitor,
	focusedWindow, selectedWindow *muon.Window,
) *muon.Window {
	switch {
	case cmd.Selector == "pointer":
		if win, mon := manager.FindWindowPointer(); mon != nil && win != nil {
			if _, ok := win.Protocols[xlib.DeleteWindowAtom]; ok {
				xlib.ClientMessage(win.Id, uint32(xlib.DeleteWindowAtom), xproto.TimeCurrentTime)
			} else {
				xproto.KillClient(xlib.Conn, uint32(win.Id))
			}
		}

	case cmd.Selector == "" && selectedWindow != nil:
		if _, ok := selectedWindow.Protocols[xlib.DeleteWindowAtom]; ok {
			xlib.ClientMessage(selectedWindow.Id, uint32(xlib.DeleteWindowAtom), xproto.TimeCurrentTime)
		} else {
			xproto.KillClient(xlib.Conn, uint32(selectedWindow.Id))
		}

	case cmd.Selector == "":
		if win := focusedMonitor.Focused(); win != nil {
			if _, ok := win.Protocols[xlib.DeleteWindowAtom]; ok {
				xlib.ClientMessage(win.Id, uint32(xlib.DeleteWindowAtom), xproto.TimeCurrentTime)
			} else {
				xproto.KillClient(xlib.Conn, uint32(win.Id))
			}
		}
	}

	return nil
}
