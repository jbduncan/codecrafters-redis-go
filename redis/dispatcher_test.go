package redis_test

import (
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/redis"
	"github.com/codecrafters-io/redis-starter-go/repeat"
)

type conn struct {
	tcpConn redis.TCPConn
	err     error
}

type mockTCPConn struct {
	readCalled   bool
	readCalledMu *sync.Mutex
}

func newMockTCPConn() *mockTCPConn {
	return &mockTCPConn{
		readCalled:   false,
		readCalledMu: new(sync.Mutex),
	}
}

func (c *mockTCPConn) Read(_ []byte) (n int, err error) {
	c.readCalledMu.Lock()
	defer c.readCalledMu.Unlock()
	c.readCalled = true

	return 0, nil
}

func (c *mockTCPConn) Write(_ []byte) (n int, err error) {
	return 0, nil
}

func (c *mockTCPConn) Close() error {
	return nil
}

func (c *mockTCPConn) ReadCalled() bool {
	c.readCalledMu.Lock()
	defer c.readCalledMu.Unlock()
	return c.readCalled
}

func netPipe() (net.Conn, net.Conn) {
	a, b := net.Pipe()
	deadline := time.Now().Add(10 * time.Second)
	_ = a.SetDeadline(deadline)
	_ = b.SetDeadline(deadline)
	return a, b
}

const (
	pingLowercase = "*1\r\n$4\r\nping\r\n"
	pingUppercase = "*1\r\n$4\r\nPING\r\n"
	pong          = "+PONG\r\n"
)

func TestDispatcher_Run(t *testing.T) {
	t.Parallel()

	// TODO: wire everything together in Dispatcher

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
			t.Parallel()

			serverConns := make(chan *conn, 2)
			clientConn, serverConn := netPipe()
			serverConns <- &conn{
				tcpConn: serverConn,
			}
			serverConns <- &conn{
				err: errors.New("no new connections left"),
			}

			dispatcher := redis.NewDispatcher(
				func() (redis.TCPConn, error) {
					// This will eventually block to stop Dispatcher's inner loop
					// from looping forever.
					c := <-serverConns
					return c.tcpConn, c.err
				},
				func() {},
			)
			go dispatcher.Run()
			t.Cleanup(dispatcher.Stop)
			if _, err := io.WriteString(clientConn, tt.request); err != nil {
				t.Fatalf("request %q not sent: %v", tt.request, err)
			}

			got, err := readResponse(clientConn, len(tt.response))
			if err != nil {
				t.Fatalf("response not read: %v", err)
			}
			if want := tt.response; got != want {
				t.Errorf(
					`Dispatcher.Run(): PING request: got response %q, want %q`,
					got,
					want,
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
			t.Parallel()

			serverConns := make(chan *conn, tt.concurrentPings)
			clientConns := make(chan net.Conn, tt.concurrentPings)
			for range tt.concurrentPings {
				clientConn, serverConn := netPipe()
				serverConns <- &conn{
					tcpConn: serverConn,
				}
				clientConns <- clientConn
			}

			dispatcher := redis.NewDispatcher(
				func() (redis.TCPConn, error) {
					// This will eventually block to stop Dispatcher's inner loop
					// from looping forever.
					c := <-serverConns
					return c.tcpConn, c.err
				},
				func() {},
			)
			go dispatcher.Run()
			t.Cleanup(dispatcher.Stop)

			var countDownLatch sync.WaitGroup
			countDownLatch.Add(tt.concurrentPings)
			var allSuccessful sync.WaitGroup
			allSuccessful.Add(tt.concurrentPings)
			errCh := make(chan struct{})

			for range tt.concurrentPings {
				go func() {
					conn := <-clientConns

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

	t.Run(
		"edge case: when acceptTCPConn returns error, then conn is not read",
		func(t *testing.T) {
			t.Parallel()

			mockConn := newMockTCPConn()
			serverConns := make(chan conn, 1)
			serverConns <- conn{
				tcpConn: mockConn,
				err:     errors.New("no new connections left"),
			}

			dispatcher := redis.NewDispatcher(
				func() (redis.TCPConn, error) {
					// This will eventually block to stop Dispatcher's inner
					// loop from looping forever.
					c := <-serverConns
					return c.tcpConn, c.err
				},
				func() {},
			)
			// This will panic if the nil connection is read
			go dispatcher.Run()
			t.Cleanup(dispatcher.Stop)

			if !repeat.Whilst(
				func() bool {
					return !mockConn.ReadCalled()
				},
				3*time.Second,
				100*time.Millisecond,
			) {
				t.Fatalf(
					"Dispatcher.Run(): expected not to call conn.Read() " +
						"but it did",
				)
			}
		})
}

func doneChan(wg *sync.WaitGroup) chan struct{} {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()
	return done
}

func readResponse(conn net.Conn, responseLength int) (string, error) {
	gotBytes := make([]byte, responseLength)
	if _, err := io.ReadFull(conn, gotBytes); err != nil {
		return "", err
	}
	got := string(gotBytes)
	return got, nil
}
