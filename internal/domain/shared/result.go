// Package shared provides Result/Option patterns for safer error handling
package shared

// Result represents a value that might fail
type Result[T any] struct {
	value T
	err   error
}

// NewResult creates a successful result
func NewResult[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// NewError creates a failed result
func NewError[T any](err error) Result[T] {
	var zero T
	return Result[T]{value: zero, err: err}
}

// IsSuccess returns true if the result contains a value
func (r Result[T]) IsSuccess() bool {
	return r.err == nil
}

// IsError returns true if the result contains an error
func (r Result[T]) IsError() bool {
	return r.err != nil
}

// Value returns the value and error
func (r Result[T]) Value() (T, error) {
	return r.value, r.err
}

// Unwrap returns the value, panicking if there's an error
func (r Result[T]) Unwrap() T {
	if r.err != nil {
		panic(r.err)
	}
	return r.value
}

// UnwrapOr returns the value or the default if there's an error
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.err != nil {
		return defaultValue
	}
	return r.value
}

// Map transforms the value if the result is successful
func (r Result[T]) Map(f func(T) T) Result[T] {
	if r.err != nil {
		return r
	}
	return NewResult(f(r.value))
}

// MapErr transforms the error if the result failed
func (r Result[T]) MapErr(f func(error) error) Result[T] {
	if r.err == nil {
		return r
	}
	return NewError[T](f(r.err))
}

// Option represents a value that might be absent
type Option[T any] struct {
	value T
	some  bool
}

// Some creates an option with a value
func Some[T any](value T) Option[T] {
	return Option[T]{value: value, some: true}
}

// None creates an empty option
func None[T any]() Option[T] {
	var zero T
	return Option[T]{value: zero, some: false}
}

// IsSome returns true if the option contains a value
func (o Option[T]) IsSome() bool {
	return o.some
}

// IsNone returns true if the option is empty
func (o Option[T]) IsNone() bool {
	return !o.some
}

// Unwrap returns the value, panicking if empty
func (o Option[T]) Unwrap() T {
	if !o.some {
		panic("called Unwrap on None option")
	}
	return o.value
}

// UnwrapOr returns the value or the default if empty
func (o Option[T]) UnwrapOr(defaultValue T) T {
	if !o.some {
		return defaultValue
	}
	return o.value
}

// Map transforms the value if the option has one
func (o Option[T]) Map(f func(T) T) Option[T] {
	if !o.some {
		return o
	}
	return Some(f(o.value))
}

// Filter returns None if the predicate fails
func (o Option[T]) Filter(f func(T) bool) Option[T] {
	if !o.some || !f(o.value) {
		return None[T]()
	}
	return o
}