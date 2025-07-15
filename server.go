package main

import (
	"github.com/codecrafters-io/redis-starter-go/redis"
)

func main() {
	(&redis.Server{}).Run()
}
