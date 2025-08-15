package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

func main() {
	// TODO: Figure out a way to make the server stop gracefully on a SIGINT or
	//       SIGTERM, allowing TCP connections to finish and to terminate any
	//       slow connections.
	//   - https://victoriametrics.com/blog/go-graceful-shutdown/
	//   - https://www.rudderstack.com/blog/implementing-graceful-shutdown-in-go/
	//   - https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
	//   - Search other resources
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	if _, err := redis.StartServer(os.Stderr); err != nil {
		return err
	}

	blockUntilInterrupted()
	return nil
}

func blockUntilInterrupted() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
