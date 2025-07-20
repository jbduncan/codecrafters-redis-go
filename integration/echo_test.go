//go:build integration

package integration_test

import (
	"io"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/iox"
)

func TestEcho(t *testing.T) {
	for _, tt := range []struct {
		name     string
		request  string
		response string
	}{
		{
			name:     "echo foo",
			request:  "*2\r\n$4\r\necho\r\n$3\r\nfoo\r\n",
			response: "$3\r\nfoo\r\n",
		},
		{
			name:     "ECHO foo",
			request:  "*2\r\n$4\r\nECHO\r\n$3\r\nfoo\r\n",
			response: "$3\r\nfoo\r\n",
		},
		{
			name:     "ECHO quux",
			request:  "*2\r\n$4\r\nECHO\r\n$4\r\nquux\r\n",
			response: "$4\r\nquux\r\n",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			runServer(t)
			conn := mustDialServer(t)

			if _, err := io.WriteString(conn, tt.request); err != nil {
				t.Fatalf("request %q not sent: %v", tt.request, err)
			}

			got, err := iox.ReadExactly(conn, len(tt.response))
			if err != nil {
				t.Errorf("response not read: %v", err)
			}
			if want := tt.response; got != want {
				t.Errorf(
					`got response %q, want %q`, got, want,
				)
			}
		})
	}
}
