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
		t.Errorf(`command expected to return "+PONG\r\n" but was %#v`, response)
	}
}

func TestSetCommand(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "(grape: banana)",
			key:   "grape",
			value: "banana",
		},
		{
			name:  "(link: zelda)",
			key:   "link",
			value: "zelda",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := map[string]string{}

			response := redis.NewSetCommand(testCase.key, testCase.value, store).Run()

			if response != "+OK\r\n" {
				t.Errorf(`command expected to return "+OK\r\n" but was %#v`, response)
			}
			if store[testCase.key] != testCase.value {
				t.Errorf(
					`command expected to contain key-value pair (%s: %s) but was %#v`,
					testCase.key, testCase.value, store,
				)
			}
		})
	}
}
