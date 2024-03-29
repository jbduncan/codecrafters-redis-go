package redis

import (
	"fmt"
)

type Command interface {
	Run() string
}

type EchoCommand string

func (e EchoCommand) Run() string {
	return bulkString(string(e))
}

func NewGetCommand(store *Store, key string) GetCommand {
	return GetCommand{
		store: store,
		key:   key,
	}
}

type GetCommand struct {
	store *Store
	key   string
}

func (g GetCommand) Run() string {
	result, ok := g.store.Get(g.key)
	if !ok {
		return "$-1\r\n"
	}
	return bulkString(result)
}

var PingCommand Command = pingCommand{}

type pingCommand struct{}

func (p pingCommand) Run() string {
	return simpleString("PONG")
}

func NewSetCommand(store *Store, key, value string) SetCommand {
	return SetCommand{
		store: store,
		key:   key,
		value: value,
	}
}

type SetCommand struct {
	store *Store
	key   string
	value string
}

func (s SetCommand) Run() string {
	s.store.Set(s.key, s.value)
	return simpleString("OK")
}

func bulkString(s string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

func simpleString(s string) string {
	return fmt.Sprintf("+%s\r\n", s)
}
