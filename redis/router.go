package redis

type Router struct {
	handlers []Handler
}

// TODO: Route method with similar signature to Handler.Handle that determines
//       which Handler to call based on "args[0] == matchingHandler.Command()"
//       and forwards args[1:] onto that Handler.
// TODO: Route method: return BulkError for unrecognised command.
// TODO: Route method: return BulkError when given zero args.

func NewRouter(handlers []Handler) *Router {
	return &Router{
		handlers: handlers,
	}
}

func (r Router) Route(value Value) Value {
	if _, ok := value.(SimpleString); ok {
		// TODO: check that value is "PING", otherwise return error
		// TODO: handle 2 handlers
		// TODO: handle 0 handlers
		return r.handlers[0].Handle(Array[BulkString]{})
	}

	// TODO: test that result of r.handlers[0].Handle is returned
	// TODO: gracefully handle values which aren't Array[BulkString]s
	// TODO: gracefully handle arrays of len 0
	// TODO: handle 2 handlers
	// TODO: handle 0 handlers
	r.handlers[0].Handle(value.(Array[BulkString])[1:])
	return Array[BulkString]{"Hello, world"}
}
