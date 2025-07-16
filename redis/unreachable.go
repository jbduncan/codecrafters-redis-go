package redis

func unreachable[T any]() T {
	panic("unreachable")
	//goland:noinspection GoUnreachableCode
	var t T
	return t
}
