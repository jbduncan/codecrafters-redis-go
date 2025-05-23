package redis_test

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

type FakeTCPConn struct {
	readBuffer  *bytes.Buffer
	writeBuffer strings.Builder
	closeCalled bool
}

func (c *FakeTCPConn) Read(b []byte) (n int, err error) {
	return c.readBuffer.Read(b)
}

func (c *FakeTCPConn) Write(b []byte) (n int, err error) {
	return c.writeBuffer.Write(b)
}

func (c *FakeTCPConn) Close() error {
	c.closeCalled = true
	return nil
}

func (c *FakeTCPConn) WrittenOutput() string {
	return c.writeBuffer.String()
}

func TestHandler(t *testing.T) {
	fakeTCPConn := &FakeTCPConn{
		readBuffer: bytes.NewBufferString("PING\r\n"),
	}
	handler := redis.NewHandler(
		func() (redis.TCPConn, error) {
			return fakeTCPConn, nil
		},
	)

	handler.Handle()

	if got, want := fakeTCPConn.WrittenOutput(), "+PONG\r\n"; got != want {
		t.Errorf(
			`Handler.Handler(): PING request: got response %q, want %q`,
			strconv.Quote(got),
			strconv.Quote(want),
		)
	}
}
