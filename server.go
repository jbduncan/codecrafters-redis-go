package main

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

func main() {
	if err := (&redis.Server{}).Run(os.Stderr); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
