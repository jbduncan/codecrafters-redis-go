package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

const defaultRedisPort = 6379

var port = flag.Int("port", defaultRedisPort, "the port to run the Redis server on")

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		printErr(err)
		os.Exit(1)
	}
	defer errorHandlingClose(listener)
	fmt.Printf("Server is listening on port %d\n", *port)

	store := redis.NewStore()
	clock := redis.RealClock{}
	redisParser := redis.NewParser(store, clock)

	for {
		conn, err := listener.Accept()
		if err != nil {
			printErr(err)
			continue
		}

		go handleConn(conn, redisParser)
	}
}

func handleConn(conn net.Conn, redisParser redis.Parser) {
	defer errorHandlingClose(conn)

	for {
		command, err := redisParser.Parse(conn)
		if err != nil {
			if err == io.EOF {
				return
			}
			printErr(err)
			return
		}

		response := command.Run()
		_, err = conn.Write([]byte(response))
		if err != nil {
			printErr(err)
			return
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
