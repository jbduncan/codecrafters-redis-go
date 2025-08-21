package integration_test

import (
	"context"
	"errors"
	"syscall"
	"testing"
)

func TestServerStop(t *testing.T) {
	s := runServer(t)

	s.Stop()

	_, err := dialServer()
	var expectedErr error
	if !errors.Is(err, syscall.ECONNREFUSED) {
		t.Fatalf("got sub-error %q, want %q", expectedErr, syscall.ECONNREFUSED)
	}
}

func TestServerStopViaContext(t *testing.T) {
	// TODO: make this test pass
	t.Skip("to be implemented later")

	ctx, cancel := context.WithCancel(context.Background())
	runServerContext(ctx, t)

	cancel()

	_, err := dialServer()
	var expectedErr error
	if !errors.Is(err, syscall.ECONNREFUSED) {
		t.Fatalf("got sub-error %q, want %q", expectedErr, syscall.ECONNREFUSED)
	}
}
