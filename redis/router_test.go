package redis_test

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

func TestRouter_Route(t *testing.T) {
	tests := []struct {
		name     string
		handlers []redis.Handler
		value    redis.Value
		want     redis.Value
		wantErr  redis.ErrorValue
	}{
		{
			name:     "PING simple string command and no arguments",
			handlers: []redis.Handler{redis.PingHandler{}},
			value:    redis.SimpleString("PING"),

			want: redis.SimpleString("PONG"),
		},
		{
			name:     "Array with PING bulk string command and one argument #1",
			handlers: []redis.Handler{redis.PingHandler{}},
			value: redis.Array[redis.BulkString]{
				redis.BulkString("PING"),
				redis.BulkString("Hello, world"),
			},
			want: redis.BulkString("Hello, world"),
		},
		{
			name:     "Array with PING bulk string command and one argument #2",
			handlers: []redis.Handler{redis.PingHandler{}},
			value: redis.Array[redis.BulkString]{
				redis.BulkString("PING"),
				redis.BulkString("Bye, world"),
			},
			want: redis.BulkString("Bye, world"),
		},
		{
			name:     "Array with PING bulk string command and two arguments",
			handlers: []redis.Handler{redis.PingHandler{}},
			value: redis.Array[redis.BulkString]{
				redis.BulkString("PING"),
				redis.BulkString("Hello, world"),
				redis.BulkString("Bye, world"),
			},
			wantErr: redis.NewBulkError("-ERR wrong number of arguments for 'ping' command"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redis.NewRouter(tt.handlers).Route(tt.value)

			testEqual(t, got, tt.want, tt.wantErr)
		})
	}
}
