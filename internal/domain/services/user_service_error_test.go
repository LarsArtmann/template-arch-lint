package services_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	domainErrors "github.com/LarsArtmann/template-arch-lint/pkg/errors"
)

func TestUserServiceErrorPaths(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "🚨 UserService Error Path Testing Suite")
}

// Mock repository that can simulate various failure scenarios.
type FailingUserRepository struct {
	saveError         error
	findByIDError     error
	findByEmailError  error
	updateError       error
	deleteError       error
	listError         error
	countError        error
	existsError       error
	findByUsernameErr error

	// Test data for success scenarios
	testUser *entities.User

	saveCallCount           int
	findByIDCallCount       int
	findByEmailCallCount    int
	updateCallCount         int
	deleteCallCount         int
	listCallCount           int
	countCallCount          int
	existsCallCount         int
	findByUsernameCallCount int
}

func NewFailingUserRepository() *FailingUserRepository {
	return &FailingUserRepository{}
}

func (r *FailingUserRepository) Save(ctx context.Context, user *entities.User) error {
	r.saveCallCount++

	// Check context cancellation first
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if r.saveError != nil {
		return r.saveError
	}
	return nil
}

func (r *FailingUserRepository) FindByID(ctx context.Context, id values.UserID) (*entities.User, error) {
	r.findByIDCallCount++

	// Check context cancellation first
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if r.findByIDError != nil {
		return nil, r.findByIDError
	}
	if r.testUser != nil {
		return r.testUser, nil
	}
	return nil, repositories.ErrUserNotFound
}

func (r *FailingUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	r.findByEmailCallCount++
	if r.findByEmailError != nil {
		return nil, r.findByEmailError
	}
	return nil, repositories.ErrUserNotFound
}

func (r *FailingUserRepository) Update(ctx context.Context, user *entities.User) error {
	r.updateCallCount++
	if r.updateError != nil {
		return r.updateError
	}
	return nil
}

func (r *FailingUserRepository) Delete(ctx context.Context, id values.UserID) error {
	r.deleteCallCount++
	if r.deleteError != nil {
		return r.deleteError
	}
	return nil
}

func (r *FailingUserRepository) List(ctx context.Context) ([]*entities.User, error) {
	r.listCallCount++
	if r.listError != nil {
		return nil, r.listError
	}
	if r.testUser != nil {
		return []*entities.User{r.testUser}, nil
	}
	return []*entities.User{}, nil
}

func (r *FailingUserRepository) Count(ctx context.Context) (int, error) {
	r.countCallCount++
	if r.countError != nil {
		return 0, r.countError
	}
	return 0, nil
}

func (r *FailingUserRepository) Exists(ctx context.Context, id values.UserID) (bool, error) {
	r.existsCallCount++
	if r.existsError != nil {
		return false, r.existsError
	}
	return false, nil
}

func (r *FailingUserRepository) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	r.findByUsernameCallCount++
	if r.findByUsernameErr != nil {
		return nil, r.findByUsernameErr
	}
	return nil, repositories.ErrUserNotFound
}

