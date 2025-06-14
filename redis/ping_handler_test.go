package redis_test

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

// TODO: consider extracting out an interface test suite for future handlers

func TestPingHandler_Command(t *testing.T) {
	if got, want := (redis.PingHandler{}).Command(), "PING"; got != want {
		t.Errorf("Command() = %v, want %v", got, want)
	}
}

func TestPingHandler_Handle(t *testing.T) {
	tests := []struct {
		name    string
		args    redis.Array
		want    redis.Value
		wantErr redis.ErrorValue
	}{
		{
			name: "No arguments",
			args: redis.Array{},
			want: redis.SimpleString("PONG"),
		},
		{
			name: `"Hello, world"`,
			args: redis.Array{
				redis.BulkString("Hello, world"),
			},
			want: redis.BulkString("Hello, world"),
		},
		{
			name: `"It's dangerous to go alone! Take this."`,
			args: redis.Array{
				redis.BulkString("It's dangerous to go alone! Take this."),
			},
			want: redis.BulkString("It's dangerous to go alone! Take this."),
		},
		{
			name: "Two arguments",
			args: redis.Array{
				redis.BulkString("foo"),
				redis.BulkString("bar"),
			},
			wantErr: redis.NewBulkError("-ERR wrong number of arguments for 'echo' command"),
		},
		{
			name: "Three arguments",
			args: redis.Array{
				redis.BulkString("foo"),
				redis.BulkString("bar"),
				redis.BulkString("baz"),
			},
			wantErr: redis.NewBulkError("-ERR wrong number of arguments for 'echo' command"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redis.PingHandler{}.Handle(tt.args)

			testEqual(t, got, tt.want, tt.wantErr)
		})
	}
}
