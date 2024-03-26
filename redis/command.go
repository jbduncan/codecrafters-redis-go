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
