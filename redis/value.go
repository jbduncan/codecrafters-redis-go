package redis

type Value interface {
	// isValue is a private method that enforces that only types in this
	// package can implement Value.
	isValue()
}

type ErrorValue interface {
	error
	Value

	Message() string
}

type SimpleString string

// isValue implements Value.
func (s SimpleString) isValue() {
	unreachable()
}

type SimpleError struct {
	message string
}

// TODO: rename to MakeSimpleError
func NewSimpleError(message string) SimpleError {
	return SimpleError{
		message: message,
	}
}

// isValue implements Value.
func (e SimpleError) isValue() {
	unreachable()
}

// Error implements error.
func (e SimpleError) Error() string {
	return e.message
}

// Message implements ErrorValue.
func (e SimpleError) Message() string {
	return e.message
}

type Integer int64

// isValue implements Value.
func (i Integer) isValue() {
	unreachable()
}

type BulkString string

// isValue implements Value.
func (s BulkString) isValue() {
	unreachable()
}

type RESP2NullBulkString struct{}

// isValue implements Value.
func (s RESP2NullBulkString) isValue() {
	unreachable()
}

type RESP2NullArray struct{}

// isValue implements Value.
func (s RESP2NullArray) isValue() {
	unreachable()
}

type Array[T Value] []T

// isValue implements Value.
func (a Array[T]) isValue() {
	unreachable()
}

type BulkError struct {
	message string
}

// TODO: rename to MakeBulkError
func NewBulkError(message string) BulkError {
	return BulkError{
		message: message,
	}
}

// isValue implements Value.
func (e BulkError) isValue() {
	unreachable()
}

// Error implements error.
func (e BulkError) Error() string {
	return e.message
}

// Message implements ErrorValue.
func (e BulkError) Message() string {
	return e.message
}
