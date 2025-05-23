package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/errorsx"
	"github.com/codecrafters-io/redis-starter-go/redis"
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
		errorsx.Log(fmt.Errorf("port %d not bound: %w", defaultPort, err))
		os.Exit(1)
	}
	defer errorsx.Close(l)

	slog.Info(fmt.Sprintf("server is listening on port %d", defaultPort))

	h := redis.NewHandler(func() (redis.TCPConn, error) { return l.Accept() })
	h.Handle()
}
