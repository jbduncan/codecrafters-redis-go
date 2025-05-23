package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
)

const (
	defaultPort = 6379
)

func main() {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", defaultPort))
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

func errorHandlingClose(closer io.Closer) {
	if err := closer.Close(); err != nil {
		printErr(err)
	}
}

func printErr(err error) {
	slog.Error(fmt.Sprintf("%v", err))
}
