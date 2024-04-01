package redis

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
)

func NewParser(config Config, store *Store, clock Clock) Parser {
	return Parser{
		store: store,
		clock: clock,
	}
}

type Parser struct {
	store *Store
	clock Clock
}

func (p Parser) Parse(reader io.Reader) (Command, error) {
	bufReader := bufio.NewReader(reader)

	bs, err := bufReader.Peek(1)
	if err != nil {
		return nil, err
	}
	switch bs[0] {
	case '*':
		return p.processArrayRequest(bufReader)
	default:
		// TODO: return error that server.go can match on
		panic("unexpected")
	}
}

func (p Parser) processArrayRequest(bufReader *bufio.Reader) (Command, error) {
	array, err := readArray(bufReader)
	if err != nil {
		return nil, err
	}

	if len(array) == 0 {
		// TODO: return error that server.go can match on
	}

	switch {
	case strings.EqualFold(array[0], "ECHO"):
		return p.makeEchoCommand(array)
	case strings.EqualFold(array[0], "GET"):
		return p.newGetCommand(array)
	case strings.EqualFold(array[0], "INFO"):
		return p.makeInfoCommand(array)
	case strings.EqualFold(array[0], "PING"):
		return p.makePingCommand(array)
	case strings.EqualFold(array[0], "SET"):
		return p.newSetCommand(array)
	}
	// TODO: return error that server.go can match on
	panic("unexpected")
}

func (p Parser) newSetCommand(array []string) (Command, error) {
	if len(array) == 5 {
		if !strings.EqualFold(array[3], "PX") {
			// TODO: return error that server.go can match on
		}
		expiryTimeInMilliseconds, err := strconv.Atoi(array[4])
		if err != nil {
			// TODO: return error that server.go can match on
		}
		if expiryTimeInMilliseconds <= 0 {
			// TODO: return error that server.go can match on
		}
		expiryTime := time.Duration(expiryTimeInMilliseconds) * time.Millisecond
		return NewSetCommand(
				p.store,
				array[1],
				array[2],
				ExpiryTime(p.clock.Now().Add(expiryTime)),
			),
			nil
	}
	return NewSetCommand(p.store, array[1], array[2]), nil
}

func (p Parser) makePingCommand(array []string) (Command, error) {
	if len(array) != 1 {
		// TODO: return error that server.go can match on
	}
	return PingCommand{}, nil
}

func (p Parser) makeInfoCommand(array []string) (Command, error) {
	if len(array) != 2 {
		// TODO: return error that server.go can match on
	}
	return InfoCommand(array[1]), nil
}

func (p Parser) newGetCommand(array []string) (Command, error) {
	if len(array) != 2 {
		// TODO: return error that server.go can match on
	}
	return NewGetCommand(p.store, p.clock, array[1]), nil
}

func (p Parser) makeEchoCommand(array []string) (Command, error) {
	if len(array) != 2 {
		// TODO: return error that server.go can match on
	}
	return EchoCommand(array[1]), nil
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

	// TODO: expect the end of the reader

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
	var buffer strings.Builder

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
		// TODO: return error that server.go can match on
		return 0, errors.New("not a digit")
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
		// TODO: return error that server.go can match on
	}

	_, err = reader.ReadByte()
	return err
}
