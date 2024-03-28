package redis_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

func TestParser_ParseEchoRequest(t *testing.T) {
	testCases := []struct {
		name    string
		request string
		echo    string
	}{
		{
			name:    "ECHO hey",
			request: "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n",
			echo:    "hey",
		},
		{
			name:    "echo hey",
			request: "*2\r\n$4\r\necho\r\n$3\r\nhey\r\n",
			echo:    "hey",
		},
		{
			name:    "EcHo hey",
			request: "*2\r\n$4\r\nEcHo\r\n$3\r\nhey\r\n",
			echo:    "hey",
		},
		{
			name:    "ECHO bye",
			request: "*2\r\n$4\r\nECHO\r\n$3\r\nbye\r\n",
			echo:    "bye",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reader := strings.NewReader(testCase.request)

			command, err := redis.Parser{}.Parse(reader)

			if err != nil {
				t.Errorf("err: expected: nil; got: %v", err)
			}
			if !reflect.DeepEqual(command, redis.EchoCommand(testCase.echo)) {
				t.Errorf(`command expected to be "%s" but was %#v`, testCase.echo, command)
			}
		})
	}
}

func TestParser_ParsePingRequest(t *testing.T) {
	testCases := []struct {
		name    string
		request string
	}{
		{
			name:    "PING",
			request: "*1\r\n$4\r\nPING\r\n",
		},
		{
			name:    "ping",
			request: "*1\r\n$4\r\nping\r\n",
		},
		{
			name:    "PiNg",
			request: "*1\r\n$4\r\nPiNg\r\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reader := strings.NewReader(testCase.request)

			command, err := redis.Parser{}.Parse(reader)

			if err != nil {
				t.Errorf("err: expected: nil; got: %v", err)
			}
			if !reflect.DeepEqual(command, redis.PingCommand) {
				t.Errorf("command expected to be redis.PingCommand but was %v", command)
			}
		})
	}
}
