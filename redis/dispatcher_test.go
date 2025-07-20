package redis_test

import (
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/iox"
	"github.com/codecrafters-io/redis-starter-go/redis"
	"github.com/codecrafters-io/redis-starter-go/redis/redistest"
	"github.com/codecrafters-io/redis-starter-go/repeat"
)

type conn struct {
	tcpConn redis.TCPConn
	err     error
}

type mockTCPConn struct {
	readCalled   bool
	readCalledMu *sync.Mutex
}

func newMockTCPConn() *mockTCPConn {
	return &mockTCPConn{
		readCalled:   false,
		readCalledMu: new(sync.Mutex),
	}
}

func (c *mockTCPConn) Read(_ []byte) (n int, err error) {
	c.readCalledMu.Lock()
	defer c.readCalledMu.Unlock()
	c.readCalled = true

	return 0, nil
}

func (c *mockTCPConn) Write(_ []byte) (n int, err error) {
	return 0, nil
}

func (c *mockTCPConn) Close() error {
	return nil
}

func (c *mockTCPConn) ReadCalled() bool {
	c.readCalledMu.Lock()
	defer c.readCalledMu.Unlock()
	return c.readCalled
}

func netPipe() (net.Conn, net.Conn) {
	a, b := net.Pipe()
	deadline := time.Now().Add(10 * time.Second)
	_ = a.SetDeadline(deadline)
	_ = b.SetDeadline(deadline)
	return a, b
}

func TestDispatcher_Run(t *testing.T) {
	t.Parallel()

	// TODO: wire everything together in Dispatcher

	tests := []struct {
		name     string
		request  string
		response string
	}{
		{
			name:     "ping",
			request:  redistest.PingLowercase,
			response: redistest.Pong,
		},
		{
			name:     "PING",
			request:  redistest.PingUppercase,
			response: redistest.Pong,
		},
		{
			name:     "Pipeline",
			request:  strings.Repeat(redistest.PingLowercase, 3),
			response: strings.Repeat(redistest.Pong, 3),
		},
		{
			name:     "echo foo",
			request:  "*2\r\n$4\r\necho\r\n$3\r\nfoo\r\n",
			response: "$3\r\nfoo\r\n",
		},
		{
			name:     "ECHO foo",
			request:  "*2\r\n$4\r\nECHO\r\n$3\r\nfoo\r\n",
			response: "$3\r\nfoo\r\n",
		},
		{
			name:     "ECHO quux",
			request:  "*2\r\n$4\r\nECHO\r\n$4\r\nquux\r\n",
			response: "$4\r\nquux\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			serverConns := make(chan *conn, 2)
			clientConn, serverConn := netPipe()
			serverConns <- &conn{
				tcpConn: serverConn,
			}
			serverConns <- &conn{
				err: errors.New("no new connections left"),
			}

			runDispatcher(
				t,
				func() (redis.TCPConn, error) {
					// This will eventually block to stop Dispatcher's inner
					// loop from looping forever.
					c := <-serverConns
					return c.tcpConn, c.err
				})
			if _, err := io.WriteString(clientConn, tt.request); err != nil {
				t.Fatalf("request %q not sent: %v", tt.request, err)
			}

			got, err := iox.ReadExactly(clientConn, len(tt.response))
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

			serverConns := make(chan *conn, tt.concurrentPings)
			clientConns := make(chan net.Conn, tt.concurrentPings)
			for range tt.concurrentPings {
				clientConn, serverConn := netPipe()
				serverConns <- &conn{
					tcpConn: serverConn,
				}
				clientConns <- clientConn
			}

			runDispatcher(
				t,
				func() (redis.TCPConn, error) {
					// This will eventually block to stop Dispatcher's inner
					// loop from looping forever.
					c := <-serverConns
					return c.tcpConn, c.err
				})

			redistest.TestConcurrentPings(
				t,
				tt.concurrentPings,
				func() net.Conn {
					return <-clientConns
				})
		})
	}

	t.Run(
		"edge case: when tcpConnAccepter returns error, then conn is not read",
		func(t *testing.T) {
			t.Parallel()

			mockConn := newMockTCPConn()
			serverConns := make(chan conn, 1)
			serverConns <- conn{
				tcpConn: mockConn,
				err:     errors.New("no new connections left"),
			}

			runDispatcher(
				t,
				func() (redis.TCPConn, error) {
					// This will eventually block to stop Dispatcher's inner
					// loop from looping forever.
					c := <-serverConns
					return c.tcpConn, c.err
				})

			if !repeat.Whilst(
				func() bool {
					return !mockConn.ReadCalled()
				},
				3*time.Second,
				100*time.Millisecond,
			) {
				t.Fatalf(
					"Dispatcher.Run(): expected not to call conn.Read() " +
						"but it did",
				)
			}
		})
}

func runDispatcher(
	t *testing.T,
	tcpConnAccepter func() (redis.TCPConn, error),
) {
	dispatcher := redis.NewDispatcher(tcpConnAccepter, func() {})
	// This will panic if the nil connection is read
	go dispatcher.Run()
	t.Cleanup(dispatcher.Stop)
}
