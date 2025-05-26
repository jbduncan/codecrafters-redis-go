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

type fakeTCPConn struct {
	delegate           redis.TCPConn
	readCalled         bool
	closeCalled        bool
	closeCalledLock    *sync.Mutex
	readErrorToReturn  error
	writeErrorToReturn error
}

func newFakeTCPConn(delegate net.Conn) *fakeTCPConn {
	return &fakeTCPConn{
		delegate:           delegate,
		readCalled:         false,
		closeCalled:        false,
		closeCalledLock:    new(sync.Mutex),
		readErrorToReturn:  nil,
		writeErrorToReturn: nil,
	}
}

func newFailingOnReadTCPConn(err error) *fakeTCPConn {
	return &fakeTCPConn{
		readCalled:         false,
		closeCalled:        false,
		closeCalledLock:    new(sync.Mutex),
		readErrorToReturn:  err,
		writeErrorToReturn: nil,
	}
}

func newFailingOnWriteTCPConn(delegate redis.TCPConn, err error) *fakeTCPConn {
	return &fakeTCPConn{
		delegate:           delegate,
		readCalled:         false,
		closeCalled:        false,
		closeCalledLock:    new(sync.Mutex),
		readErrorToReturn:  nil,
		writeErrorToReturn: err,
	}
}

func (c *fakeTCPConn) Read(b []byte) (n int, err error) {
	if c.readErrorToReturn != nil {
		return 0, c.readErrorToReturn
	}

	if c.delegate != nil {
		return c.delegate.Read(b)
	}

	return 0, nil
}

func (c *fakeTCPConn) ReadCalled() bool {
	return c.readCalled
}

func (c *fakeTCPConn) Write(b []byte) (n int, err error) {
	if c.writeErrorToReturn != nil {
		return 0, c.writeErrorToReturn
	}

	if c.delegate != nil {
		return c.delegate.Write(b)
	}

	return 0, nil
}

func (c *fakeTCPConn) Close() error {
	var err error
	if c.delegate != nil {
		err = c.delegate.Close()
	}

	c.closeCalledLock.Lock()
	defer c.closeCalledLock.Unlock()
	c.closeCalled = true

	return err
}

func (c *fakeTCPConn) CloseCalled() bool {
	c.closeCalledLock.Lock()
	defer c.closeCalledLock.Unlock()
	return c.closeCalled
}

func TestHandlerHandle(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			clientConn, serverConn := net.Pipe()
			closeInterceptingServerConn := newFakeTCPConn(serverConn)
			deadline := time.Now().Add(10 * time.Second)
			_ = clientConn.SetDeadline(deadline)
			_ = serverConn.SetDeadline(deadline)

			handler := redis.NewHandler(
				func() (redis.TCPConn, error) {
					return closeInterceptingServerConn, nil
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
				func() bool { return closeInterceptingServerConn.CloseCalled() },
				3*time.Second,
				5*time.Millisecond,
			) {
				t.Error(
					"Handler.Handle(): expected server connection to be " +
						"closed but was not")
			}
		})
	}

	t.Run(
		"edge case: when tcpConnAccepter returns error, then no attempt to "+
			"read the conn is made",
		func(t *testing.T) {
			t.Parallel()

			conn := newFailingOnReadTCPConn(errors.New("TCP conn blew up"))
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
			t.Parallel()

			clientConn, serverConn := net.Pipe()
			conn := newFailingOnWriteTCPConn(serverConn, errors.New("TCP conn blew up"))
			handler := redis.NewHandler(
				func() (redis.TCPConn, error) {
					return conn, nil
				},
			)

			go handler.Handle()
			if _, err := io.WriteString(clientConn, "PING\r\n"); err != nil {
				t.Fatalf("request not sent through clientConn: %v", err)
			}

			if !await.Until(
				func() bool { return conn.CloseCalled() },
				3*time.Second,
				5*time.Millisecond,
			) {
				t.Error(
					"Handler.Handle(): expected server connection to be " +
						"closed but was not")
			}
		})
}
