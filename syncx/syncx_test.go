package syncx_test

import (
	"sync"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/syncx"
)

func TestDoneChan(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	wg.Add(1)

	doneChan := syncx.DoneChan(&wg)

	select {
	case <-doneChan:
		t.Error("DoneChan(): want to be empty before wg.Done()")
	case <-time.After(3 * time.Second):
		// Success
	}

	wg.Done()

	select {
	case <-doneChan:
		// Success
	case <-time.After(3 * time.Second):
		t.Error("DoneChan(): want to be populated on wg.Done()")
	}
}
