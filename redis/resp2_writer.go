package redis

import (
	"fmt"
	"io"
)

type RESP2Writer struct {
	writer io.Writer
}

func NewRESP2Writer(w io.Writer) *RESP2Writer {
	rw := &RESP2Writer{
		writer: w,
	}
	return rw
}

func (w RESP2Writer) Write(value Value) error {
	var err error
	// TODO: other kinds of values from values.go
	switch value.(type) {
	case SimpleString:
		_, err = w.writer.Write([]byte(fmt.Sprintf("+%s\r\n", value)))
	case BulkString:
		_, err = w.writer.Write([]byte(fmt.Sprintf(
			"$%d\r\n%s\r\n",
			len(value.(BulkString)),
			value,
		)))
	case Integer:
		_, err = w.writer.Write([]byte(fmt.Sprintf(":%d\r\n", value)))
	}
	if err != nil {
		// TODO
	}
	return nil
}
