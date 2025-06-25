package redis_test

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

func TestEchoHandler_Command(t *testing.T) {
	if got, want := (redis.EchoHandler{}).Command(), "ECHO"; got != want {
		t.Errorf("Command() = %q, want %q", got, want)
	}
}

func TestEchoHandler_Handle(t *testing.T) {
	tests := []struct {
		name    string
		args    redis.Array[redis.BulkString]
		want    redis.Value
		wantErr redis.ErrorValue
	}{
		{
			name:    "No arguments",
			args:    redis.Array[redis.BulkString]{},
			wantErr: redis.NewBulkError("-ERR wrong number of arguments for 'echo' command"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redis.EchoHandler{}.Handle(tt.args)

			testEqual(t, got, tt.want, tt.wantErr)
		})
	}
}
