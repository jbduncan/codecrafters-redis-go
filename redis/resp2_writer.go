package redis

import (
	"fmt"
	"io"
)

type RESP2Writer struct {
	writer io.Writer
}

func NewRESP2Writer(w io.Writer) *RESP2Writer {
	return &RESP2Writer{
		writer: w,
	}
}

func (w *RESP2Writer) Write(value Value) error {
	switch value.(type) {
	case Array:
		return w.writeArray(value.(Array))
	case BulkError:
		return w.writeBulkError(value.(BulkError))
	case BulkString:
		return w.writeBulkString(value.(BulkString))
	case Integer:
		return w.writeInteger(value.(Integer))
	case NullArray:
		return w.writeNullArray()
	case NullBulkString:
		return w.writeNullBulkString()
	case SimpleError:
		return w.writeSimpleError(value.(SimpleError))
	case SimpleString:
		return w.writeSimpleString(value.(SimpleString))
	}
	return unreachable[error]()
}

func (w *RESP2Writer) writeArray(a Array) error {
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

func (w *RESP2Writer) writeBulkError(berr BulkError) error {
	s := berr.Message()
	_, err := fmt.Fprintf(w.writer, "!%d\r\n%s\r\n", len(s), s)
	return err
}

func (w *RESP2Writer) writeBulkString(s BulkString) error {
	_, err := fmt.Fprintf(w.writer, "$%d\r\n%s\r\n", len(s), s)
	return err
}

func (w *RESP2Writer) writeInteger(i Integer) error {
	_, err := fmt.Fprintf(w.writer, ":%d\r\n", i)
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

func (w *RESP2Writer) writeSimpleError(serr SimpleError) error {
	_, err := fmt.Fprintf(w.writer, "-%s\r\n", serr)
	return err
}

func (w *RESP2Writer) writeSimpleString(ss SimpleString) error {
	_, err := fmt.Fprintf(w.writer, "+%s\r\n", ss)
	return err
}

func (w *RESP2Writer) wrapAsInternalWriterError(cause error) error {
	return fmt.Errorf("internal RESP2Writer error: %w", cause)
}
