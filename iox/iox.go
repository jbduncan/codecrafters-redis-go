package iox

import (
	"io"
)

func ReadExactly(r io.Reader, numBytes int) (string, error) {
	buf := make([]byte, numBytes)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}
