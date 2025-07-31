package redistest

import (
	"io"
	"net"
	"sync"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/iox"
	"github.com/codecrafters-io/redis-starter-go/syncx"
)

func TestConcurrentPings(t testing.TB, pings int, newConn func() net.Conn) {
	var countDownLatch sync.WaitGroup
	countDownLatch.Add(pings)
	var allSuccessful sync.WaitGroup
	allSuccessful.Add(pings)
	errOccurredCh := make(chan struct{})

	for range pings {
		go func() {
			conn := newConn()

			countDownLatch.Done()
			countDownLatch.Wait()

			if _, err := io.WriteString(conn, RequestPingLowercase); err != nil {
				t.Errorf("request %q not sent: %v", RequestPingLowercase, err)
				errOccurredCh <- struct{}{}
				return
			}

			got, err := iox.ReadExactly(conn, len(ResponsePong))
			if err != nil {
				t.Errorf("response not read: %v", err)
				errOccurredCh <- struct{}{}
			}
			if got != ResponsePong {
				t.Errorf(
					`PING request: got response %q, want %q`,
					got,
					ResponsePong,
				)
				errOccurredCh <- struct{}{}
			}

			allSuccessful.Done()
		}()
	}

	select {
	case <-errOccurredCh:
		// Error has already been reported, so fail the test
		t.FailNow()
	case <-syncx.DoneChan(&allSuccessful):
		// Test has passed
	}
}
