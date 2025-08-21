package integration_test

import (
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
	} {
		t.Run(tt.name, func(t *testing.T) {
			runServer(t)
			conn := mustDialServer(t)

			redistest.TestRequestAndResponse(t, conn, tt.request, tt.response)
		})
	}
}
