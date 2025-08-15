package errors_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	domainErrors "github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
	"github.com/LarsArtmann/template-arch-lint/internal/utils/errors"
)

func TestErrors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Errors Suite")
}

var _ = Describe("WrappedError", func() {
	var (
		originalErr error
		wrappedErr  *errors.WrappedError
	)

	BeforeEach(func() {
		originalErr = fmt.Errorf("original error")
		wrappedErr = errors.NewWrappedError(originalErr, "wrapped message")
	})

	Describe("NewWrappedError", func() {
		It("should create a wrapped error with message", func() {
			Expect(wrappedErr).ToNot(BeNil())
			Expect(wrappedErr.Error()).To(Equal("wrapped message: original error"))
		})

		It("should return nil for nil error", func() {
			result := errors.NewWrappedError(nil, "message")
			Expect(result).To(BeNil())
		})
	})

	Describe("WithContext", func() {
		It("should extract context values", func() {
			ctx := context.WithValue(context.Background(), errors.CorrelationIDKey, "correlation-123")
			ctx = context.WithValue(ctx, errors.UserIDKey, "user-456")
			ctx = context.WithValue(ctx, errors.RequestIDKey, "request-789")
			ctx = context.WithValue(ctx, errors.OperationKey, "test-operation")

			result := wrappedErr.WithContext(ctx)

			context := result.Context()
			Expect(context.CorrelationID).To(Equal("correlation-123"))
			Expect(context.UserID).To(Equal("user-456"))
			Expect(context.RequestID).To(Equal("request-789"))
			Expect(context.Operation).To(Equal("test-operation"))
		})

		It("should handle nil wrapped error", func() {
			var nilErr *errors.WrappedError
			result := nilErr.WithContext(context.Background())
			Expect(result).To(BeNil())
		})
	})

	Describe("WithOperation", func() {
		It("should set operation information", func() {
			result := wrappedErr.WithOperation("database-query")
			Expect(result.Context().Operation).To(Equal("database-query"))
		})
	})

	Describe("WithExtra", func() {
		It("should add extra information", func() {
			result := wrappedErr.WithExtra("key1", "value1").WithExtra("key2", 42)

			context := result.Context()
			Expect(context.Extra).To(HaveKeyWithValue("key1", "value1"))
			Expect(context.Extra).To(HaveKeyWithValue("key2", 42))
		})
	})

	Describe("WithStackTrace", func() {
		It("should capture stack trace", func() {
			result := wrappedErr.WithStackTrace(0)
			context := result.Context()
			Expect(context.StackTrace).ToNot(BeEmpty())
		})
	})

	Describe("Unwrap", func() {
		It("should return the original error", func() {
			unwrapped := wrappedErr.Unwrap()
			Expect(unwrapped).To(Equal(originalErr))
		})
	})

	Describe("IsDomainError", func() {
		It("should detect domain errors", func() {
			domainErr := domainErrors.NewValidationError("field", "message")
			wrapped := errors.NewWrappedError(domainErr, "wrapped")

			Expect(wrapped.IsDomainError()).To(BeTrue())
		})

		It("should return false for non-domain errors", func() {
			Expect(wrappedErr.IsDomainError()).To(BeFalse())
		})
	})

	Describe("ToDomainError", func() {
		It("should convert to domain error if possible", func() {
			domainErr := domainErrors.NewValidationError("field", "message")
			wrapped := errors.NewWrappedError(domainErr, "wrapped")

			converted, ok := wrapped.ToDomainError()
			Expect(ok).To(BeTrue())
			Expect(converted).To(Equal(domainErr))
		})

		It("should return false for non-domain errors", func() {
			converted, ok := wrappedErr.ToDomainError()
			Expect(ok).To(BeFalse())
			Expect(converted).To(BeNil())
		})
	})

	Describe("ToJSON", func() {
		It("should serialize to JSON", func() {
			wrappedErr = wrappedErr.WithOperation("test").WithExtra("key", "value")

			jsonData, err := wrappedErr.ToJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(jsonData).ToNot(BeEmpty())
			Expect(string(jsonData)).To(ContainSubstring("original error"))
			Expect(string(jsonData)).To(ContainSubstring("wrapped message"))
			Expect(string(jsonData)).To(ContainSubstring("test"))
		})
	})
})

