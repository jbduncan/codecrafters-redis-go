package redis_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
	"github.com/google/go-cmp/cmp"
)

func TestRESP2Writer_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   redis.Value
		want    string
		wantErr error
	}{
		{
			name:  "Simple string #1",
			input: redis.SimpleString("PING"),
			want:  "+PING\r\n",
		},
		{
			name:  "Simple string #2",
			input: redis.SimpleString("PONG"),
			want:  "+PONG\r\n",
		},
		{
			name:  "Integer #1",
			input: redis.Integer(0),
			want:  ":0\r\n",
		},
		{
			name:  "Integer #2",
			input: redis.Integer(-1),
			want:  ":-1\r\n",
		},
		{
			name:  "Bulk string #1",
			input: redis.BulkString("PING"),
			want:  "$4\r\nPING\r\n",
		},
		{
			name:  "Bulk string #2",
			input: redis.BulkString("PONG"),
			want:  "$4\r\nPONG\r\n",
		},
		{
			name:  "Bulk string #3",
			input: redis.BulkString("GET"),
			want:  "$3\r\nGET\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var writer strings.Builder
			err := redis.NewRESP2Writer(&writer).Write(tt.input)
			got := writer.String()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Write(): got err %q, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Write(): got err %q, want <nil> err", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Write() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
