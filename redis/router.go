package redis

type Router struct{}

// TODO: initialization function with Handler arg.
// TODO: Route method with similar signature to Handler.Handle that determines
//       which Handler to call based on "args[0] == matchingHandler.Command()"
//       and forwards args[1:] onto that Handler.
// TODO: Route method: return BulkError for unrecognised command.
// TODO: Route method: return BulkError when given zero args.
