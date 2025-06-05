package redis

import (
	"bufio"
	"errors"
	"io"
)

var (
	absentSimpleError          = SimpleError{}
	missingCRLFSimpleError     = NewSimpleError("ERR missing CRLF")
	internalScannerSimpleError = NewSimpleError("ERR internal scanner error")
)

type RESP2Scanner struct {
	reader *bufio.Reader
	buf    []byte
}

func NewRESP2Scanner(r io.Reader) *RESP2Scanner {
	s := &RESP2Scanner{
		reader: bufio.NewReader(r),
	}
	s.resetBuf()
	return s
}

func (s *RESP2Scanner) Scan() Value {
	value := s.simpleString()
	s.resetBuf()
	return value
}

func (s *RESP2Scanner) simpleString() Value {
	for {
		b, err := s.advance()
		if !errors.Is(err, absentSimpleError) {
			return err
		}

		if !s.isCR(b) {
			continue
		}

		if err := s.consumeLF(); !errors.Is(err, absentSimpleError) {
			return err
		}

		return SimpleString(s.buf)
	}
}

func (s *RESP2Scanner) consumeLF() SimpleError {
	b, err := s.advance()
	if !errors.Is(err, absentSimpleError) {
		return err
	}
	if !s.isLF(b) {
		return missingCRLFSimpleError
	}
	return absentSimpleError
}

func (s *RESP2Scanner) advance() (byte, SimpleError) {
	b, err := s.reader.ReadByte()
	if errors.Is(err, io.EOF) {
		return 0, missingCRLFSimpleError
	}
	if err != nil {
		return 0, internalScannerSimpleError
	}

	if !s.isCRLF(b) {
		s.buf = append(s.buf, b)
	}
	return b, absentSimpleError
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

func (s *RESP2Scanner) resetBuf() {
	s.buf = make([]byte, 0, 16)
}
