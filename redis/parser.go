package redis

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Parser struct{}

func (p Parser) Parse(reader io.Reader) (Command, error) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, reader)
	if err != nil {
		return Command{}, err
	}
	bufReader := bufio.NewReader(&buf)

	b, err := bufReader.ReadByte()
	if err != nil {
		return Command{}, err
	}
	switch b {
	case '*':
		return processRequest(bufReader)
	default:
		return Command{}, errors.New("unrecognized command")
	}
}

func processRequest(connReader *bufio.Reader) (Command, error) {
	array, err := readArray(connReader)
	if err != nil {
		return Command{}, err
	}

	if len(array) == 0 {
		return Command{}, errors.New("incomplete request")
	}

	switch {
	case strings.EqualFold(array[0], "PING"):
		return Command{
			typ: Ping,
			f:   processPingRequest,
		}, nil
	case strings.EqualFold(array[0], "ECHO"):
		return Command{
			typ: Echo,
			f: func(w io.Writer) error {
				return processEchoRequest(array, w)
			},
		}, nil
	}
	return Command{}, errors.New("unrecognized command")
}

func processPingRequest(w io.Writer) error {
	_, err := w.Write([]byte("+PONG\r\n"))
	return err
}

func processEchoRequest(array []string, w io.Writer) error {
	if len(array) != 2 {
		return errors.New("ECHO command expected to have one argument")
	}

	echo := array[1]
	response := fmt.Sprintf("$%d\r\n%s\r\n", len(echo), echo)
	_, err := w.Write([]byte(response))
	return err
}

func readArray(reader *bufio.Reader) ([]string, error) {
	arrayLength, err := readUnsignedInt(reader)
	if err != nil {
		return nil, err
	}

	err = expectCRLF(reader)
	if err != nil {
		return nil, err
	}

	var array []string
	for i := 0; i < arrayLength; i++ {
		elem, err := readBulkString(reader)
		if err != nil {
			return nil, err
		}
		array = append(array, elem)
	}

	return array, nil
}

func readBulkString(reader *bufio.Reader) (string, error) {
	err := expect(reader, '$')
	if err != nil {
		return "", err
	}

	stringLength, err := readUnsignedInt(reader)
	if err != nil {
		return "", err
	}

	err = expectCRLF(reader)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	for i := 0; i < stringLength; i++ {
		b, err := reader.ReadByte()
		if err != nil {
			return "", err
		}
		builder.WriteByte(b)
	}

	err = expectCRLF(reader)
	if err != nil {
		return "", err
	}

	return builder.String(), nil
}

func readUnsignedInt(reader *bufio.Reader) (int, error) {
	var buffer bytes.Buffer

	b, err := readDigit(reader)
	if err != nil {
		// There were no digits at all, so return early
		return 0, err
	}

	_ = buffer.WriteByte(b)

	for {
		b, err := readDigit(reader)
		if err != nil {
			// At least one digit was read, so stop
			break
		}

		_ = buffer.WriteByte(b)
	}

	return strconv.Atoi(buffer.String())
}

func readDigit(reader *bufio.Reader) (byte, error) {
	bs, err := reader.Peek(1)
	if err != nil {
		return 0, err
	}
	if !('0' <= bs[0] && bs[0] <= '9') {
		return 0, errors.New("expected digit")
	}
	b, err := reader.ReadByte()
	if err != nil {
		return 0, err
	}
	return b, nil
}

func expectCRLF(reader *bufio.Reader) error {
	err := expect(reader, '\r')
	if err != nil {
		return err
	}
	err = expect(reader, '\n')
	if err != nil {
		return err
	}
	return nil
}

func expect(reader *bufio.Reader, b byte) error {
	readBytes, err := reader.Peek(1)
	if err != nil {
		return err
	}

	if readBytes[0] != b {
		return fmt.Errorf("expected %v but was %v", b, readBytes[0])
	}

	_, err = reader.ReadByte()
	return err
}
