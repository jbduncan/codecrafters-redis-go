package redis

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

var (
	incompleteRequestError = NewSimpleError("-ERR Protocol error: incomplete request")

	// TODO: remove
	internalScannerSimpleError = errors.New("internal scanner error")
)

const (
	bulkString = '$'
	array      = '*'
)

type RESP2Scanner struct {
	reader *bufio.Reader
	buf    []byte
}

func NewRESP2Scanner(r io.Reader) *RESP2Scanner {
	s := &RESP2Scanner{
		reader: bufio.NewReader(r),
	}
	s.resetToken()
	return s
}

// TODO: consider turning Scan() into an iter.Seq2[Value, error] or
//       a Scan()/Token()/Err() arrangement a la bufio.Scanner

// Scan returns the next redis.Value from the underlying reader.
//
// If no more values are available, an io.EOF error is returned and no more
// calls to Scan should be made.
//
// If the next value is not valid syntax according to the Redis RESP2 protocol,
// then an error of type redis.ErrorValue is returned, suitable for sending
// back to the client via redis.Writer.Write.
//
// If any other sort of error occurred, then a generic error is returned.
func (s *RESP2Scanner) Scan() (Value, error) {
	typ, err := s.advanceFirstByte()
	if err != nil {
		// This returns io.EOF if there is no more input to process.
		return nil, err
	}

	var value Value
	switch typ {
	case array:
		value, err = s.array()
	default:
		s.addToToken(typ)
		value, err = s.simpleString()
	}

	s.resetToken()
	return value, err
}

func (s *RESP2Scanner) arrayElement() (Value, error) {
	typ, err := s.advanceFirstByte()
	if errors.Is(err, io.EOF) {
		return nil, incompleteRequestError
	}
	if err != nil {
		return nil, err
	}

	var value Value
	switch typ {
	case bulkString:
		value, err = s.bulkString()
	default:
		return nil, NewSimpleError(
			fmt.Sprintf(
				"-ERR Protocol error: expected '$', got '%s'",
				s.escape(typ),
			),
		)
	}

	s.resetToken()
	return value, err
}

func (s *RESP2Scanner) bulkString() (Value, error) {
	length, err := s.unsignedInt()
	if err != nil {
		return nil, err
	}

	for range length {
		char, err := s.advance()
		if err != nil {
			return nil, err
		}
		s.addToToken(char)
	}

	if err := s.consumeCR(); err != nil {
		return nil, err
	}
	if err := s.consumeLF(); err != nil {
		return nil, err
	}

	return BulkString(s.token()), nil
}

func (s *RESP2Scanner) array() (Value, error) {
	length, err := s.unsignedInt()
	if err != nil {
		return nil, err
	}
	// TODO: error if length is greater than size of int64

	result := Array{}
	for range length {
		element, err := s.arrayElement()
		if err != nil {
			return nil, err
		}

		result = append(result, element)
	}

	return result, nil
}

func (s *RESP2Scanner) simpleString() (Value, error) {
	for {
		char, err := s.advance()
		if err != nil {
			return nil, err
		}

		if !s.isCR(char) {
			s.addToToken(char)
			continue
		}

		if err := s.consumeLF(); err != nil {
			return nil, err
		}

		return SimpleString(s.token()), nil
	}
}

func (s *RESP2Scanner) unsignedInt() (uint64, error) {
	// Process first digit.
	digit, err := s.peek()
	if err != nil {
		return 0, err
	}

	if !s.isDigit(digit) {
		return 0, NewSimpleError(
			fmt.Sprintf(
				"-ERR Protocol error: expected digit, got '%s'",
				s.escape(digit),
			),
		)
	}

	result := uint64(s.asciiDigitToInt(digit))

	// Consume the digit.
	if _, err := s.advance(); err != nil {
		return 0, err
	}

	for {
		digit, err := s.peek()
		if err != nil {
			return 0, err
		}

		if !s.isDigit(digit) {
			// All digits have been processed.
			break
		}

		result = (10 * result) + uint64(s.asciiDigitToInt(digit))

		// Consume the digit.
		if _, err := s.advance(); err != nil {
			return 0, err
		}
	}

	if err := s.consumeCR(); err != nil {
		return 0, err
	}
	if err := s.consumeLF(); err != nil {
		return 0, err
	}

	return result, nil
}

func (s *RESP2Scanner) consumeCR() error {
	cr, err := s.advance()
	if err != nil {
		return err
	}
	if !s.isCR(cr) {
		return NewSimpleError(
			fmt.Sprintf(
				`-ERR Protocol error: expected '\r', got '%s'`,
				s.escape(cr),
			),
		)
	}
	return nil
}

func (s *RESP2Scanner) consumeLF() error {
	lf, err := s.advance()
	if err != nil {
		return err
	}
	if !s.isLF(lf) {
		return NewSimpleError(
			fmt.Sprintf(
				`-ERR Protocol error: expected '\n', got '%s'`,
				s.escape(lf),
			),
		)
	}
	return nil
}

func (s *RESP2Scanner) escape(b byte) string {
	var escaped string
	if b == '\r' {
		escaped = `\r`
	} else if b == '\n' {
		escaped = `\n`
	} else {
		escaped = string(b)
	}
	return escaped
}

func (s *RESP2Scanner) advanceFirstByte() (byte, error) {
	b, err := s.reader.ReadByte()
	if errors.Is(err, io.EOF) {
		return 0, io.EOF
	}
	if err != nil {
		return 0, s.wrapAsInternalScannerSimpleError(err)
	}

	return b, nil
}

func (s *RESP2Scanner) advance() (byte, error) {
	b, err := s.reader.ReadByte()
	if errors.Is(err, io.EOF) {
		return 0, incompleteRequestError
	}
	if err != nil {
		return 0, s.wrapAsInternalScannerSimpleError(err)
	}

	return b, nil
}

func (s *RESP2Scanner) peek() (byte, error) {
	bs, err := s.reader.Peek(1)
	if errors.Is(err, io.EOF) {
		return 0, incompleteRequestError
	}
	if err != nil {
		return 0, s.wrapAsInternalScannerSimpleError(err)
	}

	return bs[0], nil
}

// TODO: remove
func (s *RESP2Scanner) wrapAsInternalScannerSimpleError(cause error) error {
	return fmt.Errorf("%v: %v", internalScannerSimpleError, cause)
}

func (s *RESP2Scanner) isCR(b byte) bool {
	return b == '\r'
}

func (s *RESP2Scanner) isLF(b byte) bool {
	return b == '\n'
}

func (s *RESP2Scanner) isCRLF(b byte) bool {
	return s.isCR(b) || s.isLF(b)
}

func (s *RESP2Scanner) isDigit(b byte) bool {
	return '0' <= b && b <= '9'
}

func (s *RESP2Scanner) asciiDigitToInt(b byte) byte {
	return b - '0'
}

func (s *RESP2Scanner) addToToken(b byte) {
	s.buf = append(s.buf, b)
}

func (s *RESP2Scanner) token() []byte {
	return s.buf
}

func (s *RESP2Scanner) resetToken() {
	s.buf = make([]byte, 0, 16)
}
