package shared

import (
	"errors"
	"testing"
)

func TestResult(t *testing.T) {
	// Test successful result
	result := Ok("test value")

	if !result.IsOk() {
		t.Error("Expected result to be successful")
	}

	if result.IsError() {
		t.Error("Expected result not to be error")
	}

	value := result.MustGet()
	if value != "test value" {
		t.Errorf("Expected 'test value', got %s", value)
	}

	if result.OrElse("default") != "test value" {
		t.Errorf("Expected 'test value', got %s", result.OrElse("default"))
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
	if err != testErr {
		t.Errorf("Expected error %v, got %v", testErr, err)
	}

	if result.OrElse("default") != "default" {
		t.Errorf("Expected 'default', got %s", result.OrElse("default"))
	}
}

func TestOption(t *testing.T) {
	// Test Some option
	option := Some("test value")

	if !option.IsPresent() {
		t.Error("Expected option to have some value")
	}

	if option.IsAbsent() {
		t.Error("Expected option not to be none")
	}

	if option.MustGet() != "test value" {
		t.Errorf("Expected 'test value', got %s", option.MustGet())
	}

	if option.OrElse("default") != "test value" {
		t.Errorf("Expected 'test value', got %s", option.OrElse("default"))
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