package redis

//region Interfaces and type hierarchy

type Value interface {
	// isValue is a private method that enforces that only types in this
	// package can implement Value.
	isValue()
}

type InputValue interface {
	// isInputValue is a private method that enforces that only types in this
	// package can implement InputValue.
	isInputValue()
}

type ErrorValue interface {
	error
	Value

	Message() string
}

var (
	_ Value = (*SimpleString)(nil)
	_ Value = (*SimpleError)(nil)
	_ Value = (*Integer)(nil)
	_ Value = (*BulkString)(nil)
	_ Value = (*NullBulkString)(nil)
	_ Value = (*Array)(nil)
	_ Value = (*NullArray)(nil)
	_ Value = (*BulkError)(nil)

	_ InputValue = (*SimpleString)(nil)
	_ InputValue = (*BulkStringArray)(nil)

	_ ErrorValue = (*SimpleError)(nil)
	_ ErrorValue = (*BulkError)(nil)
)

//endregion

//region Structs

type SimpleString string

// isValue implements Value.
func (s SimpleString) isValue() {
	unreachable[any]()
}

// isInputValue implements InputValue.
func (s SimpleString) isInputValue() {
	unreachable[any]()
}

type SimpleError struct {
	message string
}

func MakeSimpleError(message string) SimpleError {
	return SimpleError{
		message: message,
	}
}

// isValue implements Value.
func (e SimpleError) isValue() {
	unreachable[any]()
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
	unreachable[any]()
}

type BulkString string

// isValue implements Value.
func (s BulkString) isValue() {
	unreachable[any]()
}

type NullBulkString struct{}

// isValue implements Value.
func (s NullBulkString) isValue() {
	unreachable[any]()
}

type Array []Value

// isValue implements Value.
func (a Array) isValue() {
	unreachable[any]()
}

type BulkStringArray []BulkString

// isInputValue implements InputValue.
func (a BulkStringArray) isInputValue() {
	unreachable[any]()
}

type NullArray struct{}

// isValue implements Value.
func (s NullArray) isValue() {
	unreachable[any]()
}

type BulkError struct {
	message string
}

func MakeBulkError(message string) BulkError {
	return BulkError{
		message: message,
	}
}

// isValue implements Value.
func (e BulkError) isValue() {
	unreachable[any]()
}

// Error implements error.
func (e BulkError) Error() string {
	return e.message
}

// Message implements ErrorValue.
func (e BulkError) Message() string {
	return e.message
}

//endregion
