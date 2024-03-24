package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
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

		handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer errorHandlingClose(conn)

	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			printErr(err)
		}

		_, err = conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			printErr(err)
		}
	}
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
