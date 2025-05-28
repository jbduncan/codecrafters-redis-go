package redis_test

import (
	"bufio"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

type mockTCPConn struct {
	readCalled        bool
	readCalledMu      *sync.Mutex
	readErrorToReturn error
}

func newFailingOnReadMockTCPConn(err error) *mockTCPConn {
	return &mockTCPConn{
		readCalled:        false,
		readCalledMu:      new(sync.Mutex),
		readErrorToReturn: err,
	}
}

func (c *mockTCPConn) Read(_ []byte) (n int, err error) {
	c.readCalledMu.Lock()
	defer c.readCalledMu.Unlock()
	c.readCalled = true

	if c.readErrorToReturn != nil {
		return 0, c.readErrorToReturn
	}

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
	tests := []struct {
		name            string
		request         string
		numSubResponses int
	}{
		{
			name:            "PING: uppercase simple string request",
			request:         "PING\r\n",
			numSubResponses: 1,
		},
		{
			name:            "PING: lowercase simple string request",
			request:         "ping\r\n",
			numSubResponses: 1,
		},
		{
			name:            "PING: array request",
			request:         "*1\r\n$4\r\nPING\r\n",
			numSubResponses: 1,
		},
		{
			name:            "three PINGs: three pipelined simple requests",
			request:         "PING\r\nPING\r\nPING\r\n",
			numSubResponses: 3,
		},
		{
			name:            "three PINGs: three pipelined requests in array",
			request:         "*3\r\n$4\r\nPING\r\n$4\r\nPING\r\n$4\r\nPING\r\n",
			numSubResponses: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type conn struct {
				tcpConn redis.TCPConn
				err     error
			}
			serverConns := make(chan *conn, 2)
			clientConn, serverConn := netPipe()
			serverConns <- &conn{
				tcpConn: serverConn,
			}
			serverConns <- &conn{
				err: errors.New("no new connections left"),
			}

			dispatcher := redis.NewDispatcher(
				func() (redis.TCPConn, error) {
					// This will eventually block to stop Dispatcher's inner loop
					// from looping forever.
					c := <-serverConns
					return c.tcpConn, c.err
				},
			)

			go dispatcher.Run()
			if _, err := io.WriteString(clientConn, tt.request); err != nil {
				t.Fatalf("request not sent through clientConn: %v", err)
			}

			for range tt.numSubResponses {
				resp, err := bufio.NewReader(clientConn).ReadString('\n')
				if err != nil {
					t.Fatalf("resp not read: %v", err)
				}

				if got, want := resp, "+PONG\r\n"; got != want {
					t.Errorf(
						`Dispatcher.Run(): PING request: got response %q, want %q`,
						got,
						want,
					)
				}
			}
		})
	}

	t.Run("two PINGS: two concurrent requests", func(t *testing.T) {
		type conn struct {
			tcpConn redis.TCPConn
			err     error
		}
		serverConns := make(chan *conn, 3)
		clientConn1, serverConn1 := netPipe()
		clientConn2, serverConn2 := netPipe()
		serverConns <- &conn{
			tcpConn: serverConn1,
		}
		serverConns <- &conn{
			tcpConn: serverConn2,
		}
		serverConns <- &conn{
			err: errors.New("no new connections left"),
		}

		dispatcher := redis.NewDispatcher(
			func() (redis.TCPConn, error) {
				// This will eventually block to stop Dispatcher's inner loop
				// from looping forever.
				c := <-serverConns
				return c.tcpConn, c.err
			},
		)

		go dispatcher.Run()
		if _, err := io.WriteString(clientConn1, "PING\r\n"); err != nil {
			t.Fatalf("request not sent through clientConn: %v", err)
		}
		if _, err := io.WriteString(clientConn2, "PING\r\n"); err != nil {
			t.Fatalf("request not sent through clientConn: %v", err)
		}

		type resp struct {
			s   string
			err error
		}
		resps := make(chan resp)
		go func() {
			r, err := bufio.NewReader(clientConn1).ReadString('\n')
			resps <- resp{s: r, err: err}
		}()
		go func() {
			r, err := bufio.NewReader(clientConn2).ReadString('\n')
			resps <- resp{s: r, err: err}
		}()
		for range 2 {
			r := <-resps
			if r.err != nil {
				t.Fatalf("resp not read: %v", r.err)
			}
			if got, want := r.s, "+PONG\r\n"; got != want {
				t.Errorf(
					`Dispatcher.Run(): PING request: got response %q, want %q`,
					got,
					want,
				)
			}
		}
	})

	t.Run(
		"edge case: when tcpConnAccepter returns error, then no attempt to "+
			"read the conn is made",
		func(t *testing.T) {
			conn := newFailingOnReadMockTCPConn(errors.New("TCP conn blew up"))
			dispatcher := redis.NewDispatcher(
				func() (redis.TCPConn, error) {
					return conn, errors.New("no new connections left")
				},
			)

			go dispatcher.Run()

			for range 30 {
				time.Sleep(100 * time.Millisecond)
				if conn.ReadCalled() {
					t.Errorf(
						"Dispatcher.Run(): expected not to call conn.Read() " +
							"but it did",
					)
				}
			}
		})
}
