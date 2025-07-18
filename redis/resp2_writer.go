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
	// TODO: other kinds of values from value.go
	switch value.(type) {
	case Array:
		return w.writeArray(value)
	case BulkError:
		// TODO: error
		_ = w.writeBulkError(value)
	case BulkString:
		// TODO: error
		_ = w.writeBulkString(value)
	case Integer:
		// TODO: error
		_ = w.writeInteger(value)
	case NullArray:
		// TODO: error
		_ = w.writeNullArray()
	case NullBulkString:
		// TODO: error
		_ = w.writeNullBulkString()
	case SimpleError:
		// TODO: error
		_ = w.writeSimpleError(value)
	case SimpleString:
		return w.writeSimpleString(value)
	}
	// TODO: use `return unreachable[error]()`
	return nil
}

func (w *RESP2Writer) writeArray(value Value) error {
	a := value.(Array)
	if _, err := fmt.Fprintf(w.writer, "*%d\r\n", len(a)); err != nil {
		return w.wrapAsInternalWriterError(err)
	}
	for _, v := range a {
		if err := w.Write(v); err != nil {
			return w.wrapAsInternalWriterError(err)
		}
	}
	return nil
}

func (w *RESP2Writer) writeBulkError(value Value) error {
	s := value.(BulkError).Message()
	_, err := fmt.Fprintf(w.writer, "!%d\r\n%s\r\n", len(s), s)
	return err
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

func (w *RESP2Writer) writeNullArray() error {
	_, err := io.WriteString(w.writer, "*-1\r\n")
	return err
}

func (w *RESP2Writer) writeNullBulkString() error {
	_, err := io.WriteString(w.writer, "$-1\r\n")
	return err
}

func (w *RESP2Writer) writeSimpleError(value Value) error {
	_, err := fmt.Fprintf(w.writer, "-%s\r\n", value)
	return err
}

func (w *RESP2Writer) writeSimpleString(value Value) error {
	s := string(value.(SimpleString))
	_, err := fmt.Fprintf(w.writer, "+%s\r\n", s)
	return err
}

func (w *RESP2Writer) wrapAsInternalWriterError(cause error) error {
	return fmt.Errorf("internal RESP2Writer error: %w", cause)
}
