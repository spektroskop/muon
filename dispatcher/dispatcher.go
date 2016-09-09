package dispatcher

import (
	"bytes"
	"net"
	"strings"

	"github.com/Sirupsen/logrus"
)

func Listen(kind, address string) (chan []string, error) {
	socket, err := net.Listen(kind, address)
	if err != nil {
		return nil, err
	}

	cmds := make(chan []string)
	go server(socket, cmds)
	return cmds, nil
}

func server(socket net.Listener, cmds chan []string) {
	for {
		conn, err := socket.Accept()
		if err != nil {
			logrus.Error(err)
			continue
		}

		defer conn.Close()
		var buf bytes.Buffer
		buf.ReadFrom(conn)

		command := strings.TrimSpace(buf.String())
		args := strings.Split(command, " ")
		cmds <- args
	}
}
