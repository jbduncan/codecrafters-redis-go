package integration_test

import (
	"io"
	"strconv"
	"testing"
)

func TestPing(t *testing.T) {
	stop := runServer(t)
	defer stop()
	conn, err := dialServer()
	if err != nil {
		t.Fatalf("no connection to server: %v", err)
	}

	msg := "*1\r\n$4\r\nPING\r\n"
	_, err = io.WriteString(conn, msg)
	if err != nil {
		t.Fatalf("did not write message %q successfully: %v", msg, err)
	}

	b, err := io.ReadAll(conn)
	if err != nil {
		// TODO: try uncommenting once we can handle many connections
		// t.Fatalf("did not read conn successfully: %v", err)
	}
	if got, want := string(b), "+PONG\r\n"; got != want {
		t.Errorf(
			`PING request: got response %q, want %q`,
			strconv.Quote(got),
			strconv.Quote(want),
		)
	}
}
