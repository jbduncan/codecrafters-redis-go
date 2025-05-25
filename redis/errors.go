package redis

import (
	"fmt"
	"io"
	"log/slog"
)

func logError(err error) {
	slog.Error(fmt.Sprintf("%v", err))
}

func closeAndLogError(closer io.Closer) {
	if err := closer.Close(); err != nil {
		logError(err)
	}
}
