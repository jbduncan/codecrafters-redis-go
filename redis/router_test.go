package redis_test

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

func TestRouter_Route(t *testing.T) {
	tests := []struct {
		name    string
		input   redis.InputValue
		want    redis.Value
		wantErr redis.ErrorValue
	}{
		{
			name:  `s"PING": s"PONG"`,
			input: redis.SimpleString("PING"),
			want:  redis.SimpleString("PONG"),
		},
		{
			name:  `s"ping": s"PONG"`,
			input: redis.SimpleString("ping"),
			want:  redis.SimpleString("PONG"),
		},
		{
			name:    `s"FOO": error`,
			input:   redis.SimpleString("FOO"),
			wantErr: redis.MakeBulkError("ERR unknown command `FOO`"),
		},
		{
			name:    `s"BAR": error`,
			input:   redis.SimpleString("BAR"),
			wantErr: redis.MakeBulkError("ERR unknown command `BAR`"),
		},
		{
			name:  `[b"PING"]: s"PONG"`,
			input: redis.BulkStringArray{"PING"},
			want:  redis.SimpleString("PONG"),
		},
		{
			name:  `[b"ping"]: s"PONG"`,
			input: redis.BulkStringArray{"PING"},
			want:  redis.SimpleString("PONG"),
		},
		{
			name:  `[b"PING", b"Hello, world"]: b"Hello, world"`,
			input: redis.BulkStringArray{"PING", "Hello, world"},
			want:  redis.BulkString("Hello, world"),
		},
		{
			name:  `[b"ECHO", b"Hello, world"]: b"Hello, world"`,
			input: redis.BulkStringArray{"ECHO", "Hello, world"},
			want:  redis.BulkString("Hello, world"),
		},
		{
			name:  `[b"ECHO", b"Bye, world"]: b"Bye, world"`,
			input: redis.BulkStringArray{"ECHO", "Bye, world"},
			want:  redis.BulkString("Bye, world"),
		},
		{
			name:  `[b"echo", b"Hello, world"]: b"Hello, world"`,
			input: redis.BulkStringArray{"echo", "Hello, world"},
			want:  redis.BulkString("Hello, world"),
		},
		{
			name:    `[b"ECHO", b"Hello, world", b"Bye, world"]: error`,
			input:   redis.BulkStringArray{"ECHO", "Hello, world", "Bye, world"},
			wantErr: redis.MakeBulkError("ERR wrong number of arguments for 'echo' command"),
		},
		{
			name:    `[b"echo", b"Hello, world", b"Bye, world"]: error`,
			input:   redis.BulkStringArray{"echo", "Hello, world", "Bye, world"},
			wantErr: redis.MakeBulkError("ERR wrong number of arguments for 'echo' command"),
		},
		{
			name:    `[b"FOO"]: error`,
			input:   redis.BulkStringArray{"FOO"},
			wantErr: redis.MakeBulkError("ERR unknown command `FOO`"),
		},
		{
			name:    `[b"BAR"]: error`,
			input:   redis.BulkStringArray{"BAR"},
			wantErr: redis.MakeBulkError("ERR unknown command `BAR`"),
		},
		{
			name:    `[b""]: error`,
			input:   redis.BulkStringArray{""},
			wantErr: redis.MakeBulkError("ERR unknown command ``"),
		},
		{
			name:    `[]: error`,
			input:   redis.BulkStringArray{},
			wantErr: redis.MakeBulkError("ERR no command"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result :=
				redis.NewDefaultRouter(
					redis.EchoHandler{},
					redis.PingHandler{},
				).Route(tt.input)

			testEqual(t, result, tt.want, tt.wantErr)
		})
	}
}
