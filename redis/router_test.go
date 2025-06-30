package redis_test

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
	"github.com/google/go-cmp/cmp"
)

func TestRouter_Route(t *testing.T) {
	tests := []struct {
		name     string
		handler  *MockHandler
		input    redis.Value
		want     redis.Value
		wantArgs redis.Array[redis.BulkString]
	}{
		{
			name: `"PING": "PONG"`,
			handler: &MockHandler{
				CommandFunc: func() string {
					return "PING"
				},
				HandleFunc: func(args redis.Array[redis.BulkString]) redis.Value {
					return redis.SimpleString("PONG")
				},
			},
			input:    redis.SimpleString("PING"),
			want:     redis.SimpleString("PONG"),
			wantArgs: redis.Array[redis.BulkString]{},
		},
		{
			name: `["ECHO", "Hello, world"]: "Hello, world"`,
			handler: &MockHandler{
				CommandFunc: func() string {
					return "ECHO"
				},
				HandleFunc: func(args redis.Array[redis.BulkString]) redis.Value {
					return redis.BulkString("Hello, world")
				},
			},
			input:    redis.Array[redis.BulkString]{"ECHO", "Hello, world"},
			want:     redis.Array[redis.BulkString]{"Hello, world"},
			wantArgs: redis.Array[redis.BulkString]{"Hello, world"},
		},
		{
			name: `"PING": "DING"`,
			handler: &MockHandler{
				CommandFunc: func() string {
					return "PING"
				},
				HandleFunc: func(args redis.Array[redis.BulkString]) redis.Value {
					return redis.SimpleString("DING")
				},
			},
			input:    redis.SimpleString("PING"),
			want:     redis.SimpleString("DING"),
			wantArgs: redis.Array[redis.BulkString]{},
		},
		{
			name: `["ECHO", "Bye, world"]: "Hello, world"`,
			handler: &MockHandler{
				CommandFunc: func() string {
					return "ECHO"
				},
				HandleFunc: func(args redis.Array[redis.BulkString]) redis.Value {
					return redis.BulkString("Hello, world")
				},
			},
			input:    redis.Array[redis.BulkString]{"ECHO", "Bye, world"},
			want:     redis.Array[redis.BulkString]{"Hello, world"},
			wantArgs: redis.Array[redis.BulkString]{"Bye, world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := redis.NewRouter([]redis.Handler{tt.handler}).Route(tt.input)

			testEqual(t, result, tt.want, nil)
			testHandleCalledWith(t, tt.handler, tt.wantArgs)
		})
	}
}

func testHandleCalledWith(
	t *testing.T,
	handler *MockHandler,
	wantArgs redis.Array[redis.BulkString],
) {
	t.Helper()

	calls := handler.HandleCalls()
	if got, want := len(calls), 1; got != want {
		t.Fatalf(
			"number of Handler.Handle() calls: got %d, want %d",
			got,
			want,
		)
	}
	got := calls[0].Args
	if diff := cmp.Diff(wantArgs, got); diff != "" {
		t.Errorf("Route() args mismatch (-want +got):\n%s", diff)
	}
}
