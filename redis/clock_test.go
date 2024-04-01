package redis_test

import (
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

func TestRealClock_Now(t *testing.T) {
	t.Parallel()

	r := redis.RealClock{}

	got := r.Now()

	if time.Since(got) > time.Second {
		t.Errorf("Now() = %v, want real time within one second", got)
	}
	if got.Location() != time.UTC {
		t.Errorf("Now() = %v, want real time in UTC", got)
	}
}
