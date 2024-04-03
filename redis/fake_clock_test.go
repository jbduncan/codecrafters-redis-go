package redis_test

import (
	"time"
)

type FakeClock struct {
	// CurrentTime is a fixed [time.Time] that this clock will always return.
	//
	// Note: When CurrentTime is initialized with anything other than time.Now, it will never be
	// monotonic, but for the purpose of testing that's probably fine.
	CurrentTime time.Time
}

// NowMonotonic returns CurrentTime.
//
// Note: When CurrentTime is initialized with anything other than time.Now, it will never be
// monotonic, but for the purpose of testing that's probably fine.
func (c FakeClock) NowMonotonic() time.Time {
	return c.CurrentTime
}
