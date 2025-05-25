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
		{
			name:      "PING: uppercase simple string request",
			request:   "PING\r\n",
			responses: []string{"+PONG\r\n"},
		},
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
		{
			name:      "three PINGs: three pipelined simple requests",
			request:   "PING\r\nPING\r\nPING\r\n",
			responses: []string{"+PONG\r\n", "+PONG\r\n", "+PONG\r\n"},
		},
		{
			name:      "three PINGs: three pipelined requests in array",
			request:   "*3\r\n$4\r\nPING\r\n$4\r\nPING\r\n$4\r\nPING\r\n",
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
}

func loggingClose(t *testing.T, closer io.Closer) {
	err := closer.Close()
	if err != nil {
		t.Log(err)
	}
}
