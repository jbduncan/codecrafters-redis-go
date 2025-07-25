//go:build integration

package integration_test

import (
	"io"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/iox"
)

func TestEcho(t *testing.T) {
	const (
		echoFoo = "*2\r\n$4\r\necho\r\n$3\r\nfoo\r\n"
		foo     = "$3\r\nfoo\r\n"

		ECHOQuux = "*2\r\n$4\r\nECHO\r\n$4\r\nquux\r\n"
		quux     = "$4\r\nquux\r\n"
	)

	for _, tt := range []struct {
		name     string
		request  string
		response string
	}{
		{
			name:     "echo foo",
			request:  echoFoo,
			response: foo,
		},
		{
			name:     "ECHO quux",
			request:  ECHOQuux,
			response: quux,
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
