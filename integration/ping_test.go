//go:build integration

package integration_test

import (
	"io"
	"net"
	"strings"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/iox"
	"github.com/codecrafters-io/redis-starter-go/redis/redistest"
)

func TestPing(t *testing.T) {
	for _, tt := range []struct {
		name     string
		request  string
		response string
	}{
		{
			name:     "ping",
			request:  redistest.PingLowercase,
			response: redistest.Pong,
		},
		{
			name:     "PING",
			request:  redistest.PingUppercase,
			response: redistest.Pong,
		},
		// TODO: move to pipelined_requests_test.go
		{
			name:     "Pipeline",
			request:  strings.Repeat(redistest.PingLowercase, 3),
			response: strings.Repeat(redistest.Pong, 3),
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

	// TODO: move to concurrent_requests_test.go
	for _, tt := range []struct {
		name            string
		concurrentPings int
	}{
		{
			name:            "two concurrent PINGs",
			concurrentPings: 2,
		},
		{
			name:            "1,000 concurrent PINGs",
			concurrentPings: 1_000,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			runServer(t)

			redistest.TestConcurrentPings(
				t,
				tt.concurrentPings,
				func() net.Conn {
					return mustDialServer(t)
				})
		})
	}
}
