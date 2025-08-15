package integration_test

import (
	"net"
	"strings"
	"testing"

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
			request:  redistest.RequestPingLowercase,
			response: redistest.ResponsePong,
		},
		{
			name:     "PING",
			request:  redistest.RequestPingUppercase,
			response: redistest.ResponsePong,
		},
		// TODO: move to pipelined_requests_test.go
		{
			name:     "Pipeline",
			request:  strings.Repeat(redistest.RequestPingLowercase, 3),
			response: strings.Repeat(redistest.ResponsePong, 3),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			runServer(t)
			conn := mustDialServer(t)

			redistest.TestRequestAndResponse(t, conn, tt.request, tt.response)
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
