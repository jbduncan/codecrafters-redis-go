package integration_test

import (
	"errors"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

func TestTwoServersWithSamePort(t *testing.T) {
	s, err := redis.StartServer(newTLogWriter(t))
	if err != nil {
		t.Fatalf("server did not start up: %v", err)
	}
	t.Cleanup(s.Stop)

	_, err = redis.StartServer(newTLogWriter(t))
	var portNotBoundError redis.PortNotBoundError
	if !errors.As(err, &portNotBoundError) {
		t.Fatalf("got err %q, want err of type %T", err, portNotBoundError)
	}
	if got, want := portNotBoundError.Port(), defaultPort; got != want {
		t.Fatalf("got port %d, want %d", got, want)
	}
}
