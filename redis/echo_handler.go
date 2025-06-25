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
	return NewBulkError(
		fmt.Sprintf(
			"-ERR wrong number of arguments for '%s' command",
			strings.ToLower(h.Command()),
		),
	)
}
