package redis

import (
	"io"
)

type Type int

const (
	Echo Type = iota
	Ping
)

type Command struct {
	typ Type
	f   func(w io.Writer) error
}

func (c Command) Type() Type {
	return c.typ
}

func (c Command) Run(w io.Writer) error {
	return c.f(w)
}

type Command2 interface {
	Run() (string, error)
}

type EchoCommand2 string

func (e EchoCommand2) Run() (string, error) {
	//TODO implement me
	panic("implement me")
}

type PingCommand2 struct{}

func (p PingCommand2) Run() (string, error) {
	//TODO implement me
	panic("implement me")
}