var _ = Describe("ErrorFactory", func() {
	var factory *errors.ErrorFactory

	BeforeEach(func() {
		factory = errors.NewErrorFactory()
	})

	Describe("Wrap", func() {
		It("should wrap errors with messages", func() {
			originalErr := fmt.Errorf("original")
			wrapped := factory.Wrap(originalErr, "wrapped: %s", "test")

			Expect(wrapped).ToNot(BeNil())
			Expect(wrapped.Error()).To(Equal("wrapped: test: original"))
		})

		It("should return nil for nil error", func() {
			wrapped := factory.Wrap(nil, "message")
			Expect(wrapped).To(BeNil())
		})
	})

	Describe("WrapWithContext", func() {
		It("should wrap with context", func() {
			ctx := context.WithValue(context.Background(), errors.OperationKey, "test-op")
			originalErr := fmt.Errorf("original")

			wrapped := factory.WrapWithContext(ctx, originalErr, "message")

			Expect(wrapped.Context().Operation).To(Equal("test-op"))
		})
	})

	Describe("Database", func() {
		It("should create database error with operation", func() {
			originalErr := fmt.Errorf("db connection failed")
			dbErr := factory.Database(originalErr, "user-query")

			Expect(dbErr.Error()).To(ContainSubstring("database operation failed"))
			Expect(dbErr.Context().Operation).To(Equal("user-query"))
		})
	})

	Describe("Validation", func() {
		It("should create validation error with field info", func() {
			originalErr := fmt.Errorf("invalid format")
			validationErr := factory.Validation(originalErr, "email")

			Expect(validationErr.Error()).To(ContainSubstring("validation failed"))
			Expect(validationErr.Context().Extra).To(HaveKeyWithValue("field", "email"))
		})
	})

	Describe("External", func() {
		It("should create external service error", func() {
			originalErr := fmt.Errorf("service unavailable")
			externalErr := factory.External(originalErr, "payment-gateway")

			Expect(externalErr.Error()).To(ContainSubstring("external service error"))
			Expect(externalErr.Context().Extra).To(HaveKeyWithValue("service", "payment-gateway"))
		})
	})

	Describe("Authentication", func() {
		It("should create authentication error", func() {
			originalErr := fmt.Errorf("invalid token")
			authErr := factory.Authentication(originalErr, "token expired")

			Expect(authErr.Error()).To(ContainSubstring("authentication failed"))
			Expect(authErr.Context().Extra).To(HaveKeyWithValue("reason", "token expired"))
		})
	})

	Describe("Authorization", func() {
		It("should create authorization error", func() {
			originalErr := fmt.Errorf("access denied")
			authzErr := factory.Authorization(originalErr, "users", "delete")

			Expect(authzErr.Error()).To(ContainSubstring("authorization failed"))
			Expect(authzErr.Context().Extra).To(HaveKeyWithValue("resource", "users"))
			Expect(authzErr.Context().Extra).To(HaveKeyWithValue("action", "delete"))
		})
	})
})

