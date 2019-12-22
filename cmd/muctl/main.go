package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

var (
	network = flag.String("network ", "unix", "")
	address = flag.String("address ", "/tmp/muon", "")
	timeout = flag.Duration("timeout", time.Second*5, "")
	source  = flag.String("source", "", "")
)

func main() {
	flag.Parse()

	conn, err := net.DialTimeout(*network, *address, *timeout)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	conn.SetDeadline(time.Now().Add(*timeout))
	defer conn.Close()

	args := strings.Join(os.Args[1:], " ")
	fmt.Fprintln(conn, args)

	reader := bufio.NewReader(conn)
	response, _ := reader.ReadString('\n')
	if actual := strings.TrimSpace(response); actual != "" {
		fmt.Println(actual)
	}
}
