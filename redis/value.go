package redis

type Value interface {
	// isValue is a private method that enforces that only types in this
	// package can implement Value.
	isValue()
}

type ErrorValue interface {
	Value

	Message() string
}

type SimpleString string

// isValue implements Value
func (s SimpleString) isValue() {
	panic("do not call me")
}

type SimpleError struct {
	message string
}

func NewSimpleError(message string) SimpleError {
	return SimpleError{
		message: message,
	}
}

// isValue implements Value.
func (e SimpleError) isValue() {
	panic("do not call me")
}

// Error implements error.
func (e SimpleError) Error() string {
	return e.message
}

func (e SimpleError) Message() string {
	return e.message
}
