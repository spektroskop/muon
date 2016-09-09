package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

var (
	kind    = flag.String("kind", "unix", "")
	address = flag.String("address", "/tmp/muon", "")
	timeout = flag.Int("timeout", 1, "")
)

func main() {
	flag.Parse()

	conn, err := net.DialTimeout(*kind, *address, time.Duration(*timeout))
	if err != nil {
		logrus.Fatal(err)
	}

	conn.SetDeadline(time.Now().Add(time.Duration(*timeout)))
	defer conn.Close()

	args := strings.Join(os.Args[1:], " ")

	fmt.Fprintln(conn, args)
}
