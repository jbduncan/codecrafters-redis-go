package redis

type PingHandler struct{}

func (h PingHandler) Handle(args BulkStringArray) Value {
	switch len(args) {
	case 0:
		return SimpleString("PONG")
	case 1:
		return args[0]
	default:
		return wrongNumberOfArgumentsError("ping")
	}
}
