//go:build integration

package integration_test

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/await"
	"github.com/codecrafters-io/redis-starter-go/redis"
)

const defaultPort = 6379

func runServer(t *testing.T) {
	server := redis.Server{}
	go server.Run()
	t.Cleanup(func() {
		if err := awaitServerTearDown(); err != nil {
			t.Errorf("server was not torn down: %v", err)
		}
	})
	t.Cleanup(server.Stop)

	if err := awaitServerStartUp(); err != nil {
		t.Error(err.Error())
	}
}

func awaitServerStartUp() error {
	timeout := 5 * time.Second
	sleep := 50 * time.Millisecond
	if !await.Until(serverIsUp, timeout, sleep) {
		return fmt.Errorf("server did not start up in %s", timeout)
	}
	return nil
}

func awaitServerTearDown() error {
	timeout := 5 * time.Second
	sleep := 50 * time.Millisecond
	if !await.Until(serverIsDown, timeout, sleep) {
		return fmt.Errorf("server did not start up in %s", timeout)
	}
	return nil
}

func serverIsUp() bool {
	_, err := dialServer()
	return err == nil
}

func serverIsDown() bool {
	return !serverIsUp()
}

func dialServer() (net.Conn, error) {
	result, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", defaultPort))
	if err != nil {
		return nil, err
	}

	if err := result.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
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

func loggingClose(t *testing.T, closer io.Closer) {
	if err := closer.Close(); err != nil {
		t.Log(err)
	}
}
