package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
)

const (
	defaultPort = 6379
)

func main() {
	run()
}

func run() {
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", defaultPort))
	if err != nil {
		printErr(fmt.Errorf("port %d not bound: %w", defaultPort, err))
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("server is listening on port %d", defaultPort))

	_, err = l.Accept()
	if err != nil {
		printErr(err)
		os.Exit(1)
	}
	defer errorHandlingClose(l)
}
