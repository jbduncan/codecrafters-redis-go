package main

import (
	"fmt"
	"io"
	"log"
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
		// Block until we receive an incoming connection
		conn, err := listener.Accept()
		if err != nil {
			printErr(err)
			continue
		}

		// Handle client connection
		handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer errorHandlingClose(conn)

	// Read data
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		printErr(err)
		return
	}

	log.Printf("Received data %v\n", buf[:n])

	// Respond with a Redis PONG
	_, err = conn.Write([]byte("+PONG\r\n"))
	if err != nil {
		printErr(err)
		return
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
