package services_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/LarsArtmann/template-arch-lint/pkg/errors"
)

func TestUserServiceCRUD(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UserService CRUD Suite")
}

var _ = Describe("UserService CRUD Operations", func() {
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

	defaultTestEmail := "test@example.com"
	defaultTestName := "Test User"

	Describe("CreateUser", func() {
		Context("with valid input", func() {
			It("should create a new user successfully", func() {
				id := createTestUserID("test-user-1")
				user, err := userService.CreateUser(ctx, id, defaultTestEmail, defaultTestName)
				Expect(err).ToNot(HaveOccurred())
				Expect(user).ToNot(BeNil())
				Expect(user.ID).To(Equal(id))
			})
		})

		Context("with invalid email", func() {
			It("should return validation error for empty email", func() {
				id := createTestUserID("test-user-1")
				user, err := userService.CreateUser(ctx, id, "", defaultTestName)
				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
				_, isValidationError := errors.AsValidationError(err)
				Expect(isValidationError).To(BeTrue())
			})
		})
	})

	Describe("GetUser", func() {
		Context("when user exists", func() {
			It("should return the user", func() {
				id := createTestUserID("test-user-1")
				createdUser, err := userService.CreateUser(ctx, id, defaultTestEmail, defaultTestName)
				Expect(err).ToNot(HaveOccurred())

				retrievedUser, err := userService.GetUser(ctx, id)
				Expect(err).ToNot(HaveOccurred())
				Expect(retrievedUser).ToNot(BeNil())
				Expect(retrievedUser.ID).To(Equal(createdUser.ID))
			})
		})

		Context("when user does not exist", func() {
			It("should return not found error", func() {
				id := createTestUserID("non-existent-user")
				user, err := userService.GetUser(ctx, id)
				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
				_, isNotFoundError := errors.AsNotFoundError(err)
				Expect(isNotFoundError).To(BeTrue())
			})
		})
	})

	Describe("UpdateUser", func() {
		Context("with valid data", func() {
			It("should update user successfully", func() {
				id := createTestUserID("test-user-1")
				createdUser, err := userService.CreateUser(ctx, id, defaultTestEmail, defaultTestName)
				Expect(err).ToNot(HaveOccurred())

				newEmail := "updated@example.com"
				newName := "Updated User"
				updatedUser, err := userService.UpdateUser(ctx, id, newEmail, newName)
				Expect(err).ToNot(HaveOccurred())
				Expect(updatedUser).ToNot(BeNil())
				Expect(updatedUser.ID).To(Equal(createdUser.ID))
			})
		})
	})

	Describe("DeleteUser", func() {
		Context("when user exists", func() {
			It("should delete user successfully", func() {
				id := createTestUserID("test-user-1")
				_, err := userService.CreateUser(ctx, id, defaultTestEmail, defaultTestName)
				Expect(err).ToNot(HaveOccurred())

				err = userService.DeleteUser(ctx, id)
				Expect(err).ToNot(HaveOccurred())

				// Verify user is deleted
				retrievedUser, err := userService.GetUser(ctx, id)
				Expect(retrievedUser).To(BeNil())
				Expect(err).To(HaveOccurred())
				_, isNotFoundError := errors.AsNotFoundError(err)
				Expect(isNotFoundError).To(BeTrue())
			})
		})

		Context("when user does not exist", func() {
			It("should return not found error", func() {
				id := createTestUserID("non-existent-user")
				err := userService.DeleteUser(ctx, id)
				Expect(err).To(HaveOccurred())
				_, isNotFoundError := errors.AsNotFoundError(err)
				Expect(isNotFoundError).To(BeTrue())
			})
		})
	})
})
