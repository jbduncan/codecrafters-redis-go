package redis

import (
	"fmt"
	"time"
)

type Command interface {
	Run() string
}

type EchoCommand string

func (e EchoCommand) Run() string {
	return bulkString(string(e))
}

func NewGetCommand(store *Store, clock Clock, key string) *GetCommand {
	return &GetCommand{
		store: store,
		clock: clock,
		key:   key,
	}
}

type GetCommand struct {
	store *Store
	clock Clock
	key   string
}

func (g GetCommand) Run() string {
	result, ok := g.store.Get(g.key)
	if !ok {
		return "$-1\r\n"
	}

	expiryTime := result.ExpiryTime()
	if expiryTime != nil && g.clock.Now().After(*expiryTime) {
		return "$-1\r\n"
	}

	return bulkString(result.Data())
}

type PingCommand struct{}

func (p PingCommand) Run() string {
	return simpleString("PONG")
}

func NewSetCommand(store *Store, key, value string, options ...func(*SetCommand)) *SetCommand {
	result := &SetCommand{
		store: store,
		key:   key,
		value: value,
	}
	for _, option := range options {
		option(result)
	}
	return result
}

type SetCommand struct {
	store      *Store
	key        string
	value      string
	expiryTime *time.Time
}

func (s *SetCommand) Run() string {
	if s.expiryTime == nil {
		s.store.Set(s.key, s.value)
	} else {
		s.store.SetWithExpiryTime(s.key, s.value, *s.expiryTime)
	}
	return simpleString("OK")
}

func ExpiryTime(t time.Time) func(*SetCommand) {
	return func(command *SetCommand) {
		command.expiryTime = &t
	}
}

func bulkString(s string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

func simpleString(s string) string {
	return fmt.Sprintf("+%s\r\n", s)
}
