package redis

import (
	"sync"
	"time"
)

func NewStore() *Store {
	return &Store{
		entries: new(sync.Map),
	}
}

type Store struct {
	entries *sync.Map
}

func (s *Store) Get(key string) (result StoreValue, ok bool) {
	value, ok := s.entries.Load(key)
	if ok {
		result = value.(StoreValue)
	}
	return
}

func (s *Store) Set(key, value string) {
	s.entries.Store(key, StoreValue{data: value})
}

func (s *Store) SetWithExpiryTime(key, value string, expiryTime time.Time) {
	s.entries.Store(key, StoreValue{
		data:       value,
		expiryTime: &expiryTime,
	})
}

func NewStoreValue(data string) StoreValue {
	return StoreValue{
		data: data,
	}
}

func NewStoreValueWithExpiryTime(data string, expiryTime time.Time) StoreValue {
	return StoreValue{
		data:       data,
		expiryTime: &expiryTime,
	}
}

type StoreValue struct {
	data       string
	expiryTime *time.Time
}

func (s StoreValue) Data() string {
	return s.data
}

func (s StoreValue) ExpiryTime() *time.Time {
	return s.expiryTime
}
