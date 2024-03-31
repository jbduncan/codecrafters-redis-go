package redis_test

import (
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

func TestNewStoreValue(t *testing.T) {
	storeValue := redis.NewStoreValue("link")

	if storeValue.Data() != "link" {
		t.Errorf(`storeValue.Data() expected to be "link" but was %#v`, storeValue.Data())
	}
	if storeValue.ExpiryTime() != nil {
		t.Errorf(`storeValue.ExpiryTime() expected to be nil but was %#v`, storeValue.Data())
	}
}

func TestNewStoreValueWithExpiryTime(t *testing.T) {
	storeValue := redis.NewStoreValueWithExpiryTime("link", time.UnixMilli(0))

	if storeValue.Data() != "link" {
		t.Errorf(`storeValue.Data() expected to be "link" but was %#v`, storeValue.Data())
	}
	if storeValue.ExpiryTime() == nil || !(*storeValue.ExpiryTime()).Equal(time.UnixMilli(0)) {
		t.Errorf(
			`storeValue.ExpiryTime() expected to be &time.UnixMilli(0) but was %#v`,
			storeValue.ExpiryTime(),
		)
	}
}

func TestStore(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		value redis.StoreValue
	}{
		{
			name:  "(link: zelda)",
			key:   "link",
			value: redis.NewStoreValue("zelda"),
		},
		{
			name:  "(grape: banana)",
			key:   "grape",
			value: redis.NewStoreValue("banana"),
		},
	}

	store := redis.NewStore()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store.Set(testCase.key, testCase.value.Data())

			result, ok := store.Get(testCase.key)

			if !ok {
				t.Errorf(`ok expected to be true but was false`)
			}
			if result.Data() != testCase.value.Data() {
				t.Errorf(
					`result.Data() expected to be %#v but was %#v`,
					testCase.value.Data(), result.Data(),
				)
			}
			if result.ExpiryTime() != nil {
				t.Errorf(
					`result.ExpiryTime() expected to be nil but was %#v`,
					result.ExpiryTime(),
				)
			}
		})
	}

	t.Run("(link: zelda) with expiry time", func(t *testing.T) {
		key := "link"
		value := redis.NewStoreValueWithExpiryTime("zelda", time.UnixMilli(0))
		store.SetWithExpiryTime(key, value.Data(), *value.ExpiryTime())

		result, ok := store.Get(key)

		if !ok {
			t.Errorf(`ok expected to be true but was false`)
		}
		if result.Data() != value.Data() {
			t.Errorf(`result.Data() expected to be %#v but was %#v`, value.Data(), result.Data())
		}
		if !result.ExpiryTime().Equal(*value.ExpiryTime()) {
			t.Errorf(
				`result.ExpiryTime() expected to be %#v but was %#v`,
				value.ExpiryTime(), result.ExpiryTime(),
			)
		}
	})

	t.Run("absent key", func(t *testing.T) {
		result, ok := store.Get("absent-key")

		if ok {
			t.Errorf(`ok expected to be false but was true`)
		}
		if result.Data() != "" {
			t.Errorf(`result.Data() expected to be empty but was %#v`, result)
		}
		if result.ExpiryTime() != nil {
			t.Errorf(
				`result.ExpiryTime() expected to be nil but was %#v`,
				result.ExpiryTime(),
			)
		}
	})
}
