package redis_test

func ptr[T any](value T) *T {
	return &value
}
