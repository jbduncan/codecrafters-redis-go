package redis

import (
	"fmt"
	"strings"
)

type EchoHandler struct{}

func (h EchoHandler) Command() string {
	return "ECHO"
}

func (h EchoHandler) Handle(args Array[BulkString]) Value {
	switch len(args) {
	case 1:
		return args[0]
	default:
		return NewBulkError(
			fmt.Sprintf(
				"-ERR wrong number of arguments for '%s' command",
				strings.ToLower(h.Command()),
			),
		)
	}
}
