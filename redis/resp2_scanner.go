package redis

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"iter"
)

var (
	incompleteRequestError = MakeBulkError("ERR Protocol error: incomplete request")
)

const (
	bulkStringType = '$'
	arrayType      = '*'

	// 512 megabytes
	mb512 = 51_200_000_000
)

type RESP2Scanner struct {
	reader   *bufio.Reader
	tokenBuf []byte
}

func NewRESP2Scanner(r io.Reader) *RESP2Scanner {
	s := &RESP2Scanner{
		reader: bufio.NewReader(io.LimitReader(r, mb512)),
	}
	s.resetToken()
	return s
}

// ScanAll returns an iterator that loops over all [redis.InputValue] instances
// from the underlying reader.
//
// If the next value is not valid syntax according to the
// [Redis RESP2 protocol], then the iterator returns an error of type
// redis.ErrorValue, suitable for sending back to the client via
// [redis.RESP2Writer.Write].
//
// If any other sort of error occurred, then the iterator returns a generic
// error.
//
// If any sort of error is returned, including redis.ErrorValue instances and
// generic errors, then the iterator will stop.
//
// [Redis RESP2 protocol]: https://redis.io/docs/latest/develop/reference/protocol-spec
func (s *RESP2Scanner) ScanAll() iter.Seq2[InputValue, error] {
	return func(yield func(InputValue, error) bool) {
		for {
			value, err := s.scan()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				yield(nil, err)
				return
			}
			if !yield(value, nil) {
				return
			}
		}
	}
}

func (s *RESP2Scanner) scan() (InputValue, error) {
	typ, err := s.advanceFirstByte()
	if err != nil {
		// This returns io.EOF if there is no more input to process.
		return nil, err
	}

	var value InputValue
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

func (s *RESP2Scanner) arrayElement() (BulkString, error) {
	typ, err := s.advanceFirstByte()
	if errors.Is(err, io.EOF) {
		return "", incompleteRequestError
	}
	if err != nil {
		return "", err
	}

	var value BulkString
	switch typ {
	case bulkStringType:
		value, err = s.bulkString()
	default:
		return "", s.expectedGotError("'$'", typ)
	}

	s.resetToken()
	return value, err
}

func (s *RESP2Scanner) bulkString() (BulkString, error) {
	length, err := s.unsignedInt()
	if err != nil {
		return "", err
	}

	for range length {
		char, err := s.advance()
		if err != nil {
			return "", err
		}
		s.addToToken(char)
	}

	if err := s.consumeCR(); err != nil {
		return "", err
	}
	if err := s.consumeLF(); err != nil {
		return "", err
	}

	return BulkString(s.token()), nil
}

func (s *RESP2Scanner) array() (BulkStringArray, error) {
	length, err := s.unsignedInt()
	if err != nil {
		return nil, err
	}

	result := BulkStringArray{}
	for range length {
		element, err := s.arrayElement()
		if err != nil {
			return nil, err
		}

		result = append(result, element)
	}

	return result, nil
}

func (s *RESP2Scanner) simpleString() (SimpleString, error) {
	for {
		char, err := s.advance()
		if err != nil {
			return "", err
		}

		if !s.isCR(char) {
			s.addToToken(char)
			continue
		}

		if err := s.consumeLF(); err != nil {
			return "", err
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
			return 0, MakeBulkError(
				"ERR Protocol error: invalid multibulk length",
			)
		}
		sum, overflow = s.checkedAdd(sum, s.asciiDigitToInt32(digit))
		if overflow {
			return 0, MakeBulkError(
				"ERR Protocol error: invalid multibulk length",
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
	return fmt.Errorf("internal RESP2Scanner error: %w", cause)
}

func (s *RESP2Scanner) expectedGotError(expected string, got byte) error {
	return MakeBulkError(
		fmt.Sprintf(
			`ERR Protocol error: expected %s, got '%s'`,
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
	s.tokenBuf = append(s.tokenBuf, b)
}

func (s *RESP2Scanner) token() []byte {
	return s.tokenBuf
}

func (s *RESP2Scanner) resetToken() {
	s.tokenBuf = make([]byte, 0, 16)
}
