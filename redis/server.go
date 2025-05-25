package redis

import (
	"fmt"
	"log/slog"
	"net"
	"os"
)

const (
	defaultPort = 6379
)

func RunServer() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", defaultPort))
	if err != nil {
		logError(fmt.Errorf("port %d not bound: %w", defaultPort, err))
		os.Exit(1)
	}
	defer closeAndLogError(l)

	slog.Info(fmt.Sprintf("server is listening on port %d", defaultPort))

	h := NewHandler(func() (TCPConn, error) { return l.Accept() })
	h.Handle()
}
