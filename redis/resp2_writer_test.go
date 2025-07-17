package redis_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
	"github.com/google/go-cmp/cmp"
)

func TestRESP2Writer_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input redis.Value
		want  string
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
		{
			name:  "Simple error #1",
			input: redis.MakeSimpleError("ERR foo"),
			want:  "-ERR foo\r\n",
		},
		{
			name:  "Simple error #1",
			input: redis.MakeSimpleError("ERR bar"),
			want:  "-ERR bar\r\n",
		},
		{
			name:  "Null bulk string",
			input: redis.NullBulkString{},
			want:  "$-1\r\n",
		},
		{
			name:  "Bulk error #1",
			input: redis.MakeBulkError("ERR foo"),
			want:  "!7\r\nERR foo\r\n",
		},
		{
			name:  "Bulk error #2",
			input: redis.MakeBulkError("ERR blah"),
			want:  "!8\r\nERR blah\r\n",
		},
		{
			name:  "Array #1",
			input: redis.Array{},
			want:  "*0\r\n",
		},
		{
			name:  "Array #2",
			input: redis.Array{redis.SimpleString("PING")},
			want:  "*1\r\n+PING\r\n",
		},
		{
			name:  "Array #3",
			input: redis.Array{redis.SimpleString("PONG")},
			want:  "*1\r\n+PONG\r\n",
		},
		{
			name:  "Array #4",
			input: redis.Array{redis.BulkString("PING")},
			want:  "*1\r\n$4\r\nPING\r\n",
		},
		{
			name: "Array #5",
			input: redis.Array{
				redis.SimpleString("FOO"),
				redis.SimpleString("BAR"),
			},
			want: "*2\r\n+FOO\r\n+BAR\r\n",
		},
		{
			name:  "Null array",
			input: redis.NullArray{},
			want:  "*-1\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var writer strings.Builder
			err := redis.NewRESP2Writer(&writer).Write(tt.input)
			got := writer.String()

			if err != nil {
				t.Fatalf("Write(): got err %q, want <nil> err", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Write() mismatch (-want +got):\n%s", diff)
			}
		})
	}

	t.Run("Error on writing array length", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("ganondorf stole the triforce")
		writer := &limitedWriter{w: io.Discard, n: 0, e: wantErr}

		gotErr := redis.NewRESP2Writer(writer).Write(redis.Array{})

		if !errors.Is(gotErr, wantErr) {
			t.Errorf("Write(): got err %q, want %q", gotErr, wantErr)
		}
	})

	t.Run("Error on writing array element", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("ganondorf stole the triforce")
		writer := &limitedWriter{w: io.Discard, n: 4, e: wantErr}

		gotErr := redis.NewRESP2Writer(writer).Write(redis.Array{redis.Array{}})

		if !errors.Is(gotErr, wantErr) {
			t.Errorf("Write(): got err %q, want %q", gotErr, wantErr)
		}
	})

	t.Run("Error on writing simple string", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("ganondorf stole the triforce")
		writer := &limitedWriter{w: io.Discard, n: 0, e: wantErr}

		gotErr := redis.NewRESP2Writer(writer).Write(redis.SimpleString("FOO"))

		if !errors.Is(gotErr, wantErr) {
			t.Errorf("Write(): got err %q, want %q", gotErr, wantErr)
		}
	})
}

// Based on: https://github.com/golang/go/issues/54111#issuecomment-1220793565
type limitedWriter struct {
	w io.Writer
	n int64
	e error
}

func (lw *limitedWriter) Write(p []byte) (int, error) {
	if lw.n < 1 {
		return 0, lw.e
	}
	if lw.n < int64(len(p)) {
		p = p[:lw.n]
	}
	n, err := lw.w.Write(p)
	lw.n -= int64(n)
	return n, err
}
