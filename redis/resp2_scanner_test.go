package redis_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
	"github.com/google/go-cmp/cmp"
)

type badReader struct{}

func (r badReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("ganondorf stole the triforce")
}

func TestRESP2Scanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          io.Reader
		want           redis.Value
		wantErr        error
		wantGenericErr bool
	}{
		{
			name:  "Uppercase simple string",
			input: strings.NewReader("PING\r\n"),
			want:  redis.SimpleString("PING"),
		},
		{
			name:  "Lowercase simple string",
			input: strings.NewReader("ping\r\n"),
			want:  redis.SimpleString("ping"),
		},
		{
			name:  "Other simple string",
			input: strings.NewReader("Foobar\r\n"),
			want:  redis.SimpleString("Foobar"),
		},
		{
			name:  "Missing CRLF for simple string",
			input: strings.NewReader("PING"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Missing CR for simple string",
			input: strings.NewReader("PING\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Missing LF for simple string",
			input: strings.NewReader("PING\r"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:           "Simple string that returns error on LF",
			input:          io.MultiReader(strings.NewReader("PING\r"), badReader{}),
			wantGenericErr: true,
		},
		{
			name:  "Integer 0",
			input: strings.NewReader(":0\r\n"),
			want:  redis.Integer(0),
		},
		{
			name:  "Integer 1",
			input: strings.NewReader(":1\r\n"),
			want:  redis.Integer(1),
		},
		{
			name:  "Integer 10",
			input: strings.NewReader(":10\r\n"),
			want:  redis.Integer(10),
		},
		{
			name:  "Integer -1",
			input: strings.NewReader(":-1\r\n"),
			want:  redis.Integer(-1),
		},
		{
			name:  "Integer +1",
			input: strings.NewReader(":+1\r\n"),
			want:  redis.Integer(1),
		},
		{
			name:  "Integer with invalid sign symbol",
			input: strings.NewReader(":?1\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid integer with no digits",
			input: strings.NewReader(":\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid integer with plus sign but no digits",
			input: strings.NewReader(":+\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid integer with minus sign but no digits",
			input: strings.NewReader(":-\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid integer with two plus signs",
			input: strings.NewReader(":++1\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid integer with two plus signs",
			input: strings.NewReader(":--1\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid integer with missing CR",
			input: strings.NewReader(":1\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid integer with missing LF",
			input: strings.NewReader(":1\r"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:           "Integer that returns error on second byte",
			input:          io.MultiReader(strings.NewReader(":"), badReader{}),
			wantGenericErr: true,
		},
		{
			name:           "Integer that returns error after the sign",
			input:          io.MultiReader(strings.NewReader(":+"), badReader{}),
			wantGenericErr: true,
		},
		{
			name:           "Integer that returns error after the first digit",
			input:          io.MultiReader(strings.NewReader(":1"), badReader{}),
			wantGenericErr: true,
		},
		{
			name:  "Uppercase bulk string",
			input: strings.NewReader("$4\r\nPING\r\n"),
			want:  redis.BulkString("PING"),
		},
		{
			name:  "Lowercase bulk string",
			input: strings.NewReader("$4\r\nping\r\n"),
			want:  redis.BulkString("ping"),
		},
		{
			name:  "Other bulk string",
			input: strings.NewReader("$6\r\nFoobar\r\n"),
			want:  redis.BulkString("Foobar"),
		},
		{
			name:  "Longer bulk string",
			input: strings.NewReader("$10\r\nenumerable\r\n"),
			want:  redis.BulkString("enumerable"),
		},
		{
			name:  "Bulk string with CR",
			input: strings.NewReader("$1\r\n\r\r\n"),
			want:  redis.BulkString("\r"),
		},
		{
			name:  "Bulk string with LF",
			input: strings.NewReader("$1\r\n\n\r\n"),
			want:  redis.BulkString("\n"),
		},
		{
			name:  "Invalid bulk string with just the starting $",
			input: strings.NewReader("$"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid bulk string without length",
			input: strings.NewReader("$foo\r\nbar\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid bulk string with missing first CR",
			input: strings.NewReader("$3\nfoo\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid bulk string with missing first LF",
			input: strings.NewReader("$3\rfoo\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid bulk string with missing second CR",
			input: strings.NewReader("$3\r\nfoo\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid bulk string with missing second LF",
			input: strings.NewReader("$3\r\nfoo\r"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid bulk string with shorter payload than length suggests",
			input: strings.NewReader("$2\r\na\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid bulk string with less bytes in total than length suggests",
			input: strings.NewReader("$6\r\na\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid bulk string with just a length",
			input: strings.NewReader("$1"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid bulk string with just length",
			input: strings.NewReader("$1"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid bulk string with just length and first CRLF",
			input: strings.NewReader("$1\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:           "Bulk string that returns error on first digit",
			input:          io.MultiReader(strings.NewReader("$"), badReader{}),
			wantGenericErr: true,
		},
		{
			name:           "Bulk string that returns error on first CR",
			input:          io.MultiReader(strings.NewReader("$1"), badReader{}),
			wantGenericErr: true,
		},
		{
			name:           "Bulk string that returns error on first payload byte",
			input:          io.MultiReader(strings.NewReader("$1\r\n"), badReader{}),
			wantGenericErr: true,
		},
		{
			name:  "RESP2 null bulk string",
			input: strings.NewReader("$-1\r\n"),
			want:  redis.RESP2NullBulkString{},
		},
		{
			name:  "Invalid RESP2 null bulk string with incorrect length",
			input: strings.NewReader("$-2\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid RESP2 null bulk string with missing CR",
			input: strings.NewReader("$-1\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid RESP2 null bulk string with missing LF",
			input: strings.NewReader("$-1\r"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "RESP2 null array",
			input: strings.NewReader("*-1\r\n"),
			want:  redis.RESP2NullArray{},
		},
		{
			name:  "Invalid RESP2 null array with incorrect length",
			input: strings.NewReader("*-2\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid RESP2 null array with missing CR",
			input: strings.NewReader("*-1\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:  "Invalid RESP2 null array with missing LF",
			input: strings.NewReader("*-1\r"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("SYNTAX invalid syntax"),
		},
		{
			name:    "No input",
			input:   strings.NewReader(""),
			wantErr: io.EOF,
		},
		{
			name:           "Error-returning input",
			input:          badReader{},
			wantGenericErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := redis.NewRESP2Scanner(tt.input)
			got, err := s.Scan()

			if tt.wantGenericErr {
				if err == nil {
					t.Fatalf("Scan(): got a <nil> error, want a non-nil error")
				}
				if _, ok := err.(redis.Value); ok {
					t.Fatalf(
						"Scan(): got err %q, want a generic error that is "+
							"not a redis.Value",
						err,
					)
				}
				return
			}

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Scan(): got err %q, want %q", err, tt.wantErr)
				}
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Scan() mismatch (-want +got):\n%s", diff)
			}

			secondGot, secondScanErr := s.Scan()

			if !errors.Is(secondScanErr, io.EOF) {
				t.Fatalf(
					"Scan(): nothing else should have been read: "+
						"got %q, err %q; want io.EOF",
					secondGot, secondScanErr)
			}
		})
	}

	t.Run("Pipelined inputs", func(t *testing.T) {
		t.Parallel()

		input := "PING\r\nPING\r\n"
		want := redis.SimpleString("PING")
		s := redis.NewRESP2Scanner(strings.NewReader(input))
		for range 2 {
			got, _ := s.Scan()
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Scan() mismatch (-want +got):\n%s", diff)
			}
		}
	})
}
