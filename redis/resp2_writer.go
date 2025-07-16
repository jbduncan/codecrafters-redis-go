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

func (w *RESP2Writer) Write(value Value) error {
	// TODO: other kinds of values from values.go
	switch value.(type) {
	case BulkString:
		// TODO: error
		_ = w.writeBulkString(value)
	case Integer:
		// TODO: error
		_ = w.writeInteger(value)
	case SimpleString:
		// TODO: error
		_ = w.writeSimpleString(value)
	}
	return nil
}

func (w *RESP2Writer) writeBulkString(value Value) error {
	s := string(value.(BulkString))
	_, err := fmt.Fprintf(w.writer, "$%d\r\n%s\r\n", len(s), s)
	return err
}

func (w *RESP2Writer) writeInteger(value Value) error {
	_, err := fmt.Fprintf(w.writer, ":%d\r\n", value)
	return err
}

func (w *RESP2Writer) writeSimpleString(value Value) error {
	s := string(value.(SimpleString))
	_, err := fmt.Fprintf(w.writer, "+%s\r\n", s)
	return err
}
