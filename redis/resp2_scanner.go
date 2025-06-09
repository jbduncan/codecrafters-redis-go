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

const (
	integer    = ':'
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

	// TODO: Apparently, from testing and reading the docs, the real Redis only
	//       supports single simple strings, arrays, pipelines of single simple
	//       strings and/or arrays and "Inline commands" as inputs.
	//       Furthermore, it needs array elements to have a type byte, so
	//       every type except simple strings are supported as elements.
	//       Therefore, we need two versions of Scan(): a public one that only
	//       supports simple strings and arrays, and a private one for array
	//       elements which supports all other types and can call itself
	//       recursively on sub-arrays.
	//       - https://redis.io/docs/latest/develop/reference/protocol-spec/#sending-commands-to-a-redis-server
	//       - https://redis.io/docs/latest/develop/reference/protocol-spec/#multiple-commands-and-pipelining
	//       - https://redis.io/docs/latest/develop/reference/protocol-spec/#inline-commands
	//       - https://redis.io/docs/latest/develop/reference/protocol-spec/#arrays

	var value Value
	switch typ {
	case integer:
		value, err = s.signedInteger()
	case bulkString:
		next, _ := s.peek()
		if next == '-' {
			value, err = s.nullBulkString()
		} else {
			value, err = s.bulkString()
		}
	case array:
		next, _ := s.peek()
		if next == '-' {
			value, err = s.nullArray()
		} else {
			value, err = s.array()
		}
	default:
		s.addToToken(typ)
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

func (s *RESP2Scanner) nullBulkString() (Value, error) {
	if err := s.nullSuffix(); err != nil {
		return nil, err
	}

	return RESP2NullBulkString{}, nil
}

func (s *RESP2Scanner) array() (Value, error) {
	length, err := s.unsignedInt()
	if err != nil {
		// TODO: err
	}
	// TODO: error if length is greater than size of int64

	if err := s.consumeCR(); err != nil {
		return nil, err
	}
	if err := s.consumeLF(); err != nil {
		return nil, err
	}

	result := Array{}
	for range length {
		// Recursively scan for the next element.
		element, err := s.Scan()
		if err != nil {
			return nil, err
		}

		result = append(result, element)
	}

	return result, nil
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
	var result uint64
	for {
		digit, err := s.peek()
		if err != nil {
			return 0, err
		}

		if !s.isDigit(digit) {
			// All digits have been processed; return the result.
			return result, nil
		}

		result = (10 * result) + uint64(s.asciiDigitToInt(digit))

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
		digit, err := s.peek()
		if err != nil {
			return 0, err
		}

		if !s.isDigit(digit) {
			// All digits have been processed; return the result.
			if negative {
				result *= -1
			}
			return result, nil
		}

		result = s.appendInt64Digit(result, digit)

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

	digitOrSign, err := s.advance()
	if err != nil {
		return 0, false, err
	}

	switch digitOrSign {
	case '-':
		negative = true
		signSeen = true
	case '+':
		signSeen = true
	default:
		if s.isDigit(digitOrSign) {
			result = s.appendInt64Digit(result, digitOrSign)
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

func (s *RESP2Scanner) appendInt64Digit(result int64, digit byte) int64 {
	return (10 * result) + int64(s.asciiDigitToInt(digit))
}

func (s *RESP2Scanner) consumeCR() error {
	cr, err := s.advance()
	if err != nil {
		return err
	}
	if !s.isCR(cr) {
		return invalidSyntaxError
	}
	return nil
}

func (s *RESP2Scanner) consumeLF() error {
	lf, err := s.advance()
	if err != nil {
		return err
	}
	if !s.isLF(lf) {
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