var _ = Describe("ErrorChain", func() {
	var chain *errors.ErrorChain

	BeforeEach(func() {
		chain = errors.NewErrorChain()
	})

	Describe("Add", func() {
		It("should add non-nil errors", func() {
			err1 := fmt.Errorf("error 1")
			err2 := fmt.Errorf("error 2")

			chain.Add(err1).Add(nil).Add(err2)

			Expect(chain.Count()).To(Equal(2))
			Expect(chain.Errors()).To(ConsistOf(err1, err2))
		})
	})

	Describe("HasErrors", func() {
		It("should return true when errors exist", func() {
			Expect(chain.HasErrors()).To(BeFalse())

			chain.Add(fmt.Errorf("error"))

			Expect(chain.HasErrors()).To(BeTrue())
		})
	})

	Describe("FirstError", func() {
		It("should return first error", func() {
			err1 := fmt.Errorf("first")
			err2 := fmt.Errorf("second")

			chain.Add(err1).Add(err2)

			Expect(chain.FirstError()).To(Equal(err1))
		})

		It("should return nil when no errors", func() {
			Expect(chain.FirstError()).To(BeNil())
		})
	})

	Describe("ToError", func() {
		It("should return nil when no errors", func() {
			Expect(chain.ToError()).To(BeNil())
		})

		It("should return single error when one exists", func() {
			err := fmt.Errorf("single error")
			chain.Add(err)

			Expect(chain.ToError()).To(Equal(err))
		})

		It("should combine multiple errors", func() {
			chain.Add(fmt.Errorf("error 1")).Add(fmt.Errorf("error 2"))

			combinedErr := chain.ToError()
			Expect(combinedErr.Error()).To(ContainSubstring("multiple errors occurred"))
			Expect(combinedErr.Error()).To(ContainSubstring("error 1"))
			Expect(combinedErr.Error()).To(ContainSubstring("error 2"))
		})
	})
})

var _ = Describe("ErrorGroup", func() {
	var group *errors.ErrorGroup

	BeforeEach(func() {
		group = errors.NewErrorGroup()
	})

	Describe("Add", func() {
		It("should add errors to categories", func() {
			validation := fmt.Errorf("validation error")
			database := fmt.Errorf("database error")

			group.Add("validation", validation).Add("database", database)

			Expect(group.HasCategory("validation")).To(BeTrue())
			Expect(group.HasCategory("database")).To(BeTrue())
			Expect(group.ErrorsInCategory("validation")).To(ContainElement(validation))
			Expect(group.ErrorsInCategory("database")).To(ContainElement(database))
		})
	})

	Describe("Categories", func() {
		It("should return all categories with errors", func() {
			group.Add("cat1", fmt.Errorf("error1")).Add("cat2", fmt.Errorf("error2"))

			categories := group.Categories()
			Expect(categories).To(ConsistOf("cat1", "cat2"))
		})
	})

	Describe("AllErrors", func() {
		It("should return all errors from all categories", func() {
			err1 := fmt.Errorf("error1")
			err2 := fmt.Errorf("error2")
			err3 := fmt.Errorf("error3")

			group.Add("cat1", err1).Add("cat1", err2).Add("cat2", err3)

			allErrors := group.AllErrors()
			Expect(allErrors).To(ConsistOf(err1, err2, err3))
		})
	})

	Describe("ToError", func() {
		It("should return nil when no errors", func() {
			Expect(group.ToError()).To(BeNil())
		})

		It("should combine errors from multiple categories", func() {
			group.Add("validation", fmt.Errorf("val error"))
			group.Add("database", fmt.Errorf("db error"))

			combinedErr := group.ToError()
			Expect(combinedErr.Error()).To(ContainSubstring("errors occurred in multiple categories"))
			Expect(combinedErr.Error()).To(ContainSubstring("validation"))
			Expect(combinedErr.Error()).To(ContainSubstring("database"))
		})
	})
})

