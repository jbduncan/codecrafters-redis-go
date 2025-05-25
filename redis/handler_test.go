package redis_test

import (
	"bytes"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/await"
	"github.com/codecrafters-io/redis-starter-go/redis"
)

type fakeTCPConn struct {
	readBuffer         *bytes.Buffer
	readBufferRead     bool
	writeBuffer        *bytes.Buffer
	writeBufferLock    *sync.Mutex
	closeCalled        bool
	closeCalledLock    *sync.Mutex
	readErrorToReturn  error
	writeErrorToReturn error
}

func newFakeTCPConn(dataToSend string) *fakeTCPConn {
	return &fakeTCPConn{
		readBuffer:         bytes.NewBufferString(dataToSend),
		readBufferRead:     false,
		writeBuffer:        new(bytes.Buffer),
		writeBufferLock:    new(sync.Mutex),
		closeCalled:        false,
		closeCalledLock:    new(sync.Mutex),
		readErrorToReturn:  nil,
		writeErrorToReturn: nil,
	}
}

func newFailingOnReadTCPConn(err error) *fakeTCPConn {
	return &fakeTCPConn{
		readBuffer:         new(bytes.Buffer),
		readBufferRead:     false,
		writeBuffer:        new(bytes.Buffer),
		writeBufferLock:    new(sync.Mutex),
		closeCalled:        false,
		closeCalledLock:    new(sync.Mutex),
		readErrorToReturn:  err,
		writeErrorToReturn: nil,
	}
}

func newFailingOnWriteTCPConn(err error) *fakeTCPConn {
	return &fakeTCPConn{
		readBuffer:         new(bytes.Buffer),
		readBufferRead:     false,
		writeBuffer:        new(bytes.Buffer),
		writeBufferLock:    new(sync.Mutex),
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

	c.readBufferRead = true
	return c.readBuffer.Read(b)
}

func (c *fakeTCPConn) Write(b []byte) (n int, err error) {
	if c.writeErrorToReturn != nil {
		return 0, c.writeErrorToReturn
	}

	c.writeBufferLock.Lock()
	defer c.writeBufferLock.Unlock()
	return c.writeBuffer.Write(b)
}

func (c *fakeTCPConn) Close() error {
	c.closeCalledLock.Lock()
	defer c.closeCalledLock.Unlock()

	c.closeCalled = true
	return nil
}

func (c *fakeTCPConn) WrittenOutput() string {
	c.writeBufferLock.Lock()
	defer c.writeBufferLock.Unlock()
	return c.writeBuffer.String()
}

func (c *fakeTCPConn) CloseCalled() bool {
	c.closeCalledLock.Lock()
	defer c.closeCalledLock.Unlock()
	return c.closeCalled
}

func TestHandlerHandle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		request  string
		response string
	}{
		{
			name:     "PING: uppercase simple string request",
			request:  "PING\r\n",
			response: "+PONG\r\n",
		},
		{
			name:     "PING: lowercase simple string request",
			request:  "ping\r\n",
			response: "+PONG\r\n",
		},
		{
			name:     "PING: array request",
			request:  "*1\r\n$4\r\nPING\r\n",
			response: "+PONG\r\n",
		},
		{
			name:     "three PINGs: three pipelined simple requests",
			request:  "PING\r\nPING\r\nPING\r\n",
			response: "+PONG\r\n+PONG\r\n+PONG\r\n",
		},
		{
			name:     "three PINGs: three pipelined requests in array",
			request:  "*3\r\n$4\r\nPING\r\n$4\r\nPING\r\n$4\r\nPING\r\n",
			response: "+PONG\r\n+PONG\r\n+PONG\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conn := newFakeTCPConn(tt.request)
			handler := redis.NewHandler(
				func() (redis.TCPConn, error) {
					return conn, nil
				},
			)

			go handler.Handle()

			success, got, want := await.UntilEqual(
				func() (string, string) {
					got, want := conn.WrittenOutput(), tt.response
					return got, want
				},
				3*time.Second,
				100*time.Millisecond,
			)
			if !success {
				t.Errorf(
					`Handler.Handler(): PING request: got response %q, want %q`,
					got,
					want,
				)
			}
			if !await.Until(
				func() bool { return conn.CloseCalled() },
				3*time.Second,
				100*time.Millisecond,
			) {
				t.Error(
					"Handler.Handle(): expected to call conn.Close() but " +
						"did not")
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

			if got := conn.WrittenOutput(); len(got) != 0 {
				t.Errorf(
					`Handler.Handler(): got response %q, want ""`, got,
				)
			}
		})

	t.Run(
		"edge case: when TCP conn returns error on write, then conn is closed",
		func(t *testing.T) {
			t.Parallel()

			conn := newFailingOnWriteTCPConn(errors.New("TCP conn blew up"))
			handler := redis.NewHandler(
				func() (redis.TCPConn, error) {
					return conn, nil
				},
			)

			handler.Handle()

			if !conn.CloseCalled() {
				t.Error(
					"Handler.Handle(): expected to call conn.Close() but " +
						"did not")
			}
		})
}
