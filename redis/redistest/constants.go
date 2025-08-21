package redistest

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

var (
	RequestPingLowercase     = bstrs("ping")
	RequestPingUppercase     = bstrs("PING")
	RequestEchoLowercaseFoo  = bstrs("echo", "foo")
	RequestEchoUppercaseQuux = bstrs("ECHO", "quux")
	RequestSetLowercaseFoo   = bstrs("set", "foo")
	RequestGetLowercaseFoo   = bstrs("get", "foo")
	RequestInvalidSyntax     = "*1\r\n^"

	ResponsePong               = sstr("PONG")
	ResponseFoo                = bstr("foo")
	ResponseQuux               = bstr("quux")
	ResponseOK                 = sstr("OK")
	ResponseInvalidSyntaxError = berr("ERR Protocol error: expected '$', got '^'")
)

func bstr(value string) string {
	var s strings.Builder
	mustWrite(&s, redis.BulkString(value))
	return s.String()
}

func bstrs(first string, rest ...string) string {
	array := make(redis.Array, 0, len(rest)+1)
	array = append(array, redis.BulkString(first))
	for _, value := range rest {
		array = append(array, redis.BulkString(value))
	}

	var s strings.Builder
	mustWrite(&s, array)
	return s.String()
}

func sstr(value string) string {
	var s strings.Builder
	mustWrite(&s, redis.SimpleString(value))
	return s.String()
}

func berr(value string) string {
	var s strings.Builder
	mustWrite(&s, redis.MakeBulkError(value))
	return s.String()
}

func mustWrite(s *strings.Builder, value redis.Value) {
	if err := redis.NewRESP2Writer(s).Write(value); err != nil {
		panic(fmt.Sprintf("unexpected RESP2 write error: %v", err))
	}
}
