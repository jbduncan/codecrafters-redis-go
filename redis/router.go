package redis

import (
	"fmt"
	"strings"
)

type Router struct{}

func NewRouter() *Router {
	return &Router{}
}

func (r Router) Route(value InputValue) Value {
	// TODO: consider supporting "Inline commands":
	//       https://redis.io/docs/latest/develop/reference/protocol-spec/#inline-commands

	switch value.(type) {
	case SimpleString:
		cmd := string(value.(SimpleString))
		if strings.EqualFold(cmd, "ping") {
			return PingHandler{}.Handle(nil)
		}
		return unknownCommandError(cmd)
	case BulkStringArray:
		arr := value.(BulkStringArray)
		if len(arr) == 0 {
			return MakeBulkError("-ERR no command")
		}

		cmd := string(arr[0])
		args := arr[1:]
		switch strings.ToLower(cmd) {
		case "echo":
			return EchoHandler{}.Handle(args)
		case "ping":
			return PingHandler{}.Handle(args)
		}
		return unknownCommandError(cmd)
	}

	return unreachable[Value]()
}

func unknownCommandError(cmd string) BulkError {
	return MakeBulkError(fmt.Sprintf("-ERR unknown command `%s`", cmd))
}
