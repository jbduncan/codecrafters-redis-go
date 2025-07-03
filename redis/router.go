package redis

import (
	"fmt"
	"strings"
)

type Router struct{}

func NewRouter() *Router {
	return &Router{}
}

func (r Router) Route(value Value) Value {
	if cmd, ok := value.(SimpleString); ok {
		if strings.EqualFold(string(cmd), "ping") {
			return PingHandler{}.Handle(nil)
		}
		return MakeBulkError(fmt.Sprintf("-ERR unknown command `%s`", cmd))
	}

	arr, ok := value.(Array[BulkString])
	if !ok {
		return MakeBulkError("-ERR unknown type")
	}
	if len(arr) == 0 {
		return MakeBulkError("-ERR no command")
	}

	cmd := arr[0]
	args := arr[1:]
	switch strings.ToLower(string(cmd)) {
	case "echo":
		return EchoHandler{}.Handle(args)
	case "ping":
		return PingHandler{}.Handle(args)
	}
	return MakeBulkError(fmt.Sprintf("-ERR unknown command `%s`", cmd))
}
