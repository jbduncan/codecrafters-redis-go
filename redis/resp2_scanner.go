package redis

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

var (
	invalidSyntaxError         = NewSimpleError("SYNTAX invalid syntax")
	internalScannerSimpleError = errors.New("internal scanner error")
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
	// TODO: handle:
	//   - Array
	//   - Other kinds of values in
	//     https://redis.io/docs/latest/develop/reference/protocol-spec
	b, err := s.advanceFirstByte()
	if err != nil {
		// This returns io.EOF if there is no more input to process.
		return nil, err
	}

	var value Value
	switch b {
	case '$':
		next, _ := s.peek()
		if next == '-' {
			value, err = s.nullBulkString()
		} else {
			value, err = s.bulkString()
		}
	case '*':
		next, _ := s.peek()
		if next == '-' {
			value, err = s.nullArray()
		} else {
			// Array scanning will go here.
		}
	case ':':
		value, err = s.signedInteger()
	default:
		s.addToToken(b)
		value, err = s.simpleString()
	}

	s.resetToken()
	return value, err
}

func (s *RESP2Scanner) bulkString() (Value, error) {
	length, err := s.unsignedInt()
	if err != nil {
		return nil, err
	}

	if err := s.consumeCR(); err != nil {
		return nil, err
	}
	if err := s.consumeLF(); err != nil {
		return nil, err
	}

	for range length {
		b, err := s.advance()
		if err != nil {
			return nil, err
		}
		s.addToToken(b)
	}

	if err := s.consumeCR(); err != nil {
		return nil, err
	}
	if err := s.consumeLF(); err != nil {
		return nil, err
	}

	return BulkString(s.token()), nil
}

func (s *RESP2Scanner) nullBulkString() (Value, error) {
	if err := s.nullSuffix(); err != nil {
		return nil, err
	}

	return RESP2NullBulkString{}, nil
}

func (s *RESP2Scanner) nullArray() (Value, error) {
	if err := s.nullSuffix(); err != nil {
		return nil, err
	}

	return RESP2NullArray{}, nil
}

func (s *RESP2Scanner) nullSuffix() error {
	// Consume the "-".
	if _, err := s.advance(); err != nil {
		return err
	}

	b, err := s.advance()
	if err != nil {
		return err
	}
	if b != '1' {
		return invalidSyntaxError
	}
	if err := s.consumeCR(); err != nil {
		return err
	}
	if err := s.consumeLF(); err != nil {
		return err
	}
	return nil
}

func (s *RESP2Scanner) simpleString() (Value, error) {
	for {
		b, err := s.advance()
		if err != nil {
			return nil, err
		}

		if !s.isCR(b) {
			s.addToToken(b)
			continue
		}

		if err := s.consumeLF(); err != nil {
			return nil, err
		}

		return SimpleString(s.token()), nil
	}
}

func (s *RESP2Scanner) unsignedInt() (uint64, error) {
	var result uint64
	for {
		b, err := s.peek()
		if err != nil {
			return 0, err
		}

		if !s.isDigit(b) {
			return result, nil
		}

		result = (10 * result) + uint64(s.asciiDigitToInt(b))

		// Consume the digit.
		if _, err := s.advance(); err != nil {
			return 0, err
		}
	}
}

func (s *RESP2Scanner) signedInteger() (Value, error) {
	value, err := s.signedInt()
	if err != nil {
		return nil, err
	}

	if err := s.consumeCR(); err != nil {
		return nil, err
	}
	if err := s.consumeLF(); err != nil {
		return nil, err
	}
	return Integer(value), nil
}

func (s *RESP2Scanner) signedInt() (int64, error) {
	result, negative, err := s.signAndFirstDigit()
	if err != nil {
		return 0, err
	}

	for {
		b, err := s.peek()
		if err != nil {
			return 0, err
		}

		if !s.isDigit(b) {
			if negative {
				result *= -1
			}
			return result, nil
		}

		result = s.appendInt64Digit(result, b)

		// Consume the digit.
		if _, err := s.advance(); err != nil {
			return 0, err
		}
	}
}

func (s *RESP2Scanner) signAndFirstDigit() (int64, bool, error) {
	var result int64
	var negative bool
	var signSeen bool

	b, err := s.advance()
	if err != nil {
		return 0, false, err
	}

	switch b {
	case '-':
		negative = true
		signSeen = true
	case '+':
		signSeen = true
	default:
		if s.isDigit(b) {
			result = s.appendInt64Digit(result, b)
		} else {
			return 0, false, invalidSyntaxError
		}
	}

	if signSeen {
		b, err := s.advance()
		if err != nil {
			return 0, false, err
		}

		if !s.isDigit(b) {
			return 0, false, invalidSyntaxError
		}
		result = s.appendInt64Digit(result, b)
	}

	return result, negative, nil
}

func (s *RESP2Scanner) appendInt64Digit(result int64, b byte) int64 {
	return (10 * result) + int64(s.asciiDigitToInt(b))
}

func (s *RESP2Scanner) consumeCR() error {
	b, err := s.advance()
	if err != nil {
		return err
	}
	if !s.isCR(b) {
		return invalidSyntaxError
	}
	return nil
}

func (s *RESP2Scanner) consumeLF() error {
	b, err := s.advance()
	if err != nil {
		return err
	}
	if !s.isLF(b) {
		return invalidSyntaxError
	}
	return nil
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
		return 0, invalidSyntaxError
	}
	if err != nil {
		return 0, s.wrapAsInternalScannerSimpleError(err)
	}

	return b, nil
}

func (s *RESP2Scanner) peek() (byte, error) {
	bs, err := s.reader.Peek(1)
	if errors.Is(err, io.EOF) {
		return 0, invalidSyntaxError
	}
	if err != nil {
		return 0, s.wrapAsInternalScannerSimpleError(err)
	}

	return bs[0], nil
}

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
