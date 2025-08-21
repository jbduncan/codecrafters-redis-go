package integration_test

import (
	"net"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis/redistest"
)

func TestConcurrentRequests(t *testing.T) {
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
