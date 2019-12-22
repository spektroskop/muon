package main

import (
	"bufio"
	"context"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/BurntSushi/xgb"
	"go.uber.org/zap"

	"yuki.no/muon/muon"
	"yuki.no/muon/xlib"
)

func main() {
	var (
		ctx, cancel    = context.WithCancel(context.Background())
		signalChannel  = make(chan os.Signal, 1)
		eventChannel   = make(chan xgb.Event)
		errorChannel   = make(chan error)
		commandChannel = make(chan Request)
	)

	defer cancel()
	signal.Notify(signalChannel, os.Interrupt)

	defer os.Remove("/tmp/muon")
	os.Remove("/tmp/muon")
	socket, err := net.Listen("unix", "/tmp/muon")
	if err != nil {
		panic(err)
	}

	manager, err := muon.NewManager()
	if err != nil {
		panic(err)
	}

	if err := manager.Manage(); err != nil {
		panic(err)
	}

	if err := manager.Setup(); err != nil {
		panic(err)
	}

	go func() {
		for {
			conn, err := socket.Accept()
			if err != nil {
				errorChannel <- err
				continue
			}

			defer conn.Close()
			reader := bufio.NewReader(conn)
			command, err := reader.ReadString('\n')
			if err != nil {
				errorChannel <- err
				continue
			}

			if fields := strings.Fields(command); len(fields) != 0 {
				commandChannel <- Request{Conn: conn, Command: fields[0], Args: fields[1:]}
			}
		}
	}()

	go func() {
		for {
			event, err := xlib.Conn.WaitForEvent()
			if err != nil {
				errorChannel <- err
				continue
			}

			eventChannel <- event
		}
	}()

	var (
		clearSelection = time.NewTimer(0)
		selectedWindow *muon.Window
	)

	for {
		var (
			previousMonitor   = manager.Focused()
			previousWindow    = previousMonitor.Focused()
			previousSelection = selectedWindow
		)

		select {
		case <-ctx.Done():
			return

		case <-signalChannel:
			return

		case err := <-errorChannel:
			zap.S().Error(err)

		case <-clearSelection.C:
			if selectedWindow != nil {
				xlib.SetBorderColor(selectedWindow.Id, manager.NormalBorder.Value)
				selectedWindow = nil
			}

		case request := <-commandChannel:
			selectedWindow = runCommand(manager, request, previousMonitor, previousWindow, previousSelection)
			resetSelection(manager, clearSelection, selectedWindow, previousSelection)
			resetFocus(manager, previousWindow)

		case event := <-eventChannel:
			handleEvent(manager, event, &previousWindow)
			resetFocus(manager, previousWindow)
		}
	}
}

func resetSelection(manager *muon.Manager, clearSelection *time.Timer, selectedWindow, previousSelection *muon.Window) {
	if selectedWindow == previousSelection {
		clearSelection.Reset(time.Second)
		return
	}

	if selectedWindow != nil {
		zap.S().Infow("select", "window", selectedWindow)

		xlib.SetBorderColor(selectedWindow.Id, manager.SelectedBorder.Value)
		clearSelection.Reset(time.Second)
	}

	if previousSelection != nil {
		xlib.SetBorderColor(previousSelection.Id, manager.NormalBorder.Value)
	}
}

func resetFocus(manager *muon.Manager, previousWindow *muon.Window) {
	if mon := manager.Focused(); mon != nil {
		if win := mon.Focused(); win != nil {
			if previousWindow != win {
				xlib.SetBorderColor(win.Id, manager.FocusedBorder.Value)
				xlib.Raise(win.Id)
				xlib.SetFocus(win.Id)

				zap.S().Infow("focus", "window", win)

				if previousWindow != nil {
					xlib.SetBorderColor(previousWindow.Id, manager.NormalBorder.Value)
				}
			}
		}
	}
}
