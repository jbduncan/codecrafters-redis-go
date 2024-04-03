package redis_test

import (
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

func TestRealClock_NowMonotonic(t *testing.T) {
	t.Parallel()

	r := redis.RealClock{}

	got := r.NowMonotonic()

	if isNotMonotonicTime(got) {
		t.Errorf("NowMonotonic() = %v, want monotonic time", got)
	}
	if time.Since(got) > time.Second {
		t.Errorf("NowMonotonic() = %v, want real time within one second", got)
	}
}

func isNotMonotonicTime(t time.Time) bool {
	// t.Round(0) returns a new time.Time with any monotonic time component stripped off.
	//
	// `==` returns true if the two times have the same monotonic time value, otherwise false.
	//
	// Therefore, if t == t.Round(0) is true, then t's monotonic time component must have been
	// stripped off.
	return t == t.Round(0)
}
