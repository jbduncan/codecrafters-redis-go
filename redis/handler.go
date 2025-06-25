package redis

type Handler interface {
	Command() string
	Handle(args Array[BulkString]) Value
}
