package redis_test

import (
	"errors"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
	"github.com/google/go-cmp/cmp"
)

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
			got := redis.PingHandlerFunc(tt.args)

			if tt.wantErr != nil {
				gotErr, ok := got.(error)
				if !ok {
					t.Fatalf("PingHandlerFunc(): got %v, want err %q", got, tt.wantErr)
				}

				var gotErrValue redis.ErrorValue
				if !errors.As(gotErr, &gotErrValue) {
					t.Fatalf("PingHandlerFunc(): got err %q, want err %q", gotErr, tt.wantErr)
				}

				if got, want := gotErrValue.Message(), tt.wantErr.Message(); got != want {
					t.Fatalf("PingHandlerFunc(): got err %q, want err %q", gotErr, tt.wantErr)
				}
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("PingHandlerFunc() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
