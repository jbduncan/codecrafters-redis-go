package redis

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Parser struct{}

func (p Parser) Parse(reader io.Reader) (Command, error) {
	var requestBuffer bytes.Buffer
	bufReader := bufio.NewReader(io.TeeReader(reader, &requestBuffer))

	bs, err := bufReader.Peek(1)
	if err != nil {
		return nil, err
	}
	switch bs[0] {
	case '*':
		return p.processArrayRequest(bufReader, requestBuffer)
	default:
		return nil, fmt.Errorf("parse: unrecognized command: %#v", requestBuffer.String())
	}
}

func (p Parser) processArrayRequest(bufReader *bufio.Reader, requestBuffer bytes.Buffer) (Command, error) {
	array, err := readArray(bufReader)
	if err != nil {
		return nil, err
	}

	if len(array) == 0 {
		return nil, fmt.Errorf("parse: incomplete array request: %#v", requestBuffer.String())
	}

	switch {
	case strings.EqualFold(array[0], "PING"):
		return PingCommand, nil
	case strings.EqualFold(array[0], "ECHO"):
		return EchoCommand(array[1]), nil
	}
	return nil, fmt.Errorf("parse: unrecognized command: %#v", requestBuffer.String())
}

func readArray(reader *bufio.Reader) ([]string, error) {
	err := expect(reader, '*')
	if err != nil {
		return nil, err
	}

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
		// No digits were found, so return an error
		return 0, err
	}

	_ = buffer.WriteByte(b)

	for {
		b, err := readDigit(reader)
		if err != nil {
			// At least one digit was found, so return them all
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
		return 0, fmt.Errorf("parse: expected digit in range 0-9 but was byte %v", bs[0])
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
		return fmt.Errorf("parse: expected byte %v but was byte %v", b, readBytes[0])
	}

	_, err = reader.ReadByte()
	return err
}
