package redis

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

var (
	syntaxMissingCRLFError     = NewSimpleError("SYNTAX missing CRLF")
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
	value, err := s.simpleString()
	s.resetToken()
	return value, err
}

func (s *RESP2Scanner) simpleString() (Value, error) {
	for {
		b, err := s.advance()
		if errors.Is(err, io.EOF) && s.midReadingToken() {
			return nil, syntaxMissingCRLFError
		}
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

func (s *RESP2Scanner) consumeLF() error {
	b, err := s.advance()
	if errors.Is(err, io.EOF) {
		return syntaxMissingCRLFError
	}
	if err != nil {
		return fmt.Errorf("%v: %v", internalScannerSimpleError, err)
	}
	if !s.isLF(b) {
		return syntaxMissingCRLFError
	}
	return nil
}

func (s *RESP2Scanner) advance() (byte, error) {
	b, err := s.reader.ReadByte()
	if errors.Is(err, io.EOF) {
		return 0, err
	}
	if err != nil {
		return 0, fmt.Errorf("%v: %v", internalScannerSimpleError, err)
	}

	return b, nil
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

func (s *RESP2Scanner) addToToken(b byte) {
	s.buf = append(s.buf, b)
}

func (s *RESP2Scanner) midReadingToken() bool {
	return len(s.buf) != 0
}

func (s *RESP2Scanner) token() []byte {
	return s.buf
}

func (s *RESP2Scanner) resetToken() {
	s.buf = make([]byte, 0, 16)
}
