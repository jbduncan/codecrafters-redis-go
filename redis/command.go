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
