package redistest

// TODO: Consider replacing these consts with helper functions that use
//       RESP2Writer behind the scenes, e.g.,
//   - bstrs(...string) -> string repr of array of bulk strings
//   - sstr(string) -> string repr of single simple string
const (
	RequestPingLowercase     = "*1\r\n$4\r\nping\r\n"
	RequestPingUppercase     = "*1\r\n$4\r\nPING\r\n"
	RequestEchoLowercaseFoo  = "*2\r\n$4\r\necho\r\n$3\r\nfoo\r\n"
	RequestEchoUppercaseQuux = "*2\r\n$4\r\nECHO\r\n$4\r\nquux\r\n"
	RequestSetLowercaseFoo   = "*2\r\n$3\r\nset\r\n$3\r\nfoo\r\n"
	RequestGetLowercaseFoo   = "*2\r\n$3\r\nget\r\n$3\r\nfoo\r\n"
	RequestInvalidSyntax     = "*1\r\n^"

	ResponsePong               = "+PONG\r\n"
	ResponseFoo                = "$3\r\nfoo\r\n"
	ResponseQuux               = "$4\r\nquux\r\n"
	ResponseOK                 = "+OK\r\n"
	ResponseInvalidSyntaxError = "!41\r\nERR Protocol error: expected '$', got '^'\r\n"
)
