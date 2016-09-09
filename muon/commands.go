package main

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spektroskop/muon/manager"
)

func process(cmd []string) error {
	logrus.Debugf("Command %#v", cmd)

	switch cmd[0] {
	case "select-monitor":
		var selector = 1

		if len(cmd) > 1 {
			var err error
			if selector, err = strconv.Atoi(cmd[1]); err != nil {
				return err
			} else if selector == 0 {
				return nil
			}
		}

		switch {
		case selector < 0:
			for ; selector < 0; selector++ {
				manager.Focus = manager.Focus.Prev(manager.Monitors)
			}
		case selector > 0:
			for ; selector > 0; selector-- {
				manager.Focus = manager.Focus.Next(manager.Monitors)
			}
		}

		manager.WithMonitor(func(monitor *manager.Monitor) {
			monitor.SetFocus(monitor.Focus)
		})
	case "focus-window":
		monitor, window := manager.PointedWindow()
		if window != nil {
			manager.MonitorNode(monitor, func(monitor *manager.Monitor) {
				monitor.SetFocus(window)
			})
		}
	case "select-window":
		var selector = 1

		if len(cmd) > 1 {
			var err error
			if selector, err = strconv.Atoi(cmd[1]); err != nil {
				return err
			} else if selector == 0 {
				return nil
			}
		}

		manager.WithMonitor(func(monitor *manager.Monitor) {
			focus := monitor.Focus

			switch {
			case selector < 0:
				for ; selector < 0; selector++ {
					focus = focus.Prev(monitor.Windows)
				}
			case selector > 0:
				for ; selector > 0; selector-- {
					focus = focus.Next(monitor.Windows)
				}
			}

			monitor.SetFocus(focus)
		})
	case "shift-window":
		var selector = 1

		if len(cmd) > 1 {
			var err error
			if selector, err = strconv.Atoi(cmd[1]); err != nil {
				return err
			} else if selector == 0 {
				return nil
			}
		}

		manager.WithMonitor(func(monitor *manager.Monitor) {
			switch {
			case selector < 0:
				for ; selector < 0; selector++ {
					// monitor.Focus.Shift(monitor.Windows)
				}
			case selector > 0:
				for ; selector > 0; selector-- {
					// monitor.Focus.Unshift(monitor.Windows)
				}
			}

			monitor.Arrange()
		})
	case "make-root":
		manager.WithMonitor(func(monitor *manager.Monitor) {
			if monitor.Focus != nil {
				monitor.Windows.First().Swap(monitor.Focus)
				monitor.SetFocus(monitor.Windows.First())
				monitor.Arrange()
			}
		})
	case "next-layout":
		manager.WithMonitor(func(monitor *manager.Monitor) {
			monitor.Layout = monitor.Layout.Next(nil)
			monitor.Arrange()
		})
	case "reset-layout":
		manager.WithMonitor(func(monitor *manager.Monitor) {
			monitor.Reset()
			monitor.Arrange()
		})
	case "mirror-layout":
		manager.WithMonitor(func(monitor *manager.Monitor) {
			monitor.Mirrored = !monitor.Mirrored
			monitor.Arrange()
		})
	case "set-ratio":
		if len(cmd) != 2 {
			return errors.New("Argument error")
		}

		ratio, err := strconv.ParseFloat(cmd[1], 64)
		if err != nil {
			return err
		}

		manager.WithMonitor(func(monitor *manager.Monitor) {
			if strings.HasPrefix(cmd[1], "+") || strings.HasPrefix(cmd[1], "-") {
				ratio += monitor.Ratio
			}

			if ratio > 0.9 {
				ratio = 0.9
			} else if ratio < 0.1 {
				ratio = 0.1
			}

			monitor.Ratio = ratio
			monitor.Arrange()
		})
	case "set-roots":
		if len(cmd) != 2 {
			return errors.New("Argument error")
		}

		value, err := strconv.Atoi(cmd[1])
		if err != nil {
			return err
		}

		manager.WithMonitor(func(monitor *manager.Monitor) {
			if strings.HasPrefix(cmd[1], "+") || strings.HasPrefix(cmd[1], "-") {
				value += monitor.Roots
			}

			if count := monitor.Windows.Len(); value > count {
				value = count
			} else if value < 1 {
				value = 1
			}

			if value != monitor.Roots {
				monitor.Roots = value
				monitor.Arrange()
			}
		})
	}

	return nil
}
