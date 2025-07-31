//go:build integration

package integration_test

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis/redistest"
)

// TODO: test '£'
func TestInvalidSyntax(t *testing.T) {
	request := redistest.RequestInvalidSyntax
	response := redistest.ResponseInvalidSyntaxError

	runServer(t)
	conn := mustDialServer(t)

	redistest.TestRequestAndResponse(t, conn, request, response)
}
