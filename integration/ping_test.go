package integration_test

import (
	"io"
	"net"
	"testing"
)

func TestPing(t *testing.T) {
	interrupt := runServer(t)
	defer interrupt()

	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		t.Fatalf("no connection to server: %v", err)
	}

	writeToConn(t, conn, "+PING\r\n")
}

func writeToConn(t *testing.T, conn net.Conn, msg string) {
	if _, err := io.WriteString(conn, msg); err != nil {
		t.Fatalf("did not write message %q successfully: %v", msg, err)
	}
}
