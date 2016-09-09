package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/Sirupsen/logrus"
	"github.com/spektroskop/muon/dispatcher"
	"github.com/spektroskop/muon/manager"
)

var (
	debug   = flag.Bool("debug", false, "")
	kind    = flag.String("kind", "unix", "")
	address = flag.String("address", "/tmp/muon", "")
)

func main() {
	flag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if err := manager.New(); err != nil {
		logrus.Fatal(err)
	}

	cmds, err := dispatcher.Listen(*kind, *address)
	if err != nil {
		logrus.Fatal(err)
	}

	if *kind == "unix" {
		defer os.Remove(*address)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

Loop:
	for {
		select {
		case <-interrupt:
			logrus.Info("Interrupt")
			break Loop
		case cmd := <-cmds:
			if err := process(cmd); err != nil {
				logrus.Errorf("Command `%s': %s", cmd[0], err)
			}
		}
	}
}
