package main

import (
	cryptorand "crypto/rand"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

const (
	defaultRedisPort  = 6379
	replicaofFlagName = "replicaof"
)

var (
	port      uint64
	replicaOf *replicaOfFlag
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
		return fmt.Errorf("%s must be in the format '<host> <port>'", replicaofFlagName)
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
	flag.Func(
		replicaofFlagName,
		"the Redis server that this server is a replica of; "+
			"must be in the format '<host> <port>'",
		func(h string) error {
			p, err := replicaofPortValue()
			if err != nil {
				return err
			}

			replicaOf = &replicaOfFlag{
				host: h,
				port: p,
			}
			return nil
		})
	flag.Parse()

	var replicationMasterConfig *redis.ReplicationMasterConfig
	if replicaOf == nil {
		replicationMasterConfig = &redis.ReplicationMasterConfig{
			ReplID:     randomReplID(),
			ReplOffset: 0,
		}
	}

	config := &redis.Config{
		Replication: redis.ReplicationConfig{
			Master: replicationMasterConfig,
		},
	}
	store := redis.NewStore()
	clock := redis.RealClock{}
	redisParser := redis.NewParser(config, store, clock)

	startTCPServer(redisParser)
}

func replicaofPortValue() (uint64, error) {
	replicaofFlagIndex := slices.IndexFunc(os.Args, isReplicaofFlag)
	if len(os.Args) <= replicaofFlagIndex+2 {
		return 0, fmt.Errorf("%s must be in the format '<host> <port>'", replicaofFlagName)
	}
	if isFlagThatIsNotReplicaof(os.Args[replicaofFlagIndex+2]) {
		return 0, fmt.Errorf("%s must be in the format '<host> <port>'", replicaofFlagName)
	}
	port, err := strconv.ParseUint(os.Args[replicaofFlagIndex+2], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s port must be a non-negative integer: %w", replicaofFlagName, err)
	}
	return port, nil
}

func isFlagThatIsNotReplicaof(s string) bool {
	result := false
	flag.VisitAll(func(f *flag.Flag) {
		if f.Name != replicaofFlagName {
			if s == "-"+f.Name || s == "--"+f.Name {
				result = true
			}
		}
	})
	return result
}

func isReplicaofFlag(s string) bool {
	return s == "-"+replicaofFlagName || s == "--"+replicaofFlagName
}

const alphabet = "0123456789" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"abcdefghijklmnopqrstuvwxyz"

func randomReplID() string {
	result := make([]byte, 40)
	for i := 0; i < 40; i++ {
		num, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			panic(err)
		}
		result[i] = alphabet[num.Int64()]
	}
	return string(result)
}

func startTCPServer(redisParser redis.Parser) {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		printErr(err)
		os.Exit(1)
	}
	defer errorHandlingClose(listener)
	fmt.Printf("Server is listening on port %d\n", port)

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
		_, err = io.WriteString(conn, response)
		if err != nil {
			printErr(err)
			return
		}
	}
}

func errorHandlingClose(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		printErr(err)
	}
}

func printErr(err error) {
	_, _ = fmt.Printf("Error: %v\n", err)
}
