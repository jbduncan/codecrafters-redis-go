package redis_test

import (
	"github.com/codecrafters-io/redis-starter-go/redis"
	"testing"
	"time"
)

const redisNullBulkString = "$-1\r\n"

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

func TestGetCommand(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		value    string
		response string
	}{
		{
			name:     "(grape: banana)",
			key:      "grape",
			value:    "banana",
			response: "$6\r\nbanana\r\n",
		},
		{
			name:     "(link: zelda)",
			key:      "link",
			value:    "zelda",
			response: "$5\r\nzelda\r\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := redis.NewStore()
			store.Set(testCase.key, testCase.value)
			clock := FakeClock{}

			result := redis.NewGetCommand(store, clock, testCase.key).Run()

			if result != testCase.response {
				t.Errorf(`command expected to return %#v but was %#v`, testCase.response, result)
			}
		})
	}
}

func TestGetCommand_KeyIsAbsent(t *testing.T) {
	store := redis.NewStore()
	clock := FakeClock{}

	result := redis.NewGetCommand(store, clock, "link").Run()

	if result != redisNullBulkString {
		t.Errorf(`command expected to return %#v but was %#v`, redisNullBulkString, result)
	}
}

func TestGetCommand_EntryHasExpired(t *testing.T) {
	store := redis.NewStore()
	store.SetWithExpiryTime("link", "zelda", time.Unix(0, 0))
	clock := FakeClock{CurrentTime: time.Unix(0, 1)}

	result := redis.NewGetCommand(store, clock, "link").Run()

	if result != redisNullBulkString {
		t.Errorf(`command expected to return %#v but was %#v`, redisNullBulkString, result)
	}
}

func TestPingCommand(t *testing.T) {
	response := redis.PingCommand{}.Run()
	if response != "+PONG\r\n" {
		t.Errorf(`command expected to return "+PONG\r\n" but was %#v`, response)
	}
}

func TestSetCommand(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		value redis.StoreValue
	}{
		{
			name:  "(grape: banana)",
			key:   "grape",
			value: redis.NewStoreValue("banana"),
		},
		{
			name:  "(link: zelda)",
			key:   "link",
			value: redis.NewStoreValue("zelda"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := redis.NewStore()

			response := redis.NewSetCommand(store, testCase.key, testCase.value.Data()).Run()

			if response != "+OK\r\n" {
				t.Errorf(`command expected to return "+OK\r\n" but was %#v`, response)
			}
			if result, ok := store.Get(testCase.key); !ok || result != testCase.value {
				t.Errorf(
					`command expected to contain key-value pair (%s: %v) but was %#v`,
					testCase.key, testCase.value, store,
				)
			}
		})
	}
}

func TestSetCommandWithExpiryTime(t *testing.T) {
	store := redis.NewStore()

	command := redis.NewSetCommand(
		store,
		"link",
		"zelda",
		redis.ExpiryTime(time.UnixMilli(0)),
	)
	response := command.Run()

	if response != "+OK\r\n" {
		t.Errorf(`command expected to return "+OK\r\n" but was %#v`, response)
	}
	value, ok := store.Get("link")
	if !ok {
		t.Errorf(`ok expected to be false but was true`)
	}
	if value.Data() != "zelda" {
		t.Errorf(`storeValue.Data() expected to be "zelda" but was %#v`, value.Data())
	}
	if value.ExpiryTime() == nil || !(*value.ExpiryTime()).Equal(time.UnixMilli(0)) {
		t.Errorf(
			`value.ExpiryTime() expected to be &time.UnixMilli(0) but was %#v`,
			value.ExpiryTime(),
		)
	}
}

func TestSetCommand_Equal(t *testing.T) {
	emptyStore1 := redis.NewStore()
	emptyStore2 := redis.NewStore()
	singleEntryStore := redis.NewStore()
	singleEntryStore.Set("foo", "bar")
	s1 := redis.NewSetCommand(
		emptyStore1,
		"link",
		"zelda",
	)
	s2 := redis.NewSetCommand(
		emptyStore1,
		"link",
		"zelda",
	)
	s3 := redis.NewSetCommand(
		emptyStore2,
		"link",
		"zelda",
	)
	s4 := redis.NewSetCommand(
		singleEntryStore,
		"link",
		"zelda",
	)
	s5 := redis.NewSetCommand(
		emptyStore1,
		"grape",
		"zelda",
	)
	s6 := redis.NewSetCommand(
		emptyStore1,
		"link",
		"banana",
	)
	s7 := redis.NewSetCommand(
		emptyStore1,
		"link",
		"zelda",
		redis.ExpiryTime(time.UnixMilli(0)),
	)
	s8 := redis.NewSetCommand(
		emptyStore1,
		"link",
		"zelda",
		redis.ExpiryTime(time.UnixMilli(0)),
	)
	tests := []struct {
		name  string
		this  *redis.SetCommand
		other *redis.SetCommand
		want  bool
	}{
		{
			name:  "s1.Equal(s1)",
			this:  s1,
			other: s1,
			want:  true,
		},
		{
			name:  "s1.Equal(s2)",
			this:  s1,
			other: s2,
			want:  true,
		},
		{
			name:  "s1.Equal(s3)",
			this:  s1,
			other: s3,
			want:  true,
		},
		{
			name:  "!s1.Equal(s4)",
			this:  s1,
			other: s4,
			want:  false,
		},
		{
			name:  "!s1.Equal(s5)",
			this:  s1,
			other: s5,
			want:  false,
		},
		{
			name:  "!s1.Equal(s6)",
			this:  s1,
			other: s6,
			want:  false,
		},
		{
			name:  "!s1.Equal(s7)",
			this:  s1,
			other: s7,
			want:  false,
		},
		{
			name:  "s7.Equal(s8)",
			this:  s7,
			other: s8,
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.this.Equal(tt.other); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
