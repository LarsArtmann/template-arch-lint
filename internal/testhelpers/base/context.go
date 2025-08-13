// Package base provides context management utilities for tests.
// These utilities establish consistent patterns for context creation and lifecycle management.
package base

import (
	"context"
	"time"
)

// ContextManager provides standardized context creation and management for tests.
// This ensures consistent context handling across all test suites.
type ContextManager interface {
	// NewContext creates a fresh context for testing
	NewContext() context.Context

	// NewContextWithTimeout creates a context with specified timeout
	NewContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc)

	// NewContextWithDeadline creates a context with specified deadline
	NewContextWithDeadline(deadline time.Time) (context.Context, context.CancelFunc)

	// NewContextWithValue creates a context with a key-value pair
	NewContextWithValue(key, value any) context.Context
}

// StandardContextManager implements standard context management for tests.
type StandardContextManager struct {
	baseContext context.Context
}

// NewStandardContextManager creates a new context manager.
func NewStandardContextManager() *StandardContextManager {
	return &StandardContextManager{
		baseContext: context.Background(),
	}
}

// NewContext creates a fresh context for testing.
// This is the standard context used in most test scenarios.
func (cm *StandardContextManager) NewContext() context.Context {
	return context.Background()
}

// NewContextWithTimeout creates a context with specified timeout.
// This is useful for testing timeout scenarios and long-running operations.
func (cm *StandardContextManager) NewContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(cm.baseContext, timeout)
}

// NewContextWithDeadline creates a context with specified deadline.
// This is useful for testing deadline-based cancellation.
func (cm *StandardContextManager) NewContextWithDeadline(deadline time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(cm.baseContext, deadline)
}

// NewContextWithValue creates a context with a key-value pair.
// This is useful for testing context value propagation.
func (cm *StandardContextManager) NewContextWithValue(key, value any) context.Context {
	return context.WithValue(cm.baseContext, key, value)
}

// TestContextBuilder provides a fluent API for building test contexts with various configurations.
type TestContextBuilder struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewTestContextBuilder creates a new context builder starting with background context.
func NewTestContextBuilder() *TestContextBuilder {
	return &TestContextBuilder{
		ctx: context.Background(),
	}
}

// WithTimeout adds a timeout to the context.
func (b *TestContextBuilder) WithTimeout(timeout time.Duration) *TestContextBuilder {
	ctx, cancel := context.WithTimeout(b.ctx, timeout)
	if b.cancelFunc != nil {
		b.cancelFunc() // Cancel previous context if it exists
	}
	b.ctx = ctx
	b.cancelFunc = cancel
	return b
}

// WithDeadline adds a deadline to the context.
func (b *TestContextBuilder) WithDeadline(deadline time.Time) *TestContextBuilder {
	ctx, cancel := context.WithDeadline(b.ctx, deadline)
	if b.cancelFunc != nil {
		b.cancelFunc()
	}
	b.ctx = ctx
	b.cancelFunc = cancel
	return b
}

// WithValue adds a key-value pair to the context.
func (b *TestContextBuilder) WithValue(key, value any) *TestContextBuilder {
	b.ctx = context.WithValue(b.ctx, key, value)
	return b
}

// WithCancel makes the context cancellable.
func (b *TestContextBuilder) WithCancel() *TestContextBuilder {
	ctx, cancel := context.WithCancel(b.ctx)
	if b.cancelFunc != nil {
		b.cancelFunc()
	}
	b.ctx = ctx
	b.cancelFunc = cancel
	return b
}

// Build returns the configured context and optional cancel function.
func (b *TestContextBuilder) Build() (context.Context, context.CancelFunc) {
	return b.ctx, b.cancelFunc
}

// BuildContext returns only the configured context.
func (b *TestContextBuilder) BuildContext() context.Context {
	return b.ctx
}

// Cleanup cancels the context if a cancel function exists.
func (b *TestContextBuilder) Cleanup() {
	if b.cancelFunc != nil {
		b.cancelFunc()
	}
}

// Common context patterns used in tests

// NewTestContext creates a standard test context.
// This is the most common context used across tests.
func NewTestContext() context.Context {
	return context.Background()
}

// NewTestContextWithTimeout creates a test context with a reasonable timeout for tests.
func NewTestContextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// NewTestContextWithShortTimeout creates a test context with a short timeout for timeout testing.
func NewTestContextWithShortTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 100*time.Millisecond)
}

// TestUserIDContextKey is a context key for user ID in tests.
type TestUserIDContextKey struct{}

// TestRequestIDContextKey is a context key for request ID in tests.
type TestRequestIDContextKey struct{}

// NewTestContextWithUserID creates a context with a user ID for authorization testing.
func NewTestContextWithUserID(userID string) context.Context {
	return context.WithValue(context.Background(), TestUserIDContextKey{}, userID)
}

// NewTestContextWithRequestID creates a context with a request ID for tracing testing.
func NewTestContextWithRequestID(requestID string) context.Context {
	return context.WithValue(context.Background(), TestRequestIDContextKey{}, requestID)
}

// GetUserIDFromTestContext extracts user ID from test context.
func GetUserIDFromTestContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(TestUserIDContextKey{}).(string)
	return userID, ok
}

// GetRequestIDFromTestContext extracts request ID from test context.
func GetRequestIDFromTestContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(TestRequestIDContextKey{}).(string)
	return requestID, ok
}
