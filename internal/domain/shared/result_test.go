package shared

import (
	"errors"
	"testing"
)

const testValue = "test value"

func TestResult(t *testing.T) {
	// Test successful result
	result := Ok(testValue)

	if !result.IsOk() {
		t.Error("Expected result to be successful")
	}

	if result.IsError() {
		t.Error("Expected result not to be error")
	}

	value := result.MustGet()
	if value != testValue {
		t.Errorf("Expected %q, got %s", testValue, value)
	}

	if result.OrElse("default") != testValue {
		t.Errorf("Expected %q, got %s", testValue, result.OrElse("default"))
	}
}

func TestResultError(t *testing.T) {
	// Test error result
	testErr := errors.New("test error")
	result := Err[string](testErr)

	if result.IsOk() {
		t.Error("Expected result not to be successful")
	}

	if !result.IsError() {
		t.Error("Expected result to be error")
	}

	err := result.Error()
	if !errors.Is(err, testErr) {
		t.Errorf("Expected error %v, got %v", testErr, err)
	}

	if result.OrElse("default") != "default" {
		t.Errorf("Expected 'default', got %s", result.OrElse("default"))
	}
}

func TestOption(t *testing.T) {
	// Test Some option
	option := Some(testValue)

	if !option.IsPresent() {
		t.Error("Expected option to have some value")
	}

	if option.IsAbsent() {
		t.Error("Expected option not to be none")
	}

	if option.MustGet() != testValue {
		t.Errorf("Expected %q, got %s", testValue, option.MustGet())
	}

	if option.OrElse("default") != testValue {
		t.Errorf("Expected %q, got %s", testValue, option.OrElse("default"))
	}
}

func TestOptionNone(t *testing.T) {
	// Test None option
	option := None[string]()

	if option.IsPresent() {
		t.Error("Expected option not to have value")
	}

	if !option.IsAbsent() {
		t.Error("Expected option to be none")
	}

	if option.OrElse("default") != "default" {
		t.Errorf("Expected 'default', got %s", option.OrElse("default"))
	}
}

func TestEither(t *testing.T) {
	// Test Right either
	rightEither := Right[string, int](42)
	
	if !rightEither.IsRight() {
		t.Error("Expected either to be right")
	}
	
	if rightEither.IsLeft() {
		t.Error("Expected either not to be left")
	}
	
	if rightEither.MustRight() != 42 {
		t.Errorf("Expected 42, got %d", rightEither.MustRight())
	}

	// Test Left either 
	leftEither := Left[string, int]("error")
	
	if !leftEither.IsLeft() {
		t.Error("Expected either to be left")
	}
	
	if leftEither.IsRight() {
		t.Error("Expected either not to be right")
	}
	
	if leftEither.MustLeft() != "error" {
		t.Errorf("Expected 'error', got %s", leftEither.MustLeft())
	}
}