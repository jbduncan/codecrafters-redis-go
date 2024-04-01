package redis_test

import (
	"github.com/codecrafters-io/redis-starter-go/redis"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParser_ParseEchoRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.request)
			store := redis.NewStore()
			clock := FakeClock{}

			command, err := redis.NewParser(store, clock).Parse(reader)

			if err != nil {
				t.Errorf("err: expected: nil; got: %v", err)
			}
			if !reflect.DeepEqual(command, redis.EchoCommand(tt.echo)) {
				t.Errorf(`command expected to be "%s" but was %#v`, tt.echo, command)
			}
		})
	}
}

func TestParser_ParseGetRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestReader := strings.NewReader(tt.request)
			store := redis.NewStore()
			clock := FakeClock{}

			command, err := redis.NewParser(store, clock).Parse(requestReader)

			if err != nil {
				t.Errorf("err: expected: nil; got: %v", err)
			}
			want := redis.NewGetCommand(store, clock, tt.key)
			if !reflect.DeepEqual(command, want) {
				t.Errorf("command expected to be %#v but was %#v", want, command)
			}
		})
	}
}

func TestParser_ParseInfoRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		request  string
		infoKind string
	}{
		{
			name:     "INFO replication",
			request:  "*2\r\n$4\r\nINFO\r\n$11\r\nreplication\r\n",
			infoKind: "replication",
		},
		{
			name:     "info replication",
			request:  "*2\r\n$4\r\ninfo\r\n$11\r\nreplication\r\n",
			infoKind: "replication",
		},
		{
			name:     "InFo replication",
			request:  "*2\r\n$4\r\nInFo\r\n$11\r\nreplication\r\n",
			infoKind: "replication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestReader := strings.NewReader(tt.request)
			store := redis.NewStore()
			clock := FakeClock{}

			command, err := redis.NewParser(store, clock).Parse(requestReader)

			if err != nil {
				t.Errorf("err: expected: nil; got: %v", err)
			}
			want := redis.InfoCommand(tt.infoKind)
			if command != want {
				t.Errorf("command expected to be %#v but was %#v", want, command)
			}
		})
	}
}

func TestParser_ParsePingRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestReader := strings.NewReader(tt.request)
			store := redis.NewStore()
			clock := FakeClock{}

			command, err := redis.NewParser(store, clock).Parse(requestReader)

			if err != nil {
				t.Errorf("err: expected: nil; got: %v", err)
			}
			if !reflect.DeepEqual(command, redis.PingCommand{}) {
				t.Errorf("command expected to be redis.PingCommand but was %v", command)
			}
		})
	}
}

func TestParser_ParseSetRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		request string
		key     string
		value   string
		ttl     *time.Duration
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
		{
			name:    "SET grape banana px 100",
			request: "*5\r\n$3\r\nSET\r\n$5\r\ngrape\r\n$6\r\nbanana\r\n$2\r\npx\r\n$3\r\n100\r\n",
			key:     "grape",
			value:   "banana",
			ttl:     ptr(100 * time.Millisecond),
		},
		{
			name:    "SET grape banana PX 200",
			request: "*5\r\n$3\r\nSET\r\n$5\r\ngrape\r\n$6\r\nbanana\r\n$2\r\nPX\r\n$3\r\n200\r\n",
			key:     "grape",
			value:   "banana",
			ttl:     ptr(200 * time.Millisecond),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := redis.NewStore()
			clock := FakeClock{CurrentTime: time.UnixMilli(0).UTC()}
			requestReader := strings.NewReader(tt.request)

			command, err := redis.NewParser(store, clock).Parse(requestReader)

			if err != nil {
				t.Errorf("err: expected: nil; got: %v", err)
			}
			setCommand, ok := command.(*redis.SetCommand)
			if !ok {
				t.Errorf(`ok expected to be true but was false`)
			}
			want := redis.NewSetCommand(store, tt.key, tt.value)
			if tt.ttl != nil {
				want = redis.NewSetCommand(
					store,
					tt.key,
					tt.value,
					redis.ExpiryTime(clock.Now().Add(*tt.ttl)),
				)
			}
			if !setCommand.Equal(want) {
				t.Errorf("command expected to be equal to %#v but was %#v", want, command)
			}
		})
	}
}
