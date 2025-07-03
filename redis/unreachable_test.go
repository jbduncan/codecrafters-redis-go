package redis_test

func unreachable[T any]() T {
	panic("unreachable")
	//goland:noinspection GoUnreachableCode
	var t T
	return t
}
