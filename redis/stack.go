package redis

type stack []Value

func (s *stack) push(v Value) {
	*s = append(*s, v)
}

func (s *stack) pop() Value {
	n := len(*s) - 1
	result := (*s)[n]
	*s = (*s)[:n]
	return result
}
