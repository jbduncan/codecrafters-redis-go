package integration_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/await"
	"github.com/codecrafters-io/redis-starter-go/redis"
)

const (
	defaultPort = 6379
)

func runServer(t *testing.T) *redis.Server {
	t.Helper()

	server, err := redis.StartServer(context.Background(), newTLogWriter(t))

	if err != nil {
		t.Fatalf("server did not start up: %v", err)
	}
	t.Cleanup(server.Stop)

	if err := awaitServerStartUp(); err != nil {
		t.Error(err.Error())
	}

	return server
}

func runServerContext(ctx context.Context, t *testing.T) {
	if _, err := redis.StartServer(ctx, newTLogWriter(t)); err != nil {
		t.Fatalf("server did not start up: %v", err)
	}

	if err := awaitServerStartUp(); err != nil {
		t.Error(err.Error())
	}
}

type tLogWriter struct {
	t *testing.T
}

func newTLogWriter(t *testing.T) *tLogWriter {
	return &tLogWriter{t: t}
}

func (w *tLogWriter) Write(p []byte) (n int, err error) {
	w.t.Log(string(p))
	return len(p), nil
}

func awaitServerStartUp() error {
	timeout := 5 * time.Second
	sleep := 50 * time.Millisecond
	if !await.Until(serverIsUp, timeout, sleep) {
		return fmt.Errorf("server did not start up in %s", timeout)
	}
	return nil
}

func serverIsUp() bool {
	conn, err := dialServer()
	if err != nil {
		return false
	}
	defer func() {
		_ = conn.Close()
	}()
	return true
}

func mustDialServer(t *testing.T) net.Conn {
	t.Helper()

	conn, err := dialServer()
	if err != nil {
		t.Fatalf("no connection to server: %v", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Logf("connection to server could not be closed: %v", err)
		}
	})
	return conn
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
