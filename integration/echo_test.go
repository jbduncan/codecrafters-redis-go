//go:build integration

package integration_test

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis/redistest"
)

func TestEcho(t *testing.T) {

	for _, tt := range []struct {
		name     string
		request  string
		response string
	}{
		{
			name:     "echo foo",
			request:  redistest.RequestEchoLowercaseFoo,
			response: redistest.ResponseFoo,
		},
		{
			name:     "ECHO quux",
			request:  redistest.RequestEchoUppercaseQuux,
			response: redistest.ResponseQuux,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			runServer(t)
			conn := mustDialServer(t)

			redistest.TestRequestAndResponse(t, conn, tt.request, tt.response)
		})
	}
}
