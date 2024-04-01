package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

const defaultRedisPort = 6379

var (
	port uint64
	//replicaOf *replicaOfFlag
	replicaOf string
)

type replicaOfFlag struct {
	host string
	port uint64
}

func (r *replicaOfFlag) String() string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", *r)
}

func (r *replicaOfFlag) Set(value string) error {
	parts := strings.Split(value, " ")
	if len(parts) != 2 {
		return fmt.Errorf("replicaOf must be in the format '<host> <port>'")
	}

	r.host = parts[0]

	port, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return err
	}
	r.port = port

	return nil
}

func main() {
	flag.Uint64Var(&port, "port", defaultRedisPort, "the port to run the Redis server on")
	//flag.Var(
	//	replicaOf,
	//	"replicaof",
	//	"the Redis server that this server is a replica of; "+
	//		"must be in the format '<host> <port>'",
	//)
	flag.StringVar(&replicaOf, "replicaof", "", "foo")
	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		printErr(err)
		os.Exit(1)
	}
	defer errorHandlingClose(listener)
	fmt.Printf("Server is listening on port %d\n", port)

	var role redis.ReplicationRole
	if replicaOf == "" {
		role = redis.ReplicationRoleMaster
	} else {
		role = redis.ReplicationRoleSlave
	}

	config := &redis.Config{
		Replication: &redis.ReplicationConfig{
			Role: role,
		},
	}
	store := redis.NewStore()
	clock := redis.RealClock{}
	redisParser := redis.NewParser(config, store, clock)

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
