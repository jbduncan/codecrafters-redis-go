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
		name  string
		input io.Reader
		want  redis.Value
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
			name:  "Missing CRLF",
			input: strings.NewReader("PING"),
			want:  redis.NewSimpleError("ERR missing CRLF"),
		},
		{
			name:  "Missing CR",
			input: strings.NewReader("PING\n"),
			want:  redis.NewSimpleError("ERR missing CRLF"),
		},
		{
			name:  "Missing LF",
			input: strings.NewReader("PING\r"),
			want:  redis.NewSimpleError("ERR missing CRLF"),
		},
		{
			name:  "No input",
			input: strings.NewReader(""),
			want:  redis.NewSimpleError("ERR missing CRLF"),
		},
		{
			name:  "Error-returning input",
			input: badReader{},
			want:  redis.NewSimpleError("ERR internal scanner error"),
		},
		{
			name:  `Input that returns error on LF`,
			input: io.MultiReader(strings.NewReader("PING\r"), badReader{}),
			want:  redis.NewSimpleError("ERR internal scanner error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := redis.NewRESP2Scanner(tt.input)
			got := s.Scan()

			if diff := cmp.Diff(
				tt.want,
				got,
				equateSimpleErrors(),
			); diff != "" {
				t.Errorf("Scan() mismatch (-want +got):\n%s", diff)
			}
		})
	}

	t.Run("Pipelined simple strings", func(t *testing.T) {
		input := "PING\r\nPING\r\n"
		want := redis.SimpleString("PING")
		s := redis.NewRESP2Scanner(strings.NewReader(input))
		for range 2 {
			got := s.Scan()
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Scan() mismatch (-want +got):\n%s", diff)
			}
		}
	})
}

func equateSimpleErrors() cmp.Option {
	return cmp.FilterValues(areSimpleErrors, cmp.Comparer(compareSimpleErrors))
}

func areSimpleErrors(a, b any) bool {
	errA, eaOk := a.(error)
	errB, ebOk := b.(error)
	if !eaOk || !ebOk {
		return false
	}

	var simpleError redis.SimpleError
	return errors.As(errA, &simpleError) &&
		errors.As(errB, &simpleError)
}

func compareSimpleErrors(a, b any) bool {
	var simpleErrorA redis.SimpleError
	var simpleErrorB redis.SimpleError
	errors.As(a.(error), &simpleErrorA)
	errors.As(b.(error), &simpleErrorB)

	return simpleErrorA.Message() == simpleErrorB.Message()
}
