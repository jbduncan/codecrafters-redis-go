package redis_test

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

func TestEchoCommand(t *testing.T) {
	testCases := []struct {
		echo     string
		response string
	}{
		{
			echo:     "hey",
			response: "$3\r\nhey\r\n",
		},
		{
			echo:     "goodbye",
			response: "$7\r\ngoodbye\r\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.echo, func(t *testing.T) {
			response := redis.EchoCommand(testCase.echo).Run()
			if response != testCase.response {
				t.Errorf(`command expected to return %#v but was %#v`, testCase.response, response)
			}
		})
	}
}

func TestPingCommand(t *testing.T) {
	response := redis.PingCommand.Run()
	if response != "+PONG\r\n" {
		t.Errorf(`command expected to return "+PONG\r\n" but was %s`, response)
	}
}
