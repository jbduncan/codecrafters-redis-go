package redis

import "time"

type Clock interface {
	// NowMonotonic returns the current time with a "monotonic time" component. This makes it
	// appropriate for measuring time with Time.After, Time.Before, Time.Compare and Time.Sub.
	NowMonotonic() time.Time
}

type RealClock struct{}

func (r RealClock) NowMonotonic() time.Time {
	return time.Now()
}
