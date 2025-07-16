package redis

type EchoHandler struct{}

func (h EchoHandler) Handle(args BulkStringArray) Value {
	switch len(args) {
	case 1:
		return args[0]
	default:
		return wrongNumberOfArgumentsError("echo")
	}
}
