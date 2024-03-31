package redis_test

import "time"

type FakeClock struct {
	CurrentTime time.Time
}

func (c FakeClock) Now() time.Time {
	return c.CurrentTime
}
