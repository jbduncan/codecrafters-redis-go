package await

import "time"

func Until(
	fn func() bool,
	timeout time.Duration,
	sleep time.Duration,
) bool {
	if fn() {
		// Success
		return true
	}

	ticker := time.NewTicker(sleep)
	defer ticker.Stop()
	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-timeoutTimer.C:
			// Timed out
			return false
		case <-ticker.C:
			if fn() {
				// Success
				return true
			}
		}
	}
}
