package redis_test

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

func TestStore(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "(link: zelda)",
			key:   "link",
			value: "zelda",
		},
		{
			name:  "(grape: banana)",
			key:   "grape",
			value: "banana",
		},
	}

	store := redis.NewStore()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store.Set(testCase.key, testCase.value)

			result, ok := store.Get(testCase.key)

			if !ok {
				t.Errorf(`ok expected to be true but was false`)
			}
			if result != testCase.value {
				t.Errorf(`result expected to be %#v but was %#v`, testCase.value, result)
			}
		})
	}

	t.Run("absent key", func(t *testing.T) {
		result, ok := store.Get("absent-key")

		if ok {
			t.Errorf(`ok expected to be false but was true`)
		}
		if result != "" {
			t.Errorf(`result expected to be empty but was %#v`, result)
		}
	})
}
