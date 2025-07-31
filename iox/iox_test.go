package iox_test

import (
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/codecrafters-io/redis-starter-go/iox"
)

func TestReadExactly(t *testing.T) {
	t.Parallel()

	type args struct {
		r        io.Reader
		numBytes int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no input",
			args: args{
				r:        strings.NewReader(""),
				numBytes: 0,
			},
			want: "",
		},
		{
			name: "one-byte input",
			args: args{
				r:        strings.NewReader("a"),
				numBytes: 1,
			},
			want: "a",
		},
		{
			name: "one byte from two-byte input",
			args: args{
				r:        strings.NewReader("ab"),
				numBytes: 1,
			},
			want: "a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, _ := iox.ReadExactly(tt.args.r, tt.args.numBytes)

			if got != tt.want {
				t.Errorf("ReadExactly(): got %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("error", func(t *testing.T) {
		wantErr := errors.New("ganondorf stole the triforce")

		_, err := iox.ReadExactly(iotest.ErrReader(wantErr), 1)

		if !errors.Is(err, wantErr) {
			t.Errorf(
				"ReadExactly(): got error %q, want %q",
				err,
				wantErr,
			)
		}
	})
}
