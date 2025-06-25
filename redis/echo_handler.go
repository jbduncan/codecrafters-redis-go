package redis

type EchoHandler struct{}

func (h EchoHandler) Command() string {
	return "ECHO"
}

func (h EchoHandler) Handle(args Array[BulkString]) Value {
	switch len(args) {
	case 1:
		return args[0]
	default:
		return wrongNumberOfArgumentsError(h)
	}
}
