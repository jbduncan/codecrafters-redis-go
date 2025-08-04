package redis

import (
	"fmt"
	"strings"
)

type Router interface {
	Route(value InputValue) Value
}

type DefaultRouter struct {
	echoHandler Handler
	pingHandler Handler
}

func NewDefaultRouter(
	echoHandler Handler,
	pingHandler Handler,
) *DefaultRouter {
	return &DefaultRouter{
		echoHandler: echoHandler,
		pingHandler: pingHandler,
	}
}

func (r DefaultRouter) Route(value InputValue) Value {
	// TODO: consider supporting "Inline commands":
	//       https://redis.io/docs/latest/develop/reference/protocol-spec/#inline-commands

	switch value.(type) {
	case SimpleString:
		cmd := string(value.(SimpleString))
		if strings.EqualFold(cmd, "ping") {
			return r.pingHandler.Handle(nil)
		}
		return unknownCommandError(cmd)
	case BulkStringArray:
		arr := value.(BulkStringArray)
		if len(arr) == 0 {
			return MakeBulkError("ERR no command")
		}

		cmd := string(arr[0])
		args := arr[1:]
		switch strings.ToLower(cmd) {
		case "echo":
			return r.echoHandler.Handle(args)
		case "ping":
			return r.pingHandler.Handle(args)
		}
		return unknownCommandError(cmd)
	}

	return unreachable[Value]()
}

func unknownCommandError(cmd string) BulkError {
	return MakeBulkError(fmt.Sprintf("ERR unknown command `%s`", cmd))
}
