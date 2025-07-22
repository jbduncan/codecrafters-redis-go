package redis_test

import (
	"errors"
	"io"
	"iter"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/codecrafters-io/redis-starter-go/redis"
	"github.com/google/go-cmp/cmp"
)

var (
	badError  = errors.New("ganondorf stole the triforce")
	badReader = iotest.ErrReader(badError)
)

func iterSeq2Pull(
	t *testing.T,
	values iter.Seq2[redis.InputValue, error],
) func() (redis.InputValue, error, bool) {
	next, stop := iter.Pull2(values)
	t.Cleanup(stop)
	return next
}

func TestRESP2Scanner_ScanAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          io.Reader
		want           redis.InputValue
		wantErr        error
		wantGenericErr error
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
			name:    "Missing CRLF for simple string",
			input:   strings.NewReader("PING"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: incomplete request`),
		},
		{
			name:    "Missing CR for simple string",
			input:   strings.NewReader("PING\n"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: incomplete request`),
		},
		{
			name:    "Missing LF for simple string",
			input:   strings.NewReader("PING\r"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: incomplete request`),
		},
		{
			name:           "Simple string that returns error on LF",
			input:          io.MultiReader(strings.NewReader("PING\r"), badReader),
			wantGenericErr: badError,
		},
		{
			name:  "Empty array",
			input: strings.NewReader("*0\r\n"),
			want:  redis.BulkStringArray{},
		},
		{
			name:  "One element array with uppercase bulk string",
			input: strings.NewReader("*1\r\n$4\r\nPING\r\n"),
			want: redis.BulkStringArray{
				"PING",
			},
		},
		{
			name:  "One element array with lowercase bulk string",
			input: strings.NewReader("*1\r\n$4\r\nping\r\n"),
			want: redis.BulkStringArray{
				"ping",
			},
		},
		{
			name:  "One element array with some other bulk string",
			input: strings.NewReader("*1\r\n$6\r\nFoobar\r\n"),
			want: redis.BulkStringArray{
				"Foobar",
			},
		},
		{
			name:  "One element array with longer bulk string",
			input: strings.NewReader("*1\r\n$10\r\nenumerable\r\n"),
			want: redis.BulkStringArray{
				"enumerable",
			},
		},
		{
			name:  "One element array with bulk string with CR",
			input: strings.NewReader("*1\r\n$1\r\n\r\r\n"),
			want: redis.BulkStringArray{
				"\r",
			},
		},
		{
			name:  "One element array with bulk string with LF",
			input: strings.NewReader("*1\r\n$1\r\n\n\r\n"),
			want: redis.BulkStringArray{
				"\n",
			},
		},
		{
			name:    "Zero element array with missing CR",
			input:   strings.NewReader("*0\n"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: expected '\r', got '\n'`),
		},
		{
			name:    "Invalid array with missing LF",
			input:   strings.NewReader("*0\r"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: incomplete request`),
		},
		{
			name:    "Invalid array with two CRs",
			input:   strings.NewReader("*0\r\r"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: expected '\n', got '\r'`),
		},
		{
			name:    "One element array with missing element",
			input:   strings.NewReader("*1\r\n"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: incomplete request`),
		},
		{
			name:    "One element array with invalid bulk string with just the starting $",
			input:   strings.NewReader("*1\r\n$"),
			wantErr: redis.MakeBulkError("ERR Protocol error: incomplete request"),
		},
		{
			name:    "One element array with invalid bulk string without length",
			input:   strings.NewReader("*1\r\n$foo\r\nbar\r\n"),
			wantErr: redis.MakeBulkError("ERR Protocol error: expected digit, got 'f'"),
		},
		{
			name:    "One element array with invalid bulk string with missing first CR",
			input:   strings.NewReader("*1\r\n$3\nfoo\r\n"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: expected '\r', got '\n'`),
		},
		{
			name:    "One element array with invalid bulk string with missing first LF",
			input:   strings.NewReader("*1\r\n$3\rfoo\r\n"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: expected '\n', got 'f'`),
		},
		{
			name:    "One element array with invalid bulk string with missing second CR",
			input:   strings.NewReader("*1\r\n$3\r\nfoo\n"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: expected '\r', got '\n'`),
		},
		{
			name:    "One element array with invalid bulk string with missing second LF",
			input:   strings.NewReader("*1\r\n$3\r\nfoo\r"),
			wantErr: redis.MakeBulkError("ERR Protocol error: incomplete request"),
		},
		{
			name:    "One element array with bulk string with shorter content than length suggests",
			input:   strings.NewReader("*1\r\n$2\r\na\r\n"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: expected '\r', got '\n'`),
		},
		{
			name:    "One element array with bulk string with longer content than length suggests",
			input:   strings.NewReader("*1\r\n$1\r\nab\r\n"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: expected '\r', got 'b'`),
		},
		{
			name:    "One element array with bulk string with less bytes in total than length suggests",
			input:   strings.NewReader("*1\r\n$6\r\na\r\n"),
			wantErr: redis.MakeBulkError("ERR Protocol error: incomplete request"),
		},
		{
			name:    "One element array with invalid bulk string with just a length",
			input:   strings.NewReader("*1\r\n$1"),
			wantErr: redis.MakeBulkError("ERR Protocol error: incomplete request"),
		},
		{
			name:    "One element array with invalid bulk string with just length and first CRLF",
			input:   strings.NewReader("*1\r\n$1\r\n"),
			wantErr: redis.MakeBulkError("ERR Protocol error: incomplete request"),
		},
		{
			name:           "One element array with bulk string that returns error on first digit",
			input:          io.MultiReader(strings.NewReader("*1\r\n$"), badReader),
			wantGenericErr: badError,
		},
		{
			name:           "One element array with bulk string that returns error on first CR",
			input:          io.MultiReader(strings.NewReader("*1\r\n$1"), badReader),
			wantGenericErr: badError,
		},
		{
			name:           "One element array with bulk string that returns error on first payload byte",
			input:          io.MultiReader(strings.NewReader("*1\r\n$1\r\n"), badReader),
			wantGenericErr: badError,
		},
		{
			name:    "One element array with bulk string of negative length",
			input:   strings.NewReader("*1\r\n$-1\r\n"),
			wantErr: redis.MakeBulkError("ERR Protocol error: expected digit, got '-'"),
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
			want: redis.BulkStringArray{
				"A",
				"B",
				"C",
				"D",
				"E",
				"F",
				"G",
				"H",
				"I",
				"J",
			},
		},
		{
			name:    "One element array with inner array",
			input:   strings.NewReader("*1\r\n*0\r\n"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: expected '$', got '*'`),
		},
		{
			name:    "One element array with negative length",
			input:   strings.NewReader("*-1\r\n$1\r\na\r\n"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: expected digit, got '-'`),
		},
		{
			name:    "One element array with integer array element",
			input:   strings.NewReader("*1\r\n:1\r\n"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: expected '$', got ':'`),
		},
		{
			name:    "One element array with CR array element",
			input:   strings.NewReader("*1\r\n\r\r\n"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: expected '$', got '\r'`),
		},
		{
			name:    "One element array with LF array element",
			input:   strings.NewReader("*1\r\n\n\r\n"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: expected '$', got '\n'`),
		},
		{
			name:    "Array with CR length",
			input:   strings.NewReader("*\r"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: expected digit, got '\r'`),
		},
		{
			name:    "Array with LF length",
			input:   strings.NewReader("*\n"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: expected digit, got '\n'`),
		},
		{
			name:    "Array with length ending with alphabetical character",
			input:   strings.NewReader("*1a"),
			wantErr: redis.MakeBulkError(`ERR Protocol error: expected '\r', got 'a'`),
		},
		{
			name:    "Array with length exceeding maximum signed 64-bit integer",
			input:   strings.NewReader("*9223372036854775808"),
			wantErr: redis.MakeBulkError("ERR Protocol error: invalid multibulk length"),
		},
		{
			name:           "Error-returning input",
			input:          badReader,
			wantGenericErr: badError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := redis.NewRESP2Scanner(tt.input).ScanAll()
			next := iterSeq2Pull(t, s)

			got, err, ok := next()

			if !ok {
				t.Fatalf("ScanAll(): got zero elements, want one element")
			}
			if tt.wantGenericErr != nil {
				if !errors.Is(err, tt.wantGenericErr) {
					t.Fatalf(
						"ScanAll(): got generic err %q, want %q",
						err,
						tt.wantGenericErr,
					)
				}
				return
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ScanAll(): got err %q, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ScanAll(): got err %q, want <nil> err", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ScanAll() mismatch (-want +got):\n%s", diff)
			}

			got2, err2, ok2 := next()

			if err2 != nil {
				t.Fatalf("ScanAll(): got err %q, want <nil> err", err2)
			}
			if ok2 {
				t.Fatalf(
					"ScanAll(): nothing else should have been read; "+
						"got value %q",
					got2,
				)
			}
		})
	}

	t.Run("No input", func(t *testing.T) {
		t.Parallel()

		noInput := strings.NewReader("")
		s := redis.NewRESP2Scanner(noInput)
		next := iterSeq2Pull(t, s.ScanAll())

		got, err, ok := next()
		if err != nil {
			t.Fatalf("ScanAll(): got err %q, want <nil> err", err)
		}
		if ok {
			t.Fatalf("ScanAll(): got element %q, want 0 elements", got)
		}
	})

	t.Run("Pipelined inputs", func(t *testing.T) {
		t.Parallel()

		input := "PING\r\nPING\r\n"
		want := redis.SimpleString("PING")

		s := redis.NewRESP2Scanner(strings.NewReader(input))
		next := iterSeq2Pull(t, s.ScanAll())

		for i := range 2 {
			got, err, ok := next()
			if err != nil {
				t.Fatalf("ScanAll(): got err %q, want <nil> err", err)
			}
			if !ok {
				t.Fatalf("ScanAll(): got %d element(s), want 2 elements", i)
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("ScanAll() mismatch (-want +got):\n%s", diff)
			}
		}

		got, err, ok := next()
		if err != nil {
			t.Fatalf("ScanAll(): got err %q, want <nil> err", err)
		}
		if ok {
			t.Fatalf(
				"ScanAll(): nothing else should have been read; "+
					"got value %q",
				got,
			)
		}
	})
}
