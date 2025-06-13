package redis

type Handler interface {
	Handle(value Value) Value
}

type HandlerFunc func(value Value) Value

func (f HandlerFunc) Handle(value Value) Value {
	return f(value)
}
