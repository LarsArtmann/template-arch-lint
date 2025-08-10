package shared

import (
	"errors"
	"testing"
)

func TestResult(t *testing.T) {
	// Test successful result
	result := NewResult("test value")

	if !result.IsSuccess() {
		t.Error("Expected result to be successful")
	}

	if result.IsError() {
		t.Error("Expected result not to be error")
	}

	value, err := result.Value()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if value != "test value" {
		t.Errorf("Expected 'test value', got %s", value)
	}

	if result.Unwrap() != "test value" {
		t.Errorf("Expected 'test value', got %s", result.Unwrap())
	}

	if result.UnwrapOr("default") != "test value" {
		t.Errorf("Expected 'test value', got %s", result.UnwrapOr("default"))
	}
}

func TestResultError(t *testing.T) {
	// Test error result
	testErr := errors.New("test error")
	result := NewError[string](testErr)

	if result.IsSuccess() {
		t.Error("Expected result not to be successful")
	}

	if !result.IsError() {
		t.Error("Expected result to be error")
	}

	_, err := result.Value()
	if err != testErr {
		t.Errorf("Expected error %v, got %v", testErr, err)
	}

	if result.UnwrapOr("default") != "default" {
		t.Errorf("Expected 'default', got %s", result.UnwrapOr("default"))
	}
}

func TestResultMap(t *testing.T) {
	// Test mapping successful result
	result := NewResult(10)
	mapped := result.Map(func(x int) int { return x * 2 })

	if !mapped.IsSuccess() {
		t.Error("Expected mapped result to be successful")
	}

	if mapped.Unwrap() != 20 {
		t.Errorf("Expected 20, got %d", mapped.Unwrap())
	}

	// Test mapping error result
	errorResult := NewError[int](errors.New("test"))
	mappedError := errorResult.Map(func(x int) int { return x * 2 })

	if !mappedError.IsError() {
		t.Error("Expected mapped error result to be error")
	}
}

func TestOption(t *testing.T) {
	// Test Some option
	option := Some("test value")

	if !option.IsSome() {
		t.Error("Expected option to have some value")
	}

	if option.IsNone() {
		t.Error("Expected option not to be none")
	}

	if option.Unwrap() != "test value" {
		t.Errorf("Expected 'test value', got %s", option.Unwrap())
	}

	if option.UnwrapOr("default") != "test value" {
		t.Errorf("Expected 'test value', got %s", option.UnwrapOr("default"))
	}
}

func TestOptionNone(t *testing.T) {
	// Test None option
	option := None[string]()

	if option.IsSome() {
		t.Error("Expected option not to have value")
	}

	if !option.IsNone() {
		t.Error("Expected option to be none")
	}

	if option.UnwrapOr("default") != "default" {
		t.Errorf("Expected 'default', got %s", option.UnwrapOr("default"))
	}
}

func TestOptionMap(t *testing.T) {
	// Test mapping some option
	option := Some(10)
	mapped := option.Map(func(x int) int { return x * 2 })

	if !mapped.IsSome() {
		t.Error("Expected mapped option to have value")
	}

	if mapped.Unwrap() != 20 {
		t.Errorf("Expected 20, got %d", mapped.Unwrap())
	}

	// Test mapping none option
	noneOption := None[int]()
	mappedNone := noneOption.Map(func(x int) int { return x * 2 })

	if !mappedNone.IsNone() {
		t.Error("Expected mapped none option to be none")
	}
}

func TestOptionFilter(t *testing.T) {
	// Test filtering some option with passing predicate
	option := Some(10)
	filtered := option.Filter(func(x int) bool { return x > 5 })

	if !filtered.IsSome() {
		t.Error("Expected filtered option to have value")
	}

	if filtered.Unwrap() != 10 {
		t.Errorf("Expected 10, got %d", filtered.Unwrap())
	}

	// Test filtering some option with failing predicate
	failFiltered := option.Filter(func(x int) bool { return x < 5 })

	if !failFiltered.IsNone() {
		t.Error("Expected filtered option to be none")
	}

	// Test filtering none option
	noneOption := None[int]()
	noneFiltered := noneOption.Filter(func(x int) bool { return x > 5 })

	if !noneFiltered.IsNone() {
		t.Error("Expected filtered none option to be none")
	}
}
