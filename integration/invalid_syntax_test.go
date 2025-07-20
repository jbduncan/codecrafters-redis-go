//go:build integration

package integration_test

import (
	"io"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/iox"
)

// TODO: test '£'
func TestInvalidSyntax(t *testing.T) {
	request := "*1\r\n^"
	response := "!42\r\n-ERR Protocol error: expected '$', got '^'\r\n"

	runServer(t)
	conn := mustDialServer(t)

	if _, err := io.WriteString(conn, request); err != nil {
		t.Fatalf("request %q not sent: %v", request, err)
	}

	got, err := iox.ReadExactly(conn, len(response))
	if err != nil {
		t.Errorf("response not read: %v", err)
	}
	if want := response; got != want {
		t.Errorf(
			`got response %q, want %q`, got, want,
		)
	}
}
