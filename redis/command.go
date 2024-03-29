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

func NewGetCommand(key string, store map[string]string) GetCommand {
	return GetCommand{
		key:   key,
		store: store,
	}
}

type GetCommand struct {
	key   string
	store map[string]string
}

func (g GetCommand) Run() string {
	result, ok := g.store[g.key]
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

func NewSetCommand(key, value string, store map[string]string) SetCommand {
	return SetCommand{
		key:   key,
		value: value,
		store: store,
	}
}

type SetCommand struct {
	key   string
	value string
	store map[string]string
}

func (s SetCommand) Run() string {
	s.store[s.key] = s.value
	return simpleString("OK")
}

func bulkString(s string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

func simpleString(s string) string {
	return fmt.Sprintf("+%s\r\n", s)
}
