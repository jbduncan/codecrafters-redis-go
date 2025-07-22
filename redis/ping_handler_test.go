package redis_test

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

func TestPingHandler_Handle(t *testing.T) {
	tests := []struct {
		name    string
		args    redis.BulkStringArray
		want    redis.Value
		wantErr redis.ErrorValue
	}{
		{
			name: "No arguments",
			args: redis.BulkStringArray{},
			want: redis.SimpleString("PONG"),
		},
		{
			name: `"Hello, world"`,
			args: redis.BulkStringArray{
				redis.BulkString("Hello, world"),
			},
			want: redis.BulkString("Hello, world"),
		},
		{
			name: `"It's dangerous to go alone! Take this."`,
			args: redis.BulkStringArray{
				redis.BulkString("It's dangerous to go alone! Take this."),
			},
			want: redis.BulkString("It's dangerous to go alone! Take this."),
		},
		{
			name: "Two arguments",
			args: redis.BulkStringArray{
				redis.BulkString("foo"),
				redis.BulkString("bar"),
			},
			wantErr: redis.MakeBulkError("ERR wrong number of arguments for 'ping' command"),
		},
		{
			name: "Three arguments",
			args: redis.BulkStringArray{
				redis.BulkString("foo"),
				redis.BulkString("bar"),
				redis.BulkString("baz"),
			},
			wantErr: redis.MakeBulkError("ERR wrong number of arguments for 'ping' command"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redis.PingHandler{}.Handle(tt.args)

			testEqual(t, got, tt.want, tt.wantErr)
		})
	}
}
