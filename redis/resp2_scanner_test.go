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
			wantErr: redis.NewSimpleError(`-ERR Protocol error: incomplete request`),
		},
		{
			name:  "Missing CR for simple string",
			input: strings.NewReader("PING\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError(`-ERR Protocol error: incomplete request`),
		},
		{
			name:  "Missing LF for simple string",
			input: strings.NewReader("PING\r"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError(`-ERR Protocol error: incomplete request`),
		},
		{
			name:           "Simple string that returns error on LF",
			input:          io.MultiReader(strings.NewReader("PING\r"), badReader{}),
			wantGenericErr: true,
		},
		{
			name:  "Empty array",
			input: strings.NewReader("*0\r\n"),
			want:  redis.Array{},
		},
		{
			name:  "One element array with uppercase bulk string",
			input: strings.NewReader("*1\r\n$4\r\nPING\r\n"),
			want: redis.Array{
				redis.BulkString("PING"),
			},
		},
		{
			name:  "One element array with lowercase bulk string",
			input: strings.NewReader("*1\r\n$4\r\nping\r\n"),
			want: redis.Array{
				redis.BulkString("ping"),
			},
		},
		{
			name:  "One element array with some other bulk string",
			input: strings.NewReader("*1\r\n$6\r\nFoobar\r\n"),
			want: redis.Array{
				redis.BulkString("Foobar"),
			},
		},
		{
			name:  "One element array with longer bulk string",
			input: strings.NewReader("*1\r\n$10\r\nenumerable\r\n"),
			want: redis.Array{
				redis.BulkString("enumerable"),
			},
		},
		{
			name:  "One element array with bulk string with CR",
			input: strings.NewReader("*1\r\n$1\r\n\r\r\n"),
			want: redis.Array{
				redis.BulkString("\r"),
			},
		},
		{
			name:  "One element array with bulk string with LF",
			input: strings.NewReader("*1\r\n$1\r\n\n\r\n"),
			want: redis.Array{
				redis.BulkString("\n"),
			},
		},
		{
			name:  "Zero element array with missing CR",
			input: strings.NewReader("*0\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError(`-ERR Protocol error: expected '\r', got '\n'`),
		},
		{
			name:  "Invalid array with missing LF",
			input: strings.NewReader("*0\r"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError(`-ERR Protocol error: incomplete request`),
		},
		{
			name:  "Invalid array with two CRs",
			input: strings.NewReader("*0\r\r"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError(`-ERR Protocol error: expected '\n', got '\r'`),
		},
		{
			name:  "One element array with missing element",
			input: strings.NewReader("*1\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError(`-ERR Protocol error: incomplete request`),
		},
		{
			name:  "One element array with invalid bulk string with just the starting $",
			input: strings.NewReader("*1\r\n$"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("-ERR Protocol error: incomplete request"),
		},
		{
			name:  "One element array with invalid bulk string without length",
			input: strings.NewReader("*1\r\n$foo\r\nbar\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("-ERR Protocol error: expected digit, got 'f'"),
		},
		{
			name:  "One element array with invalid bulk string with missing first CR",
			input: strings.NewReader("*1\r\n$3\nfoo\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError(`-ERR Protocol error: expected '\r', got '\n'`),
		},
		{
			name:  "One element array with invalid bulk string with missing first LF",
			input: strings.NewReader("*1\r\n$3\rfoo\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError(`-ERR Protocol error: expected '\n', got 'f'`),
		},
		{
			name:  "One element array with invalid bulk string with missing second CR",
			input: strings.NewReader("*1\r\n$3\r\nfoo\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError(`-ERR Protocol error: expected '\r', got '\n'`),
		},
		{
			name:  "One element array with invalid bulk string with missing second LF",
			input: strings.NewReader("*1\r\n$3\r\nfoo\r"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("-ERR Protocol error: incomplete request"),
		},
		{
			name:  "One element array with bulk string with shorter content than length suggests",
			input: strings.NewReader("*1\r\n$2\r\na\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError(`-ERR Protocol error: expected '\r', got '\n'`),
		},
		{
			name:  "One element array with bulk string with less bytes in total than length suggests",
			input: strings.NewReader("*1\r\n$6\r\na\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("-ERR Protocol error: incomplete request"),
		},
		{
			name:  "One element array with invalid bulk string with just a length",
			input: strings.NewReader("*1\r\n$1"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("-ERR Protocol error: incomplete request"),
		},
		{
			name:  "One element array with invalid bulk string with just length and first CRLF",
			input: strings.NewReader("*1\r\n$1\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("-ERR Protocol error: incomplete request"),
		},
		{
			name:           "One element array with bulk string that returns error on first digit",
			input:          io.MultiReader(strings.NewReader("*1\r\n$"), badReader{}),
			wantGenericErr: true,
		},
		{
			name:           "One element array with bulk string that returns error on first CR",
			input:          io.MultiReader(strings.NewReader("*1\r\n$1"), badReader{}),
			wantGenericErr: true,
		},
		{
			name:           "One element array with bulk string that returns error on first payload byte",
			input:          io.MultiReader(strings.NewReader("*1\r\n$1\r\n"), badReader{}),
			wantGenericErr: true,
		},
		{
			name:  "One element array with bulk string of negative length",
			input: strings.NewReader("*1\r\n$-1\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError("-ERR Protocol error: expected digit, got '-'"),
		},
		{
			name: "Ten element array of bulk strings",
			input: strings.NewReader(
				"*10\r\n" +
					"$1\r\nA\r\n" +
					"$1\r\nB\r\n" +
					"$1\r\nC\r\n" +
					"$1\r\nD\r\n" +
					"$1\r\nE\r\n" +
					"$1\r\nF\r\n" +
					"$1\r\nG\r\n" +
					"$1\r\nH\r\n" +
					"$1\r\nI\r\n" +
					"$1\r\nJ\r\n"),
			want: redis.Array{
				redis.BulkString("A"),
				redis.BulkString("B"),
				redis.BulkString("C"),
				redis.BulkString("D"),
				redis.BulkString("E"),
				redis.BulkString("F"),
				redis.BulkString("G"),
				redis.BulkString("H"),
				redis.BulkString("I"),
				redis.BulkString("J"),
			},
		},
		{
			name:  "One element array with inner array",
			input: strings.NewReader("*1\r\n*0\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError(`-ERR Protocol error: expected '$', got '*'`),
		},
		{
			name:  "One element array with negative length",
			input: strings.NewReader("*-1\r\n$1\r\na\r\n"),
			// TODO: return as a redis.BulkError
			wantErr: redis.NewSimpleError(`-ERR Protocol error: expected digit, got '-'`),
		},
		{
			name:    "One element array with integer array element",
			input:   strings.NewReader("*1\r\n:1\r\n"),
			wantErr: redis.NewSimpleError(`-ERR Protocol error: expected '$', got ':'`),
		},
		{
			name:    "One element array with CR array element",
			input:   strings.NewReader("*1\r\n\r\r\n"),
			wantErr: redis.NewSimpleError(`-ERR Protocol error: expected '$', got '\r'`),
		},
		{
			name:    "One element array with LF array element",
			input:   strings.NewReader("*1\r\n\n\r\n"),
			wantErr: redis.NewSimpleError(`-ERR Protocol error: expected '$', got '\n'`),
		},
		{
			name:    "Array with CR length",
			input:   strings.NewReader("*\r"),
			wantErr: redis.NewSimpleError(`-ERR Protocol error: expected digit, got '\r'`),
		},
		{
			name:    "Array with LF length",
			input:   strings.NewReader("*\n"),
			wantErr: redis.NewSimpleError(`-ERR Protocol error: expected digit, got '\n'`),
		},
		{
			name:    "Array with length ending with alphabetical character",
			input:   strings.NewReader("*1a"),
			wantErr: redis.NewSimpleError(`-ERR Protocol error: expected '\r', got 'a'`),
		},
		{
			name:    "One element array with bulk string with longer content than length suggests",
			input:   strings.NewReader("*1\r\n$1\r\nab\r\n"),
			wantErr: redis.NewSimpleError(`-ERR Protocol error: expected '\r', got 'b'`),
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
