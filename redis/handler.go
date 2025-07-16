package redis

import (
	"fmt"
)

type Handler interface {
	Handle(args BulkStringArray) Value
}

func wrongNumberOfArgumentsError(cmd string) BulkError {
	return MakeBulkError(
		fmt.Sprintf("-ERR wrong number of arguments for '%s' command", cmd),
	)
}
