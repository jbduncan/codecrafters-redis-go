package redis_test

import (
	"bufio"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/await"
	"github.com/codecrafters-io/redis-starter-go/redis"
)

type spyTCPConn struct {
	delegate           redis.TCPConn
	closeCalled        bool
	closeCalledLock    *sync.Mutex
	writeErrorToReturn error
}

func newSpyTCPConn(delegate net.Conn) *spyTCPConn {
	return &spyTCPConn{
		delegate:           delegate,
		closeCalled:        false,
		closeCalledLock:    new(sync.Mutex),
		writeErrorToReturn: nil,
	}
}

func newFailingOnWriteSpyTCPConn(
	delegate redis.TCPConn,
	err error,
) *spyTCPConn {
	return &spyTCPConn{
		delegate:           delegate,
		closeCalled:        false,
		closeCalledLock:    new(sync.Mutex),
		writeErrorToReturn: err,
	}
}

func (c *spyTCPConn) Read(b []byte) (n int, err error) {
	if c.delegate != nil {
		return c.delegate.Read(b)
	}

	return 0, nil
}

func (c *spyTCPConn) Write(b []byte) (n int, err error) {
	if c.writeErrorToReturn != nil {
		return 0, c.writeErrorToReturn
	}

	if c.delegate != nil {
		return c.delegate.Write(b)
	}

	return 0, nil
}

func (c *spyTCPConn) Close() error {
	var err error
	if c.delegate != nil {
		err = c.delegate.Close()
	}

	c.closeCalledLock.Lock()
	defer c.closeCalledLock.Unlock()
	c.closeCalled = true

	return err
}

func (c *spyTCPConn) CloseCalled() bool {
	c.closeCalledLock.Lock()
	defer c.closeCalledLock.Unlock()
	return c.closeCalled
}

type mockTCPConn struct {
	readCalled        bool
	readErrorToReturn error
}

func newFailingOnReadMockTCPConn(err error) *mockTCPConn {
	return &mockTCPConn{
		readCalled:        false,
		readErrorToReturn: err,
	}
}

func (c *mockTCPConn) Read(_ []byte) (n int, err error) {
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
	return c.readCalled
}

func netPipe() (net.Conn, net.Conn) {
	a, b := net.Pipe()
	deadline := time.Now().Add(10 * time.Second)
	_ = a.SetDeadline(deadline)
	_ = b.SetDeadline(deadline)
	return a, b
}

func TestHandlerHandle(t *testing.T) {
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
			serverConns := make(chan *spyTCPConn, 1)
			clientConn, serverConn := netPipe()
			spyServerConn := newSpyTCPConn(serverConn)
			serverConns <- spyServerConn

			handler := redis.NewHandler(
				func() (redis.TCPConn, error) {
					// This will eventually block to stop Handler's inner loop
					// from looping forever.
					return <-serverConns, nil
				},
			)

			go handler.Handle()
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
						`Handler.Handler(): PING request: got response %q, want %q`,
						got,
						want,
					)
				}
			}

			_ = clientConn.Close()
			if !await.Until(
				func() bool { return spyServerConn.CloseCalled() },
				10*time.Second,
				10*time.Millisecond,
			) {
				t.Error(
					"Handler.Handle(): expected server connection to be " +
						"closed but was not")
			}
		})
	}

	t.Run("two PINGS: two concurrent requests", func(t *testing.T) {
		serverConns := make(chan *spyTCPConn, 2)
		clientConn1, serverConn1 := netPipe()
		serverConns <- newSpyTCPConn(serverConn1)
		clientConn2, serverConn2 := netPipe()
		serverConns <- newSpyTCPConn(serverConn2)
		handler := redis.NewHandler(
			func() (redis.TCPConn, error) {
				// This will eventually block to stop Handler's inner loop from
				// looping forever.
				return <-serverConns, nil
			},
		)

		go handler.Handle()
		if _, err := io.WriteString(clientConn1, "PING\r\n"); err != nil {
			t.Fatalf("request not sent through clientConn: %v", err)
		}
		if _, err := io.WriteString(clientConn2, "PING\r\n"); err != nil {
			t.Fatalf("request not sent through clientConn: %v", err)
		}

		for _, c := range []redis.TCPConn{clientConn1, clientConn2} {
			resp, err := bufio.NewReader(c).ReadString('\n')
			if err != nil {
				t.Fatalf("resp not read: %v", err)
			}
			if got, want := resp, "+PONG\r\n"; got != want {
				t.Errorf(
					`Handler.Handler(): PING request: got response %q, want %q`,
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
			handler := redis.NewHandler(
				func() (redis.TCPConn, error) {
					return conn, errors.New("TCP conn blew up")
				},
			)

			handler.Handle()

			if conn.ReadCalled() {
				t.Errorf(
					"Handler.Handler(): expected not to call conn.Read() " +
						"but it did",
				)
			}
		})

	t.Run(
		"edge case: when TCP conn returns error on write, then conn is closed",
		func(t *testing.T) {
			serverConns := make(chan *spyTCPConn, 1)
			clientConn, serverConn := netPipe()
			spyServerConn := newFailingOnWriteSpyTCPConn(
				serverConn, errors.New("TCP conn blew up"),
			)
			serverConns <- spyServerConn

			handler := redis.NewHandler(
				func() (redis.TCPConn, error) {
					// This will eventually block to stop Handler's inner loop
					// from looping forever.
					return <-serverConns, nil
				},
			)
			go handler.Handle()
			if _, err := io.WriteString(clientConn, "PING\r\n"); err != nil {
				t.Fatalf("request not sent through clientConn: %v", err)
			}

			if !await.Until(
				func() bool {
					return spyServerConn.CloseCalled()
				},
				10*time.Second,
				10*time.Millisecond,
			) {
				t.Error(
					"Handler.Handle(): expected server connection to be " +
						"closed but was not")
			}
		})
}
