package redis_test

import (
	"errors"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
	"github.com/google/go-cmp/cmp"
)

func testEqual(
	t *testing.T,
	got redis.Value,
	want redis.Value,
	wantErr redis.ErrorValue,
) {
	if wantErr != nil {
		gotErr, ok := got.(error)
		if !ok {
			t.Fatalf("Route(): got %v, want err %q", got, wantErr)
		}

		var gotErrValue redis.ErrorValue
		if !errors.As(gotErr, &gotErrValue) {
			t.Fatalf("Route(): got err %q, want err %q", gotErr, wantErr)
		}

		gotMessage, wantMessage := gotErrValue.Message(), wantErr.Message()
		if gotMessage != wantMessage {
			t.Fatalf("Route(): got err %q, want err %q", gotErr, wantErr)
		}
		return
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Route() mismatch (-want +got):\n%s", diff)
	}
}
