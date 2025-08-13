package services_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

func TestUserService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UserService Suite")
}

var _ = Describe("UserService", func() {
	var (
		userService *services.UserService
		userRepo    repositories.UserRepository
		ctx         context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		userRepo = repositories.NewInMemoryUserRepository()
		userService = services.NewUserService(userRepo)
	})

	// Test helper functions
	createTestUserID := func(id string) values.UserID {
		userID, err := values.NewUserID(id)
		Expect(err).ToNot(HaveOccurred())
		return userID
	}

	assertValidationError := func(user *entities.User, err error) {
		Expect(user).To(BeNil())
		Expect(err).To(HaveOccurred())
		_, isValidationError := errors.AsValidationError(err)
		Expect(isValidationError).To(BeTrue())
	}

	createUserAndExpectValidationError := func(email, name string) {
		id := createTestUserID("test-user-1")
		user, err := userService.CreateUser(ctx, id, email, name)
		assertValidationError(user, err)
	}

	createValidTestUser := func(userIDSuffix, email, name string) *entities.User {
		id := createTestUserID(userIDSuffix)
		user, err := userService.CreateUser(ctx, id, email, name)
		Expect(err).ToNot(HaveOccurred())
		return user
	}

	Describe("CreateUser", func() {
		Context("with valid input", func() {
			It("should create a new user successfully", func() {
				id, err := values.NewUserID("test-user-1")
				Expect(err).ToNot(HaveOccurred())
				email := "test@example.com"
				name := "Test User"

				user, err := userService.CreateUser(ctx, id, email, name)

				Expect(err).ToNot(HaveOccurred())
				Expect(user).ToNot(BeNil())
				Expect(user.ID).To(Equal(id))
				Expect(user.Email).To(Equal(email))
				Expect(user.Name).To(Equal(name))
			})
		})

		Context("with invalid email", func() {
			It("should return validation error for empty email", func() {
				createUserAndExpectValidationError("", "Test User")
			})

			It("should return validation error for invalid email format", func() {
				createUserAndExpectValidationError("invalid-email", "Test User")
			})

			It("should return validation error for email with spaces", func() {
				createUserAndExpectValidationError("test @example.com", "Test User")
			})
		})

		Context("with invalid name", func() {
			It("should return validation error for empty name", func() {
				createUserAndExpectValidationError("test@example.com", "")
			})

			It("should return validation error for too short name", func() {
				createUserAndExpectValidationError("test@example.com", "A")
			})

			It("should return validation error for name without letters", func() {
				createUserAndExpectValidationError("test@example.com", "123")
			})
		})

		Context("when user already exists", func() {
			It("should return conflict error", func() {
				id1, err := values.NewUserID("test-user-1")
				Expect(err).ToNot(HaveOccurred())
				id2, err := values.NewUserID("test-user-2")
				Expect(err).ToNot(HaveOccurred())
				email := "test@example.com"
				name := "Test User"

				// Create first user
				_, err = userService.CreateUser(ctx, id1, email, name)
				Expect(err).ToNot(HaveOccurred())

				// Try to create second user with same email
				user, err := userService.CreateUser(ctx, id2, email, name)

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(repositories.ErrUserAlreadyExists))
			})
		})
	})

	Describe("GetUser", func() {
		Context("when user exists", func() {
			It("should return the user", func() {
				id, err := values.NewUserID("test-user-1")
				Expect(err).ToNot(HaveOccurred())
				email := "test@example.com"
				name := "Test User"

				// Create user first
				createdUser, err := userService.CreateUser(ctx, id, email, name)
				Expect(err).ToNot(HaveOccurred())

				// Get user
				user, err := userService.GetUser(ctx, id)

				Expect(err).ToNot(HaveOccurred())
				Expect(user).ToNot(BeNil())
				Expect(user.ID).To(Equal(createdUser.ID))
				Expect(user.Email).To(Equal(createdUser.Email))
				Expect(user.Name).To(Equal(createdUser.Name))
			})
		})

		Context("when user does not exist", func() {
			It("should return not found error", func() {
				id, err := values.NewUserID("nonexistent-user")
				Expect(err).ToNot(HaveOccurred())

				user, err := userService.GetUser(ctx, id)

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("UpdateUser", func() {
		var existingUser *entities.User

		BeforeEach(func() {
			existingUser = createValidTestUser("test-user-1", "test@example.com", "Test User")
		})

		Context("with valid input", func() {
			It("should update user successfully", func() {
				newEmail := "updated@example.com"
				newName := "Updated User"

				user, err := userService.UpdateUser(ctx, existingUser.ID, newEmail, newName)

				Expect(err).ToNot(HaveOccurred())
				Expect(user).ToNot(BeNil())
				Expect(user.Email).To(Equal(newEmail))
				Expect(user.Name).To(Equal(newName))
			})
		})

		Context("when user does not exist", func() {
			It("should return error", func() {
				id, err := values.NewUserID("nonexistent-user")
				Expect(err).ToNot(HaveOccurred())
				email := "updated@example.com"
				name := "Updated User"

				user, err := userService.UpdateUser(ctx, id, email, name)

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with invalid email", func() {
			It("should return validation error", func() {
				invalidEmail := "invalid-email"
				name := "Updated User"

				user, err := userService.UpdateUser(ctx, existingUser.ID, invalidEmail, name)

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("DeleteUser", func() {
		var existingUser *entities.User

		BeforeEach(func() {
			existingUser = createValidTestUser("test-user-1", "test@example.com", "Test User")
		})

		Context("when user exists", func() {
			It("should delete user successfully", func() {
				err := userService.DeleteUser(ctx, existingUser.ID)

				Expect(err).ToNot(HaveOccurred())

				// Verify user is deleted
				user, err := userService.GetUser(ctx, existingUser.ID)
				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when user does not exist", func() {
			It("should return error", func() {
				id, err := values.NewUserID("nonexistent-user")
				Expect(err).ToNot(HaveOccurred())

				err = userService.DeleteUser(ctx, id)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("ListUsers", func() {
		Context("when users exist", func() {
			It("should return all users", func() {
				// Create multiple users
				id1, err := values.NewUserID("user-1")
				Expect(err).ToNot(HaveOccurred())
				user1, err := userService.CreateUser(ctx, id1, "user1@example.com", "User One")
				Expect(err).ToNot(HaveOccurred())

				id2, err := values.NewUserID("user-2")
				Expect(err).ToNot(HaveOccurred())
				user2, err := userService.CreateUser(ctx, id2, "user2@example.com", "User Two")
				Expect(err).ToNot(HaveOccurred())

				users, err := userService.ListUsers(ctx)

				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(HaveLen(2))
				Expect(users).To(ContainElement(user1))
				Expect(users).To(ContainElement(user2))
			})
		})

		Context("when no users exist", func() {
			It("should return empty list", func() {
				users, err := userService.ListUsers(ctx)

				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(HaveLen(0))
			})
		})
	})

	Describe("Functional Programming Methods", func() {
		Describe("FilterActiveUsers", func() {
			It("should filter users created in the last 30 days", func() {
				// Create a user
				id, err := values.NewUserID("user-1")
				Expect(err).ToNot(HaveOccurred())
				_, err = userService.CreateUser(ctx, id, "user1@example.com", "User One")
				Expect(err).ToNot(HaveOccurred())

				activeUsers, err := userService.FilterActiveUsers(ctx)

				Expect(err).ToNot(HaveOccurred())
				Expect(activeUsers).To(HaveLen(1))
			})
		})

		Describe("GetUserEmailsWithResult", func() {
			It("should return user emails using Result pattern", func() {
				// Create users
				id1, err := values.NewUserID("user-1")
				Expect(err).ToNot(HaveOccurred())
				_, err = userService.CreateUser(ctx, id1, "user1@example.com", "User One")
				Expect(err).ToNot(HaveOccurred())

				id2, err := values.NewUserID("user-2")
				Expect(err).ToNot(HaveOccurred())
				_, err = userService.CreateUser(ctx, id2, "user2@example.com", "User Two")
				Expect(err).ToNot(HaveOccurred())

				result := userService.GetUserEmailsWithResult(ctx)

				Expect(result.IsOk()).To(BeTrue())
				emails, _ := result.Get()
				Expect(emails).To(HaveLen(2))
				Expect(emails).To(ContainElement("user1@example.com"))
				Expect(emails).To(ContainElement("user2@example.com"))
			})
		})

		Describe("GetUserStats", func() {
			It("should return user statistics", func() {
				// Create users
				id1, err := values.NewUserID("user-1")
				Expect(err).ToNot(HaveOccurred())
				_, err = userService.CreateUser(ctx, id1, "user1@example.com", "User One")
				Expect(err).ToNot(HaveOccurred())

				id2, err := values.NewUserID("user-2")
				Expect(err).ToNot(HaveOccurred())
				_, err = userService.CreateUser(ctx, id2, "user2@example.com", "User Two")
				Expect(err).ToNot(HaveOccurred())

				stats, err := userService.GetUserStats(ctx)

				Expect(err).ToNot(HaveOccurred())
				Expect(stats["total"]).To(Equal(2))
				Expect(stats["active"]).To(Equal(2))
				Expect(stats["domains"]).To(Equal(1))
				Expect(stats).To(HaveKey("avg_days_since_registration"))
			})
		})
	})

	Describe("Result Pattern Methods", func() {
		Describe("CreateUserWithResult", func() {
			Context("with valid input", func() {
				It("should create user successfully", func() {
					id, err := values.NewUserID("test-user-1")
					Expect(err).ToNot(HaveOccurred())
					email := "test@example.com"
					name := "Test User"

					result := userService.CreateUserWithResult(ctx, id, email, name)

					Expect(result.IsOk()).To(BeTrue())
					user, _ := result.Get()
					Expect(user.ID).To(Equal(id))
					Expect(user.Email).To(Equal(email))
					Expect(user.Name).To(Equal(name))
				})
			})

			Context("with invalid input", func() {
				It("should return error result", func() {
					id, err := values.NewUserID("test-user-1")
					Expect(err).ToNot(HaveOccurred())
					email := "invalid-email"
					name := "Test User"

					result := userService.CreateUserWithResult(ctx, id, email, name)

					Expect(result.IsError()).To(BeTrue())
					_, isValidationError := errors.AsValidationError(result.Error())
					Expect(isValidationError).To(BeTrue())
				})
			})
		})

		Describe("FindUserByEmailOption", func() {
			Context("when user exists", func() {
				It("should return Some with user", func() {
					id, err := values.NewUserID("test-user-1")
					Expect(err).ToNot(HaveOccurred())
					email := "test@example.com"
					name := "Test User"

					// Create user first
					_, err = userService.CreateUser(ctx, id, email, name)
					Expect(err).ToNot(HaveOccurred())

					option := userService.FindUserByEmailOption(ctx, email)

					Expect(option.IsPresent()).To(BeTrue())
					user, _ := option.Get()
					Expect(user.Email).To(Equal(email))
				})
			})

			Context("when user does not exist", func() {
				It("should return None", func() {
					email := "nonexistent@example.com"

					option := userService.FindUserByEmailOption(ctx, email)

					Expect(option.IsAbsent()).To(BeTrue())
				})
			})

			Context("with invalid email", func() {
				It("should return None", func() {
					email := "invalid-email"

					option := userService.FindUserByEmailOption(ctx, email)

					Expect(option.IsAbsent()).To(BeTrue())
				})
			})
		})
	})
})