var _ = Describe("🚨 UserService Error Path Testing", func() {
	var (
		userService *services.UserService
		failingRepo *FailingUserRepository
		ctx         context.Context
	)

	// Test helper functions
	createTestUserID := func(id string) values.UserID {
		userID, err := values.NewUserID(id)
		Expect(err).ToNot(HaveOccurred())
		return userID
	}

	// Helper to check if error is wrapped InternalError with specific cause
	expectInternalErrorWithCause := func(err, expectedCause error, expectedMessagePrefix string) {
		Expect(domainErrors.IsDomainError(err)).To(BeTrue())
		internalErr, ok := domainErrors.AsInternalError(err)
		Expect(ok).To(BeTrue())
		Expect(internalErr.Cause()).To(Equal(expectedCause))
		Expect(internalErr.Error()).To(ContainSubstring(expectedMessagePrefix))
	}

	BeforeEach(func() {
		ctx = context.Background()
		failingRepo = NewFailingUserRepository()
		userService = services.NewUserService(failingRepo)
	})

	Describe("🔥 Repository Error Propagation", func() {
		Context("CreateUser with repository failures", func() {
			It("should handle FindByEmail repository errors", func() {
				failingRepo.findByEmailError = sql.ErrConnDone

				id := createTestUserID("test-user")
				user, err := userService.CreateUser(ctx, id, "test@example.com", "Test User")

				Expect(user).To(BeNil())
				// Service should wrap repository error in InternalError
				Expect(domainErrors.IsDomainError(err)).To(BeTrue())
				internalErr, ok := domainErrors.AsInternalError(err)
				Expect(ok).To(BeTrue())
				Expect(internalErr.Cause()).To(Equal(sql.ErrConnDone))
				Expect(internalErr.Error()).To(Equal("failed to check existing user: sql: connection is already closed"))
				Expect(failingRepo.findByEmailCallCount).To(Equal(1))
				Expect(failingRepo.saveCallCount).To(Equal(0)) // Should not reach Save
			})

			It("should handle Save repository errors", func() {
				failingRepo.saveError = sql.ErrTxDone

				id := createTestUserID("test-user")
				user, err := userService.CreateUser(ctx, id, "test@example.com", "Test User")

				Expect(user).To(BeNil())
				// Service should wrap repository error in InternalError
				Expect(domainErrors.IsDomainError(err)).To(BeTrue())
				internalErr, ok := domainErrors.AsInternalError(err)
				Expect(ok).To(BeTrue())
				Expect(internalErr.Cause()).To(Equal(sql.ErrTxDone))
				Expect(internalErr.Error()).To(Equal("failed to save user: sql: transaction has already been committed or rolled back"))
				Expect(failingRepo.findByEmailCallCount).To(Equal(1))
				Expect(failingRepo.saveCallCount).To(Equal(1))
			})

			It("should handle concurrent access errors", func() {
				failingRepo.saveError = errors.New("UNIQUE constraint failed: users.email")

				id := createTestUserID("test-user")
				user, err := userService.CreateUser(ctx, id, "test@example.com", "Test User")

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("UNIQUE constraint failed"))
			})
		})

		Context("GetUser with repository failures", func() {
			It("should handle FindByID repository errors", func() {
				failingRepo.findByIDError = sql.ErrNoRows

				id := createTestUserID("test-user")
				user, err := userService.GetUser(ctx, id)

				Expect(user).To(BeNil())
				expectInternalErrorWithCause(err, sql.ErrNoRows, "failed to get user")
				Expect(failingRepo.findByIDCallCount).To(Equal(1))
			})

			It("should handle network timeout errors", func() {
				failingRepo.findByIDError = context.DeadlineExceeded

				id := createTestUserID("test-user")
				user, err := userService.GetUser(ctx, id)

				Expect(user).To(BeNil())
				expectInternalErrorWithCause(err, context.DeadlineExceeded, "failed to get user")
				Expect(failingRepo.findByIDCallCount).To(Equal(1))
			})
		})

		Context("UpdateUser with repository failures", func() {
			It("should handle FindByID errors during update", func() {
				failingRepo.findByIDError = sql.ErrConnDone

				id := createTestUserID("test-user")
				user, err := userService.UpdateUser(ctx, id, "new@example.com", "New Name")

				Expect(user).To(BeNil())
				expectInternalErrorWithCause(err, sql.ErrConnDone, "failed to get for update user")
				Expect(failingRepo.findByIDCallCount).To(Equal(1))
				Expect(failingRepo.updateCallCount).To(Equal(0)) // Should not reach Update
			})

			It("should handle Update repository errors", func() {
				// Make FindByID succeed but Save fail during update
				testUser, _ := entities.NewUser(createTestUserID("test-user"), "test@example.com", "Test User")
				failingRepo.testUser = testUser
				failingRepo.findByIDError = nil
				failingRepo.saveError = sql.ErrTxDone // Save method is called during update

				id := createTestUserID("test-user")
				user, err := userService.UpdateUser(ctx, id, "new@example.com", "New Name")

				// Update should fail during repository Save operation
				Expect(err).To(HaveOccurred())
				Expect(user).To(BeNil())
				Expect(failingRepo.saveCallCount).To(Equal(1))
			})
		})

		Context("DeleteUser with repository failures", func() {
			It("should handle Delete repository errors", func() {
				// Make FindByID succeed but Delete fail
				testUser, _ := entities.NewUser(createTestUserID("test-user"), "test@example.com", "Test User")
				failingRepo.testUser = testUser
				failingRepo.findByIDError = nil
				failingRepo.deleteError = sql.ErrConnDone

				id := createTestUserID("test-user")
				err := userService.DeleteUser(ctx, id)

				expectInternalErrorWithCause(err, sql.ErrConnDone, "failed to delete user")
				Expect(failingRepo.deleteCallCount).To(Equal(1))
			})

			It("should handle foreign key constraint errors", func() {
				// Make FindByID succeed but Delete fail
				testUser, _ := entities.NewUser(createTestUserID("test-user"), "test@example.com", "Test User")
				failingRepo.testUser = testUser
				failingRepo.findByIDError = nil
				failingRepo.deleteError = errors.New("FOREIGN KEY constraint failed")

				id := createTestUserID("test-user")
				err := userService.DeleteUser(ctx, id)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to delete user: FOREIGN KEY constraint failed"))
			})
		})

		Context("ListUsers with repository failures", func() {
			It("should handle List repository errors", func() {
				// Make List fail
				testUser, _ := entities.NewUser(createTestUserID("test-user"), "test@example.com", "Test User")
				failingRepo.testUser = testUser
				failingRepo.findByIDError = nil
				failingRepo.listError = sql.ErrConnDone

				users, err := userService.ListUsers(ctx)

				Expect(users).To(BeNil())
				expectInternalErrorWithCause(err, sql.ErrConnDone, "failed to list users")
				Expect(failingRepo.listCallCount).To(Equal(1))
			})
		})
	})

	Describe("⏰ Context Cancellation and Timeouts", func() {
		Context("with cancelled context", func() {
			It("should handle context cancellation in CreateUser", func() {
				cancelledCtx, cancel := context.WithCancel(ctx)
				cancel()

				id := createTestUserID("test-user")
				user, err := userService.CreateUser(cancelledCtx, id, "test@example.com", "Test User")

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
				// The repository should handle context cancellation
				// Note: Context cancellation is handled by the service/repository, not wrapped
			})

			It("should handle context cancellation in GetUser", func() {
				cancelledCtx, cancel := context.WithCancel(ctx)
				cancel()

				id := createTestUserID("test-user")
				user, err := userService.GetUser(cancelledCtx, id)

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
				// Note: Context cancellation is handled by service/repository, not wrapped
			})
		})

		Context("with timeout context", func() {
			It("should handle timeout in operations", func() {
				timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()

				// Add delay to ensure timeout
				time.Sleep(10 * time.Nanosecond)

				id := createTestUserID("test-user")
				user, err := userService.CreateUser(timeoutCtx, id, "test@example.com", "Test User")

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
				// Note: Timeout is handled by context, not wrapped in service
			})
		})
	})

	Describe("🚫 Validation Error Scenarios", func() {
		Context("with invalid user data", func() {
			It("should return proper validation errors for invalid email formats", func() {
				invalidEmails := []string{
					"",
					"invalid",
					"@example.com",
					"user@",
					"user@@example.com",
					"user..name@example.com",
					".user@example.com",
					"user.@example.com",
					"user name@example.com",
				}

				for _, email := range invalidEmails {
					id := createTestUserID("test-user")
					user, err := userService.CreateUser(ctx, id, email, "Test User")

					Expect(user).To(BeNil(), "should reject email: %s", email)
					Expect(err).To(HaveOccurred(), "should return error for email: %s", email)

					_, isValidationError := domainErrors.AsValidationError(err)
					Expect(isValidationError).To(BeTrue(), "should be validation error for email: %s", email)
				}
			})

			It("should return proper validation errors for invalid names", func() {
				invalidNames := []string{
					"",
					"A",
					"123",
					"   ",
					"@#$%",
					" John Doe ",
				}

				for _, name := range invalidNames {
					id := createTestUserID("test-user")
					user, err := userService.CreateUser(ctx, id, "test@example.com", name)

					Expect(user).To(BeNil(), "should reject name: %s", name)
					Expect(err).To(HaveOccurred(), "should return error for name: %s", name)

					_, isValidationError := domainErrors.AsValidationError(err)
					Expect(isValidationError).To(BeTrue(), "should be validation error for name: %s", name)
				}
			})
		})

		Context("with empty UserID", func() {
			It("should handle empty UserID gracefully", func() {
				emptyID := values.UserID{}
				user, err := userService.CreateUser(ctx, emptyID, "test@example.com", "Test User")

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("🔀 Race Condition and Concurrent Error Scenarios", func() {
		Context("with simulated race conditions", func() {
			It("should handle repository state changes between operations", func() {
				// Simulate a scenario where user exists during check but is deleted before save
				failingRepo.findByEmailError = repositories.ErrUserNotFound // User doesn't exist
				failingRepo.saveError = errors.New("user was deleted by another process")

				id := createTestUserID("test-user")
				user, err := userService.CreateUser(ctx, id, "test@example.com", "Test User")

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("deleted by another process"))
			})

			It("should handle email conflict during concurrent creation", func() {
				// Simulate scenario where email check passes but save fails due to constraint
				failingRepo.findByEmailError = repositories.ErrUserNotFound
				failingRepo.saveError = repositories.ErrUserAlreadyExists

				id := createTestUserID("test-user")
				user, err := userService.CreateUser(ctx, id, "test@example.com", "Test User")

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to save user: user already exists"))
			})
		})
	})

	Describe("🔍 Functional Programming Error Propagation", func() {
		Context("FilterActiveUsers with repository errors", func() {
			It("should propagate List errors", func() {
				failingRepo.listError = sql.ErrConnDone

				activeUsers, err := userService.FilterActiveUsers(ctx)

				Expect(activeUsers).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to list users: sql: connection is already closed"))
			})
		})

		Context("GetUserEmailsWithResult with repository errors", func() {
			It("should return error result for List failures", func() {
				failingRepo.listError = sql.ErrConnDone

				result := userService.GetUserEmailsWithResult(ctx)

				Expect(result.IsError()).To(BeTrue())
				Expect(result.Error()).To(Equal(sql.ErrConnDone))
			})
		})

		Context("GetUserStats with repository errors", func() {
			It("should handle List errors gracefully", func() {
				failingRepo.listError = sql.ErrConnDone

				stats, err := userService.GetUserStats(ctx)

				Expect(stats).To(BeNil())
				Expect(err).To(Equal(sql.ErrConnDone))
			})
		})

		Context("CreateUserWithResult error handling", func() {
			It("should return error result for repository failures", func() {
				failingRepo.saveError = sql.ErrConnDone

				id := createTestUserID("test-user")
				result := userService.CreateUserWithResult(ctx, id, "test@example.com", "Test User")

				Expect(result.IsError()).To(BeTrue())
				Expect(result.Error()).To(Equal(sql.ErrConnDone))
			})
		})

		Context("FindUserByEmailOption error handling", func() {
			It("should return None for repository errors", func() {
				failingRepo.findByEmailError = sql.ErrConnDone

				option := userService.FindUserByEmailOption(ctx, "test@example.com")

				Expect(option.IsAbsent()).To(BeTrue())
			})
		})
	})

	Describe("📊 Error Recovery and Resilience", func() {
		Context("partial failure scenarios", func() {
			It("should maintain consistency during partial failures", func() {
				// Test scenario where some operations succeed and others fail
				id := createTestUserID("partial-failure-user")

				// First operation succeeds
				failingRepo.findByEmailError = repositories.ErrUserNotFound
				failingRepo.saveError = nil

				user, err := userService.CreateUser(ctx, id, "test@example.com", "Test User")
				Expect(err).ToNot(HaveOccurred())
				Expect(user).ToNot(BeNil())

				// Second operation (update) fails
				failingRepo.findByIDError = sql.ErrConnDone

				updatedUser, err := userService.UpdateUser(ctx, id, "new@example.com", "New Name")
				Expect(updatedUser).To(BeNil())
				Expect(err).To(Equal(sql.ErrConnDone))

				// Service should handle this gracefully without corruption
				Expect(failingRepo.findByIDCallCount).To(Equal(1))
				Expect(failingRepo.updateCallCount).To(Equal(0))
			})
		})

		Context("error type preservation", func() {
			It("should preserve specific error types through service layer", func() {
				customError := errors.New("custom database error")
				failingRepo.saveError = customError

				id := createTestUserID("test-user")
				user, err := userService.CreateUser(ctx, id, "test@example.com", "Test User")

				Expect(user).To(BeNil())
				Expect(err).To(Equal(customError))
				Expect(err.Error()).To(Equal("custom database error"))
			})
		})
	})
})
