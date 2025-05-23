package main

import (
	"fmt"
	"io"
	"log/slog"
)

func printErr(err error) {
	slog.Error(fmt.Sprintf("%v", err))
}

func errorHandlingClose(closer io.Closer) {
	if err := closer.Close(); err != nil {
		printErr(err)
	}
}
