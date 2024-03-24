package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	const redisPort = 6379

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", redisPort))
	if err != nil {
		printErr(err)
		os.Exit(1)
	}

	defer errorHandlingClose(listener)

	fmt.Println("Server is listening on port 6379")

	for {
		conn, err := listener.Accept()
		if err != nil {
			printErr(err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer errorHandlingClose(conn)

	connReader := bufio.NewReader(conn)

	for {
		b, err := connReader.ReadByte()
		if err != nil {
			printErr(err)
			break
		}
		switch b {
		case '*':
			err = processRequest(connReader, conn)
			break
		default:
			err = errors.New("unrecognized command")
		}
		if err != nil {
			printErr(err)
			// TODO: Respond with Redis error message
		}
	}
}

func processRequest(connReader *bufio.Reader, connWriter io.Writer) error {
	array, err := readArray(connReader)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", array)

	if len(array) == 0 {
		return errors.New("incomplete request")
	}

	switch {
	case strings.EqualFold(array[0], "PING"):
		return processPingRequest(err, connWriter)
	case strings.EqualFold(array[0], "ECHO"):
		return processEchoRequest(array, err, connWriter)
	}
	return errors.New("unrecognized command")
}

func processPingRequest(err error, connWriter io.Writer) error {
	_, err = connWriter.Write([]byte("+PONG\r\n"))
	return err
}

func processEchoRequest(array []string, err error, connWriter io.Writer) error {
	if len(array) != 2 {
		return errors.New("ECHO command expected to have one argument")
	}

	echo := array[1]
	response := fmt.Sprintf("$%d\r\n%s\r\n", len(echo), echo)
	_, err = connWriter.Write([]byte(response))
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

func printErr(err error) {
	_, _ = fmt.Printf("Error: %v\n", err)
}

func errorHandlingClose(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		printErr(err)
	}
}
