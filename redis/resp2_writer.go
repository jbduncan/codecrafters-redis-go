package redis

import (
	"fmt"
	"io"
	"slices"
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
	s := &stack{value}
	for len(*s) > 0 {
		if err := w.doWrite(s); err != nil {
			return w.wrapAsInternalWriterError(err)
		}
	}
	return nil
}

func (w *RESP2Writer) doWrite(s *stack) error {
	switch v := s.pop().(type) {
	case Array:
		return w.writeArray(v, s)
	case BulkError:
		return w.writeBulkError(v)
	case BulkString:
		return w.writeBulkString(v)
	case Integer:
		return w.writeInteger(v)
	case NullArray:
		return w.writeNullArray()
	case NullBulkString:
		return w.writeNullBulkString()
	case SimpleError:
		return w.writeSimpleError(v)
	case SimpleString:
		return w.writeSimpleString(v)
	default:
		return unreachable[error]()
	}
}

func (w *RESP2Writer) writeArray(a Array, s *stack) error {
	if _, err := fmt.Fprintf(w.writer, "*%d\r\n", len(a)); err != nil {
		return err
	}
	for _, v := range slices.Backward(a) {
		s.push(v)
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
