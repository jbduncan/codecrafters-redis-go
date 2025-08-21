package redis

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
)

const (
	defaultPort = 6379
)

type Server struct {
	tcpListener net.Listener
	router      Router
	events      chan event
	quit        chan struct{}
	wg          *sync.WaitGroup
	stopOnce    *sync.Once
	logger      *slog.Logger
}

func StartServer(ctx context.Context, stderr io.Writer) (*Server, error) {
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
	router := NewDefaultRouter(
		EchoHandler{},
		PingHandler{},
	)

	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", defaultPort))
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %w", PortNotBoundError{port: defaultPort}, err,
		)
	}
	logger.Info(fmt.Sprintf("server is listening on port %d", defaultPort))

	s := &Server{
		tcpListener: l,
		router:      router,
		events:      make(chan event, 512),
		quit:        make(chan struct{}),
		wg:          new(sync.WaitGroup),
		stopOnce:    new(sync.Once),
		logger:      logger,
	}
	s.run()
	return s, nil
}

func (s *Server) run() {
	go func() {
		for {
			tcpConn, err := s.tcpListener.Accept()
			if err != nil {
				select {
				case <-s.quit:
					return
				default:
					s.logError(err)
				}
				continue
			}

			s.wg.Add(1)
			go s.handleConn(tcpConn)
		}
	}()

	go func() {
		for e := range s.events {
			if err := NewRESP2Writer(e.conn).Write(e.value); err != nil {
				s.logError(err)
			}
		}
	}()
}

func (s *Server) handleConn(tcpConn net.Conn) {
	defer s.wg.Done()

	for value, err := range NewRESP2Scanner(tcpConn).ScanAll() {
		var errorValue ErrorValue
		if errors.As(err, &errorValue) {
			s.events <- event{
				value: errorValue,
				conn:  tcpConn,
			}
			return
		}
		if err != nil {
			s.logError(err)
		}

		result := s.router.Route(value)
		s.events <- event{
			value: result,
			conn:  tcpConn,
		}
	}
}

func (s *Server) Stop() {
	// Stops more TCP connections from being accepted and allows Run() to drain
	// all remaining events in d.events before terminating.
	s.stopOnce.Do(func() {
		close(s.quit)
		if err := s.tcpListener.Close(); err != nil {
			s.logError(err)
		}
		s.wg.Wait()
	})
}

func (s *Server) logError(err error) {
	s.logger.Error(err.Error())
}

type event struct {
	value Value
	conn  net.Conn
}

type PortNotBoundError struct {
	port int
}

func (p PortNotBoundError) Error() string {
	return fmt.Sprintf("port %d not bound", defaultPort)
}

func (p PortNotBoundError) Port() int {
	return p.port
}
