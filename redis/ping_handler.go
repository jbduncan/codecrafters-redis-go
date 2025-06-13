package redis

func PingHandlerFunc(args Array) Value {
	switch len(args) {
	case 0:
		return SimpleString("PONG")
	case 1:
		return args[0]
	default:
		return NewBulkError("-ERR wrong number of arguments for 'echo' command")
	}
}
