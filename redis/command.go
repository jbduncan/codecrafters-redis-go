package redis

import (
	"fmt"
)

type Command interface {
	Run() string
}

type EchoCommand string

func (e EchoCommand) Run() string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(e), e)
}

var PingCommand Command = pingCommand{}

type pingCommand struct{}

func (p pingCommand) Run() string {
	return "+PONG\r\n"
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
	return "+OK\r\n"
}
