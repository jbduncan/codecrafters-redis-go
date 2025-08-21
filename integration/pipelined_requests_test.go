package integration_test

import (
	"strings"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis/redistest"
)

func TestPipelinedRequests(t *testing.T) {
	runServer(t)

	conn := mustDialServer(t)

	request := strings.Repeat(redistest.RequestPingLowercase, 3)
	response := strings.Repeat(redistest.ResponsePong, 3)
	redistest.TestRequestAndResponse(t, conn, request, response)
}
