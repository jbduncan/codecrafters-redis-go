package redis_test

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

type fakeTCPConn struct {
	readBuffer         *bytes.Buffer
	writeBuffer        strings.Builder
	closeCalled        bool
	writeErrorToReturn error
}

func (c *fakeTCPConn) Read(b []byte) (n int, err error) {
	return c.readBuffer.Read(b)
}

func (c *fakeTCPConn) Write(b []byte) (n int, err error) {
	if c.writeErrorToReturn != nil {
		return 0, c.writeErrorToReturn
	}
	return c.writeBuffer.Write(b)
}

func (c *fakeTCPConn) Close() error {
	c.closeCalled = true
	return nil
}

func (c *fakeTCPConn) WrittenOutput() string {
	return c.writeBuffer.String()
}

func TestHandlerHandle(t *testing.T) {
	tests := []struct {
		name    string
		request string
	}{
		{
			name:    "PING: uppercase simple string request",
			request: "PING\r\n",
		},
		{
			name:    "PING: lowercase simple string request",
			request: "ping\r\n",
		},
		{
			name:    "PING: array request",
			request: "*1\r\n$4\r\nping\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &fakeTCPConn{
				readBuffer: bytes.NewBufferString(tt.request),
			}
			handler := redis.NewHandler(
				func() (redis.TCPConn, error) {
					return conn, nil
				},
			)

			handler.Handle()

			if got, want := conn.WrittenOutput(), "+PONG\r\n"; got != want {
				t.Errorf(
					`Handler.Handler(): PING request: got response %q, want %q`,
					strconv.Quote(got),
					strconv.Quote(want),
				)
			}
			if !conn.closeCalled {
				t.Error(
					"Handler.Handle(): expected to call conn.Close() but " +
						"did not")
			}
		})
	}

	t.Run(
		"edge case: when tcpConnAccepter returns error, then no attempt to "+
			"write to conn is made",
		func(t *testing.T) {
			conn := &fakeTCPConn{
				readBuffer: bytes.NewBufferString("PING\r\n"),
			}
			handler := redis.NewHandler(
				func() (redis.TCPConn, error) {
					return conn, errors.New("TCP conn blew up")
				},
			)

			handler.Handle()

			if got := conn.WrittenOutput(); len(got) != 0 {
				t.Errorf(
					`Handler.Handler(): got response %s, want ""`,
					strconv.Quote(got),
				)
			}
		})

	t.Run(
		"edge case: when TCP conn returns error on write, then conn is closed",
		func(t *testing.T) {
			conn := &fakeTCPConn{
				writeErrorToReturn: errors.New("TCP conn blew up"),
			}
			handler := redis.NewHandler(
				func() (redis.TCPConn, error) {
					return conn, nil
				},
			)

			handler.Handle()

			if !conn.closeCalled {
				t.Error(
					"Handler.Handle(): expected to call conn.Close() but " +
						"did not")
			}
		})
}
