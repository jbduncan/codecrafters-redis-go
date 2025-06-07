//go:build integration

package integration_test

import (
	"bufio"
	"io"
	"testing"
)

func TestPing(t *testing.T) {
	tests := []struct {
		name      string
		request   string
		responses []string
	}{
		// TODO: remove test as it's not a valid request:
		//       https://redis.io/docs/latest/develop/reference/protocol-spec/#sending-commands-to-a-redis-server
		{
			name:      "PING: uppercase simple string request",
			request:   "PING\r\n",
			responses: []string{"+PONG\r\n"},
		},
		// TODO: change to an array request:
		//       https://redis.io/docs/latest/develop/reference/protocol-spec/#sending-commands-to-a-redis-server
		{
			name:      "PING: lowercase simple string request",
			request:   "ping\r\n",
			responses: []string{"+PONG\r\n"},
		},
		{
			name:      "PING: array request",
			request:   "*1\r\n$4\r\nPING\r\n",
			responses: []string{"+PONG\r\n"},
		},
		// TODO: change to a pipeline of array requests:
		//       https://redis.io/docs/latest/develop/reference/protocol-spec/#sending-commands-to-a-redis-server
		{
			name:      "three PINGs: three pipelined simple requests",
			request:   "PING\r\nPING\r\nPING\r\n",
			responses: []string{"+PONG\r\n", "+PONG\r\n", "+PONG\r\n"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stop := runServer(t)
			defer stop()
			conn, err := dialServer()
			if err != nil {
				t.Fatalf("no connection to server: %v", err)
			}
			defer loggingClose(t, conn)
			connReader := bufio.NewReader(conn)

			if _, err = io.WriteString(conn, tt.request); err != nil {
				t.Fatalf("did not write message %q successfully: %v", tt.request, err)
			}

			for _, want := range tt.responses {
				got, err := connReader.ReadString('\n')
				if err != nil {
					t.Errorf("did not read conn successfully: %v", err)
				}
				if got != want {
					t.Errorf(
						`PING request: got response %q, want %q`, got, want,
					)
				}
			}
		})
	}

	t.Run("two PINGS: two concurrent requests", func(t *testing.T) {
		stop := runServer(t)
		defer stop()
		conn1, err := dialServer()
		if err != nil {
			t.Fatalf("no connection to server: %v", err)
		}
		defer loggingClose(t, conn1)
		conn2, err := dialServer()
		if err != nil {
			t.Fatalf("no connection to server: %v", err)
		}
		defer loggingClose(t, conn2)
		conn1Reader := bufio.NewReader(conn1)
		conn2Reader := bufio.NewReader(conn2)

		if _, err = io.WriteString(conn1, "PING\r\n"); err != nil {
			t.Fatalf(`did not write message "PING\r\n" successfully: %v`, err)
		}
		if _, err = io.WriteString(conn2, "PING\r\n"); err != nil {
			t.Fatalf(`did not write message "PING\r\n" successfully: %v`, err)
		}

		got, err := conn1Reader.ReadString('\n')
		if err != nil {
			t.Errorf("did not read conn1Reader successfully: %v", err)
		}
		if got != "+PONG\r\n" {
			t.Errorf(
				`PING request: got response %q, want "+PONG\r\n"`, got,
			)
		}

		got, err = conn2Reader.ReadString('\n')
		if err != nil {
			t.Errorf("did not read conn2Reader successfully: %v", err)
		}
		if got != "+PONG\r\n" {
			t.Errorf(
				`PING request: got response %q, want "+PONG\r\n"`, got,
			)
		}
	})
}

func loggingClose(t *testing.T, closer io.Closer) {
	err := closer.Close()
	if err != nil {
		t.Log(err)
	}
}
