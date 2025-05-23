package errorsx

import (
	"fmt"
	"io"
	"log/slog"
)

func Log(err error) {
	slog.Error(fmt.Sprintf("%v", err))
}

func Close(closer io.Closer) {
	if err := closer.Close(); err != nil {
		Log(err)
	}
}
