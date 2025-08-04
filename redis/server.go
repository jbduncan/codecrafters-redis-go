package redis

import (
	"fmt"
	"io"
	"log/slog"
	"net"
)

const (
	defaultPort = 6379
)

type Server struct {
	d *Dispatcher
}

func (s *Server) Run(stderr io.Writer) error {
	// TODO: Figure out a way to make the TCP server stop gracefully on a
	//       SIGINT or SIGTERM, including terminating slow TCP connections.
	//   - https://victoriametrics.com/blog/go-graceful-shutdown/
	//   - https://www.rudderstack.com/blog/implementing-graceful-shutdown-in-go/
	//   - https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
	//   - Search other resources

	levelVar := &slog.LevelVar{}
	levelVar.Set(slog.LevelInfo)
	logger := slog.New(
		slog.NewTextHandler(
			stderr,
			&slog.HandlerOptions{
				Level: levelVar,
			},
		),
	)

	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", defaultPort))
	if err != nil {
		return fmt.Errorf("port %d not bound: %w", defaultPort, err)
	}

	slog.Info(fmt.Sprintf("server is listening on port %d", defaultPort))

	s.d = NewDispatcher(
		tcpConnAccepter{delegate: l},
		NewDefaultRouter(
			EchoHandler{},
			PingHandler{},
		),
		logger,
	)
	s.d.Run()
	return nil
}

func (s *Server) Stop() {
	s.d.Stop()
}

type tcpConnAccepter struct {
	delegate net.Listener
	logger   slog.Logger
}

func (a tcpConnAccepter) Accept() (TCPConn, error) {
	return a.delegate.Accept()
}

func (a tcpConnAccepter) Close() {
	if err := a.delegate.Close(); err != nil {
		a.logger.Error(fmt.Sprintf("%v", err))
	}
}
