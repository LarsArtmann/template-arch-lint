// Package shared provides functional programming patterns using samber/mo
package shared

import (
	"github.com/samber/mo"
)

// Option provides a nullable value wrapper using mo.Option.
type Option[T any] = mo.Option[T]

// Some creates an option with a value.
func Some[T any](value T) Option[T] {
	return mo.Some(value)
}

// None creates an empty option.
func None[T any]() Option[T] {
	return mo.None[T]()
}

// Result provides error handling with values using mo.Result.
type Result[T any] = mo.Result[T]

// Ok creates a successful result.
func Ok[T any](value T) Result[T] {
	return mo.Ok(value)
}

// Err creates a failed result.
func Err[T any](err error) Result[T] {
	return mo.Err[T](err)
}

// Either provides dual-type return values using mo.Either.
type Either[L, R any] = mo.Either[L, R]

// Left creates an Either with left value (typically error case).
func Left[L, R any](value L) Either[L, R] {
	return mo.Left[L, R](value)
}

// Right creates an Either with right value (typically success case).
func Right[L, R any](value R) Either[L, R] {
	return mo.Right[L, R](value)
}

// Backward compatibility aliases for existing code
// These will be removed in a future refactor

// NewResult creates a successful result (deprecated - use Ok).
func NewResult[T any](value T) Result[T] {
	return Ok(value)
}

// NewError creates a failed result (deprecated - use Err).
func NewError[T any](err error) Result[T] {
	return Err[T](err)
}
