package redis_test

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
	"github.com/google/go-cmp/cmp"
)

func TestSimpleError_Message(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  redis.ErrorValue
		want string
	}{
		{
			name: "ERR foo",
			err:  redis.MakeSimpleError("ERR foo"),
			want: "ERR foo",
		},
		{
			name: "ERR bar",
			err:  redis.MakeSimpleError("ERR bar"),
			want: "ERR bar",
		},
		{
			name: "ERR foo",
			err:  redis.MakeBulkError("ERR foo"),
			want: "ERR foo",
		},
		{
			name: "ERR bar",
			err:  redis.MakeBulkError("ERR bar"),
			want: "ERR bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.err.Message()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
