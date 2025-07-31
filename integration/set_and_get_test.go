//go:build integration

package integration_test

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis/redistest"
)

func TestSetAndGet(t *testing.T) {
	t.Skip("To un-skip later when set-and-get is implemented")

	for _, tt := range []struct {
		name        string
		setRequest  string
		setResponse string
		getRequest  string
		getResponse string
	}{
		{
			name:        "set foo; get foo",
			setRequest:  redistest.RequestSetLowercaseFoo,
			setResponse: redistest.ResponseOK,
			getRequest:  redistest.RequestGetLowercaseFoo,
			getResponse: redistest.ResponseFoo,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			runServer(t)
			conn := mustDialServer(t)

			redistest.TestRequestAndResponse(t, conn, tt.setRequest, tt.setResponse)
			redistest.TestRequestAndResponse(t, conn, tt.getRequest, tt.getResponse)
		})
	}
}
