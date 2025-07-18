//go:build integration

package integration_test

import (
	"io"
	"strings"
	"sync"
	"testing"
)

const (
	pingLowercase = "*1\r\n$4\r\nping\r\n"
	pingUppercase = "*1\r\n$4\r\nPING\r\n"
	pong          = "+PONG\r\n"
)

func TestPing(t *testing.T) {

	tests := []struct {
		name     string
		request  string
		response string
	}{
		{
			name:     "PING: lowercase array request",
			request:  pingLowercase,
			response: pong,
		},
		{
			name:     "PING: uppercase array request",
			request:  pingUppercase,
			response: pong,
		},
		{
			name:     "three PINGs: three pipelined simple requests",
			request:  strings.Repeat(pingLowercase, 3),
			response: strings.Repeat(pong, 3),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runServer(t)
			conn := mustDialServer(t)

			if _, err := io.WriteString(conn, tt.request); err != nil {
				t.Fatalf("request %q not sent: %v", tt.request, err)
			}

			got, err := readResponse(conn, len(tt.response))
			if err != nil {
				t.Errorf("response not read: %v", err)
			}
			if want := tt.response; got != want {
				t.Errorf(
					`PING request: got response %q, want %q`, got, want,
				)
			}
		})
	}

	for _, tt := range []struct {
		name            string
		concurrentPings int
	}{
		{
			name:            "two concurrent PINGs",
			concurrentPings: 2,
		},
		{
			name:            "1,000 concurrent PINGs",
			concurrentPings: 1_000,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			runServer(t)

			var countDownLatch sync.WaitGroup
			countDownLatch.Add(tt.concurrentPings)
			var allSuccessful sync.WaitGroup
			allSuccessful.Add(tt.concurrentPings)
			errCh := make(chan struct{})

			for range tt.concurrentPings {
				go func() {
					conn := mustDialServer(t)

					countDownLatch.Done()
					countDownLatch.Wait()

					if _, err := io.WriteString(conn, pingLowercase); err != nil {
						t.Errorf("request %q not sent: %v", pingLowercase, err)
						errCh <- struct{}{}
						return
					}

					got, err := readResponse(conn, len(pong))
					if err != nil {
						t.Errorf("response not read: %v", err)
						errCh <- struct{}{}
					}
					if got != pong {
						t.Errorf(
							`PING request: got response %q, want %q`, got, pong,
						)
						errCh <- struct{}{}
					}

					allSuccessful.Done()
				}()
			}

			select {
			case <-errCh:
				t.FailNow()
			case <-doneChan(&allSuccessful):
				// Test has passed
			}
		})
	}
}
