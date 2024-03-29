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

			command, err := redis.NewParser(nil).Parse(reader)

			if err != nil {
				t.Errorf("err: expected: nil; got: %v", err)
			}
			if !reflect.DeepEqual(command, redis.EchoCommand(testCase.echo)) {
				t.Errorf(`command expected to be "%s" but was %#v`, testCase.echo, command)
			}
		})
	}
}

func TestParser_ParseGetRequest(t *testing.T) {
	testCases := []struct {
		name    string
		request string
		key     string
		value   string
	}{
		{
			name:    "GET grape banana",
			request: "*3\r\n$3\r\nGET\r\n$5\r\ngrape\r\n$6\r\nbanana\r\n",
			key:     "grape",
			value:   "banana",
		},
		{
			name:    "get grape banana",
			request: "*3\r\n$3\r\nget\r\n$5\r\ngrape\r\n$6\r\nbanana\r\n",
			key:     "grape",
			value:   "banana",
		},
		{
			name:    "GeT grape banana",
			request: "*3\r\n$3\r\nGeT\r\n$5\r\ngrape\r\n$6\r\nbanana\r\n",
			key:     "grape",
			value:   "banana",
		},
		{
			name:    "GET link zelda",
			request: "*3\r\n$3\r\nGET\r\n$4\r\nlink\r\n$5\r\nzelda\r\n",
			key:     "link",
			value:   "zelda",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			requestReader := strings.NewReader(testCase.request)
			store := map[string]string{}

			command, err := redis.NewParser(store).Parse(requestReader)

			if err != nil {
				t.Errorf("err: expected: nil; got: %v", err)
			}
			want := redis.NewGetCommand(testCase.key, store)
			if !reflect.DeepEqual(command, want) {
				t.Errorf("command expected to be %#v but was %#v", want, command)
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
			requestReader := strings.NewReader(testCase.request)

			command, err := redis.NewParser(nil).Parse(requestReader)

			if err != nil {
				t.Errorf("err: expected: nil; got: %v", err)
			}
			if !reflect.DeepEqual(command, redis.PingCommand) {
				t.Errorf("command expected to be redis.PingCommand but was %v", command)
			}
		})
	}
}

func TestParser_ParseSetRequest(t *testing.T) {
	testCases := []struct {
		name    string
		request string
		key     string
		value   string
	}{
		{
			name:    "SET grape banana",
			request: "*3\r\n$3\r\nSET\r\n$5\r\ngrape\r\n$6\r\nbanana\r\n",
			key:     "grape",
			value:   "banana",
		},
		{
			name:    "set grape banana",
			request: "*3\r\n$3\r\nset\r\n$5\r\ngrape\r\n$6\r\nbanana\r\n",
			key:     "grape",
			value:   "banana",
		},
		{
			name:    "SeT grape banana",
			request: "*3\r\n$3\r\nSeT\r\n$5\r\ngrape\r\n$6\r\nbanana\r\n",
			key:     "grape",
			value:   "banana",
		},
		{
			name:    "SET link zelda",
			request: "*3\r\n$3\r\nSET\r\n$4\r\nlink\r\n$5\r\nzelda\r\n",
			key:     "link",
			value:   "zelda",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			requestReader := strings.NewReader(testCase.request)
			store := map[string]string{}

			command, err := redis.NewParser(store).Parse(requestReader)

			if err != nil {
				t.Errorf("err: expected: nil; got: %v", err)
			}
			want := redis.NewSetCommand(testCase.key, testCase.value, store)
			if !reflect.DeepEqual(command, want) {
				t.Errorf("command expected to be %#v but was %#v", want, command)
			}
		})
	}
}

func TestParser_SetThenGet(t *testing.T) {
	store := map[string]string{}
	setRequestReader := strings.NewReader("*3\r\n$3\r\nSET\r\n$4\r\nlink\r\n$5\r\nzelda\r\n")
	setCommand, _ := redis.NewParser(store).Parse(setRequestReader)
	_ = setCommand.Run()
	getRequestReader := strings.NewReader("*2\r\n$3\r\nGET\r\n$4\r\nlink\r\n")
	getCommand, _ := redis.NewParser(store).Parse(getRequestReader)

	result := getCommand.Run()

	if result != "$3\r\nzelda\r\n" {
		t.Errorf(`expected "zelda" but was %#v`, result)
	}
}