var _ = Describe("ErrorRecovery", func() {
	var recovery *errors.ErrorRecovery

	BeforeEach(func() {
		recovery = errors.NewErrorRecovery()
	})

	Describe("Try", func() {
		It("should capture panics as errors", func() {
			err := recovery.Try(func() error {
				panic("test panic")
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("panic recovered"))
			Expect(err.Error()).To(ContainSubstring("test panic"))
		})

		It("should return normal errors unchanged", func() {
			originalErr := fmt.Errorf("normal error")
			err := recovery.Try(func() error {
				return originalErr
			})

			Expect(err).To(Equal(originalErr))
		})

		It("should return nil when no error", func() {
			err := recovery.Try(func() error {
				return nil
			})

			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("WithRetry", func() {
		It("should retry on retryable errors", func() {
			callCount := 0
			retryableErr := fmt.Errorf("retryable")

			err := recovery.WithRetry(func() error {
				callCount++
				if callCount < 3 {
					return retryableErr
				}
				return nil
			}, 3, func(err error) bool {
				return err == retryableErr
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(callCount).To(Equal(3))
		})

		It("should not retry on non-retryable errors", func() {
			callCount := 0
			nonRetryableErr := fmt.Errorf("non-retryable")

			err := recovery.WithRetry(func() error {
				callCount++
				return nonRetryableErr
			}, 3, func(err error) bool {
				return false
			})

			Expect(err).To(Equal(nonRetryableErr))
			Expect(callCount).To(Equal(1))
		})
	})
})

var _ = Describe("ErrorAggregator", func() {
	var aggregator *errors.ErrorAggregator

	BeforeEach(func() {
		aggregator = errors.NewErrorAggregator(3)
	})

	Describe("Add", func() {
		It("should collect errors and trigger when threshold reached", func() {
			Expect(aggregator.Add(fmt.Errorf("error 1"))).To(BeFalse())
			Expect(aggregator.Add(fmt.Errorf("error 2"))).To(BeFalse())
			Expect(aggregator.Add(fmt.Errorf("error 3"))).To(BeTrue()) // Threshold reached
		})

		It("should ignore nil errors", func() {
			aggregator.Add(nil)
			Expect(aggregator.Count()).To(Equal(0))
		})
	})

	Describe("ShouldTrigger", func() {
		It("should return true when threshold is met", func() {
			aggregator.Add(fmt.Errorf("1"))
			aggregator.Add(fmt.Errorf("2"))
			aggregator.Add(fmt.Errorf("3"))

			Expect(aggregator.ShouldTrigger()).To(BeTrue())
		})
	})

	Describe("Reset", func() {
		It("should clear all errors", func() {
			aggregator.Add(fmt.Errorf("error"))
			aggregator.Reset()

			Expect(aggregator.Count()).To(Equal(0))
			Expect(aggregator.ShouldTrigger()).To(BeFalse())
		})
	})
})

var _ = Describe("Utility Functions", func() {
	Describe("IgnoreError", func() {
		It("should ignore specified errors", func() {
			targetErr := fmt.Errorf("target error")
			otherErr := fmt.Errorf("other error")

			result := errors.IgnoreError(targetErr, targetErr, otherErr)
			Expect(result).To(BeNil())
		})

		It("should not ignore unspecified errors", func() {
			targetErr := fmt.Errorf("target error")
			ignoreErr := fmt.Errorf("ignore error")

			result := errors.IgnoreError(targetErr, ignoreErr)
			Expect(result).To(Equal(targetErr))
		})

		It("should return nil for nil error", func() {
			result := errors.IgnoreError(nil, fmt.Errorf("ignore"))
			Expect(result).To(BeNil())
		})
	})

	Describe("FormatErrorChain", func() {
		It("should format error chain", func() {
			innerErr := fmt.Errorf("inner error")
			middleErr := fmt.Errorf("middle error: %w", innerErr)
			outerErr := fmt.Errorf("outer error: %w", middleErr)

			formatted := errors.FormatErrorChain(outerErr)
			Expect(formatted).To(ContainSubstring("outer error"))
			Expect(formatted).To(ContainSubstring("middle error"))
			Expect(formatted).To(ContainSubstring("inner error"))
			Expect(formatted).To(ContainSubstring("->"))
		})

		It("should return empty string for nil error", func() {
			formatted := errors.FormatErrorChain(nil)
			Expect(formatted).To(BeEmpty())
		})
	})
})
