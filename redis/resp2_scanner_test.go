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
			name:    "Missing CRLF",
			input:   strings.NewReader("PING"),
			wantErr: redis.NewSimpleError("SYNTAX missing CRLF"),
		},
		{
			name:    "Missing CR",
			input:   strings.NewReader("PING\n"),
			wantErr: redis.NewSimpleError("SYNTAX missing CRLF"),
		},
		{
			name:    "Missing LF",
			input:   strings.NewReader("PING\r"),
			wantErr: redis.NewSimpleError("SYNTAX missing CRLF"),
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
		{
			name:           `Input that returns error on LF`,
			input:          io.MultiReader(strings.NewReader("PING\r"), badReader{}),
			wantGenericErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := redis.NewRESP2Scanner(tt.input)
			got, err := s.Scan()

			if tt.wantGenericErr {
				if err == nil {
					t.Fatalf("Scan(): got a <nil> error, want a non-nil error")
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

			_, secondScanErr := s.Scan()

			if !errors.Is(secondScanErr, io.EOF) {
				t.Fatalf(
					"Scan(): nothing else should have been read: "+
						"got %q, want io.EOF",
					err)
			}
		})
	}

	t.Run("Pipelined inputs", func(t *testing.T) {
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
