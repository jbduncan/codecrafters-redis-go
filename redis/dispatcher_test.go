package redis_test

import (
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/iox"
	"github.com/codecrafters-io/redis-starter-go/redis"
	"github.com/codecrafters-io/redis-starter-go/redis/redistest"
)

func TestDispatcher_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		request  string
		response string
	}{
		{
			name:     "ping",
			request:  redistest.RequestPingLowercase,
			response: redistest.ResponsePong,
		},
		{
			name:     "PING",
			request:  redistest.RequestPingUppercase,
			response: redistest.ResponsePong,
		},
		{
			name:     "Pipeline",
			request:  strings.Repeat(redistest.RequestPingLowercase, 3),
			response: strings.Repeat(redistest.ResponsePong, 3),
		},
		{
			name:     "echo foo",
			request:  "*2\r\n$4\r\necho\r\n$3\r\nfoo\r\n",
			response: "$3\r\nfoo\r\n",
		},
		{
			name:     "Invalid syntax",
			request:  "*1\r\n^",
			response: "!41\r\nERR Protocol error: expected '$', got '^'\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conns := runDispatcher(t, 1)
			c := <-conns

			if _, err := io.WriteString(c, tt.request); err != nil {
				t.Fatalf("request %q not sent: %v", tt.request, err)
			}

			got, err := iox.ReadExactly(c, len(tt.response))
			if err != nil {
				t.Fatalf("response not read: %v", err)
			}
			if want := tt.response; got != want {
				t.Errorf(
					`Dispatcher.Run(): PING request: got response %q, want %q`,
					got,
					want,
				)
			}
		})
	}

	for _, tt := range []struct {
		name            string
		concurrentPings int
	}{
		{
			name:            "two concurrent PINGs",
			concurrentPings: 2,
		},
		{
			name:            "1,000 concurrent PINGs",
			concurrentPings: 1_000,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conns := runDispatcher(t, tt.concurrentPings)

			redistest.TestConcurrentPings(
				t,
				tt.concurrentPings,
				func() net.Conn {
					return <-conns
				},
			)
		})
	}
}

func runDispatcher(t *testing.T, conns int) <-chan net.Conn {
	serverConns := make(chan net.Conn, conns)
	clientConns := make(chan net.Conn, conns)
	c := make([]net.Conn, 0, 2*conns)
	for range conns {
		clientConn, serverConn := netPipe()
		serverConns <- serverConn
		clientConns <- clientConn
		c = append(c, clientConn, serverConn)
	}

	dispatcher := redis.NewDispatcher(
		nopCloseTCPConnAccepter{
			delegate: func() (redis.TCPConn, error) {
				// This will eventually block to stop Dispatcher's inner
				// loop from looping forever.
				return <-serverConns, nil
			},
		},
		redis.NewDefaultRouter(
			redis.EchoHandler{},
			redis.PingHandler{},
		),
		slog.New(slog.DiscardHandler),
	)
	go dispatcher.Run()
	t.Cleanup(func() {
		for _, c := range c {
			_ = c.Close()
		}
		dispatcher.Stop()
	})

	return clientConns
}

func netPipe() (net.Conn, net.Conn) {
	a, b := net.Pipe()
	deadline := time.Now().Add(10 * time.Second)
	_ = a.SetDeadline(deadline)
	_ = b.SetDeadline(deadline)
	return a, b
}

type nopCloseTCPConnAccepter struct {
	delegate func() (redis.TCPConn, error)
}

func (a nopCloseTCPConnAccepter) Accept() (redis.TCPConn, error) {
	return a.delegate()
}

func (a nopCloseTCPConnAccepter) Close() {
	// nop
}
