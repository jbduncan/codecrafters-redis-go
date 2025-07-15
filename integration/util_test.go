//go:build integration

package integration_test

import (
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/await"
	"github.com/codecrafters-io/redis-starter-go/redis"
)

const defaultPort = 6379

func runServer(t *testing.T) {
	server := redis.Server{}
	go server.Run()
	t.Cleanup(server.Stop)

	if err := awaitServerStartup(); err != nil {
		t.Error(err.Error())
	}
}

func awaitServerStartup() error {
	timeout := 5 * time.Second
	sleep := 50 * time.Millisecond
	if !await.Until(serverIsUp, timeout, sleep) {
		return fmt.Errorf("server did not start up in %s", timeout)
	}
	return nil
}

func serverIsUp() bool {
	_, err := dialServer()
	return err == nil
}

func dialServer() (net.Conn, error) {
	result, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", defaultPort))
	if err != nil {
		return nil, err
	}

	if err := result.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		return nil, err
	}

	return result, nil
}

func mustDialServer(t *testing.T) net.Conn {
	conn, err := dialServer()
	if err != nil {
		t.Fatalf("no connection to server: %v", err)
	}
	t.Cleanup(func() {
		loggingClose(t, conn)
	})
	return conn
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

func loggingClose(t *testing.T, closer io.Closer) {
	err := closer.Close()
	if err != nil {
		t.Log(err)
	}
}
