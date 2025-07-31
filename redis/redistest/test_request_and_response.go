package redistest

import (
	"io"
	"net"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/iox"
)

func TestRequestAndResponse(t *testing.T, conn net.Conn, request, response string) {
	t.Helper()

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
