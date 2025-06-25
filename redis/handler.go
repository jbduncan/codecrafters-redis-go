package redis

import (
	"fmt"
	"strings"
)

type Handler interface {
	Command() string
	Handle(args Array[BulkString]) Value
}

func wrongNumberOfArgumentsError(h Handler) BulkError {
	return NewBulkError(
		fmt.Sprintf(
			"-ERR wrong number of arguments for '%s' command",
			strings.ToLower(h.Command()),
		),
	)
}
