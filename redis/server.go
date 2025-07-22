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

type Server struct {
	d *Dispatcher
}

func (s *Server) Run() {
	// TODO: Figure out a way to make the TCP server stop gracefully on a
	//       SIGINT or SIGTERM, including terminating slow TCP connections.
	//   - https://victoriametrics.com/blog/go-graceful-shutdown/
	//   - https://www.rudderstack.com/blog/implementing-graceful-shutdown-in-go/
	//   - https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
	//   - Search other resources

	slog.SetLogLoggerLevel(slog.LevelDebug)

	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", defaultPort))
	if err != nil {
		logError(fmt.Errorf("port %d not bound: %w", defaultPort, err))
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("server is listening on port %d", defaultPort))

	s.d = NewDispatcher(tcpConnAccepter{delegate: l})
	s.d.Run()
}

func (s *Server) Stop() {
	s.d.Stop()
}

type tcpConnAccepter struct {
	delegate net.Listener
}

func (a tcpConnAccepter) Accept() (TCPConn, error) {
	return a.delegate.Accept()
}

func (a tcpConnAccepter) Close() {
	closeAndLogError(a.delegate)
}
