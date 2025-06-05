package repeat

import "time"

func Whilst(
	fn func() bool,
	timeout time.Duration,
	sleep time.Duration,
) bool {
	if !fn() {
		// Failure
		return false
	}

	ticker := time.NewTicker(sleep)
	defer ticker.Stop()
	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-timeoutTimer.C:
			// Timed out; all successes!
			return true
		case <-ticker.C:
			if !fn() {
				// Failure
				return false
			}
		}
	}
}
