package redis

type PingHandler struct{}

func (h PingHandler) Command() string {
	return "PING"
}

func (h PingHandler) Handle(args Array[BulkString]) Value {
	switch len(args) {
	case 0:
		return SimpleString("PONG")
	case 1:
		return args[0]
	default:
		return wrongNumberOfArgumentsError(h)
	}
}
