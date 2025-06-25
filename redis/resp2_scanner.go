package redis

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

var (
	incompleteRequestError = NewBulkError("-ERR Protocol error: incomplete request")
)

const (
	bulkStringType = '$'
	arrayType      = '*'
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
	case arrayType:
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
	case bulkStringType:
		value, err = s.bulkString()
	default:
		return nil, s.expectedGotError("'$'", typ)
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

	result := Array[Value]{}
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
	result, err := s.firstUnsignedDigit()
	if err != nil {
		return 0, err
	}

	result, err = s.remainingUnsignedDigits(result)
	if err != nil {
		return 0, err
	}

	if err := s.consumeCR(); err != nil {
		return 0, err
	}
	if err := s.consumeLF(); err != nil {
		return 0, err
	}

	return uint64(result), nil
}

func (s *RESP2Scanner) remainingUnsignedDigits(sum int32) (int32, error) {
	for {
		digit, err := s.peek()
		if err != nil {
			return 0, err
		}

		if !s.isDigit(digit) {
			// All digits have been processed.
			break
		}

		var overflow bool
		sum, overflow = s.checkedMultiply(10, sum)
		if overflow {
			return 0, NewBulkError(
				"-ERR Protocol error: invalid multibulk length",
			)
		}
		sum, overflow = s.checkedAdd(sum, s.asciiDigitToInt32(digit))
		if overflow {
			return 0, NewBulkError(
				"-ERR Protocol error: invalid multibulk length",
			)
		}

		// Consume the digit.
		if _, err := s.advance(); err != nil {
			return 0, err
		}
	}

	return sum, nil
}

func (s *RESP2Scanner) firstUnsignedDigit() (int32, error) {
	// Process first digit.
	digit, err := s.peek()
	if err != nil {
		return 0, err
	}

	if !s.isDigit(digit) {
		return 0, s.expectedGotError("digit", digit)
	}

	result := s.asciiDigitToInt32(digit)

	// Consume the digit.
	if _, err := s.advance(); err != nil {
		return 0, err
	}

	return result, nil
}

func (s *RESP2Scanner) checkedAdd(a, b int32) (int32, bool) {
	result := int64(a) + int64(b)
	if result != int64(int32(result)) {
		return 0, true
	}
	return int32(result), false
}

func (s *RESP2Scanner) checkedMultiply(a, b int32) (int32, bool) {
	result := int64(a) * int64(b)
	if result != int64(int32(result)) {
		return 0, true
	}
	return int32(result), false
}

func (s *RESP2Scanner) consumeCR() error {
	cr, err := s.advance()
	if err != nil {
		return err
	}
	if !s.isCR(cr) {
		return s.expectedGotError(`'\r'`, cr)
	}
	return nil
}

func (s *RESP2Scanner) consumeLF() error {
	lf, err := s.advance()
	if err != nil {
		return err
	}
	if !s.isLF(lf) {
		return s.expectedGotError(`'\n'`, lf)
	}
	return nil
}

func (s *RESP2Scanner) escape(b byte) string {
	switch b {
	case '\r':
		return `\r`
	case '\n':
		return `\n`
	default:
		return string(b)
	}
}

func (s *RESP2Scanner) advanceFirstByte() (byte, error) {
	b, err := s.reader.ReadByte()
	if errors.Is(err, io.EOF) {
		return 0, io.EOF
	}
	if err != nil {
		return 0, s.wrapAsInternalScannerError(err)
	}

	return b, nil
}

func (s *RESP2Scanner) advance() (byte, error) {
	b, err := s.reader.ReadByte()
	if errors.Is(err, io.EOF) {
		return 0, incompleteRequestError
	}
	if err != nil {
		return 0, s.wrapAsInternalScannerError(err)
	}

	return b, nil
}

func (s *RESP2Scanner) peek() (byte, error) {
	b, err := s.reader.ReadByte()
	if errors.Is(err, io.EOF) {
		return 0, incompleteRequestError
	}
	if err != nil {
		return 0, s.wrapAsInternalScannerError(err)
	}

	_ = s.reader.UnreadByte()

	return b, nil
}

func (s *RESP2Scanner) wrapAsInternalScannerError(cause error) error {
	return fmt.Errorf("internal scanner error: %v", cause)
}

func (s *RESP2Scanner) expectedGotError(expected string, got byte) error {
	return NewBulkError(
		fmt.Sprintf(
			`-ERR Protocol error: expected %s, got '%s'`,
			expected,
			s.escape(got),
		),
	)
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

func (s *RESP2Scanner) asciiDigitToInt32(b byte) int32 {
	return int32(b - '0')
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
