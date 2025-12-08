package services_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/LarsArtmann/template-arch-lint/pkg/errors"
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

	// Additional test helpers to eliminate duplication
	createTestUser := func(idSuffix, email, name string) (*entities.User, error) {
		id := createTestUserID(idSuffix)
		return userService.CreateUser(ctx, id, email, name)
	}

	defaultTestEmail := "test@example.com"
	defaultTestName := "Test User"

	createDefaultTestUser := func(idSuffix string) *entities.User {
		return createValidTestUser(idSuffix, defaultTestEmail, defaultTestName)
	}

	expectSuccessfulUserCreation := func(user *entities.User, err error, expectedID values.UserID, expectedEmail, expectedName string) {
		Expect(err).ToNot(HaveOccurred())
		Expect(user).ToNot(BeNil())
		Expect(user.ID).To(Equal(expectedID))
		Expect(user.GetEmail().String()).To(Equal(expectedEmail))
		Expect(user.GetUserName().String()).To(Equal(expectedName))
	}

	Describe("CreateUser", func() {
		Context("with valid input", func() {
			It("should create a new user successfully", func() {
				id := createTestUserID("test-user-1")
				user, err := createTestUser("test-user-1", defaultTestEmail, defaultTestName)
				expectSuccessfulUserCreation(user, err, id, defaultTestEmail, defaultTestName)
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
				// Create first user
				createDefaultTestUser("test-user-1")

				// Try to create second user with same email
				id2 := createTestUserID("test-user-2")
				user, err := userService.CreateUser(ctx, id2, defaultTestEmail, defaultTestName)

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(repositories.ErrUserAlreadyExists))
			})
		})
	})

	Describe("GetUser", func() {
		Context("when user exists", func() {
			It("should return the user", func() {
				// Create user first
				createdUser := createDefaultTestUser("test-user-1")

				// Get user
				user, err := userService.GetUser(ctx, createdUser.ID)

				Expect(err).ToNot(HaveOccurred())
				Expect(user).ToNot(BeNil())
				Expect(user.ID).To(Equal(createdUser.ID))
				Expect(user.GetEmail().String()).To(Equal(createdUser.GetEmail().String()))
				Expect(user.GetUserName().String()).To(Equal(createdUser.GetUserName().String()))
			})
		})

		Context("when user does not exist", func() {
			It("should return not found error", func() {
				id := createTestUserID("nonexistent-user")

				user, err := userService.GetUser(ctx, id)

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("UpdateUser", func() {
		var existingUser *entities.User

		BeforeEach(func() {
			existingUser = createDefaultTestUser("test-user-1")
		})

		Context("with valid input", func() {
			It("should update user successfully", func() {
				newEmail := "updated@example.com"
				newName := "Updated User"

				user, err := userService.UpdateUser(ctx, existingUser.ID, newEmail, newName)

				Expect(err).ToNot(HaveOccurred())
				Expect(user).ToNot(BeNil())
				Expect(user.GetEmail().String()).To(Equal(newEmail))
				Expect(user.GetUserName().String()).To(Equal(newName))
			})
		})

		Context("when user does not exist", func() {
			It("should return error", func() {
				id := createTestUserID("nonexistent-user")

				user, err := userService.UpdateUser(ctx, id, "updated@example.com", "Updated User")

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
				user1 := createValidTestUser("user-1", "user1@example.com", "User One")
				user2 := createValidTestUser("user-2", "user2@example.com", "User Two")

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
				Expect(users).To(BeEmpty())
			})
		})
	})

	Describe("Functional Programming Methods", func() {
		Describe("FilterActiveUsers", func() {
			It("should filter users created in the last 30 days", func() {
				// Create a user
				createValidTestUser("user-1", "user1@example.com", "User One")

				activeUsers, err := userService.FilterActiveUsers(ctx)

				Expect(err).ToNot(HaveOccurred())
				Expect(activeUsers).To(HaveLen(1))
			})
		})

		Describe("GetUserEmailsWithResult", func() {
			It("should return user emails using Result pattern", func() {
				// Create users
				createValidTestUser("user-1", "user1@example.com", "User One")
				createValidTestUser("user-2", "user2@example.com", "User Two")

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
				createValidTestUser("user-1", "user1@example.com", "User One")
				createValidTestUser("user-2", "user2@example.com", "User Two")

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
					Expect(user.GetEmail().String()).To(Equal(email))
					Expect(user.GetUserName().String()).To(Equal(name))
				})
			})

			Context("with invalid input", func() {
				It("should return error result", func() {
					id := createTestUserID("test-user-1")

					result := userService.CreateUserWithResult(ctx, id, "invalid-email", defaultTestName)

					Expect(result.IsError()).To(BeTrue())
					_, isValidationError := errors.AsValidationError(result.Error())
					Expect(isValidationError).To(BeTrue())
				})
			})
		})

		Describe("FindUserByEmailOption", func() {
			Context("when user exists", func() {
				It("should return Some with user", func() {
					// Create user first
					createDefaultTestUser("test-user-1")

					option := userService.FindUserByEmailOption(ctx, defaultTestEmail)

					Expect(option.IsPresent()).To(BeTrue())
					user, _ := option.Get()
					Expect(user.GetEmail().String()).To(Equal(defaultTestEmail))
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

	Describe("🧪 Edge Cases and Complex Business Logic", func() {
		Describe("Email Validation Edge Cases", func() {
			DescribeTable("should handle complex email validation scenarios",
				func(email string, shouldSucceed bool, description string) {
					id := createTestUserID("edge-case-user")
					user, err := userService.CreateUser(ctx, id, email, "Test User")

					if shouldSucceed {
						Expect(err).ToNot(HaveOccurred(), description)
						Expect(user).ToNot(BeNil(), description)
						Expect(user.GetEmail().String()).To(Equal(email), description)
					} else {
						Expect(err).To(HaveOccurred(), description)
						Expect(user).To(BeNil(), description)
						_, isValidationError := errors.AsValidationError(err)
						Expect(isValidationError).To(BeTrue(), description)
					}
				},
				Entry("Valid email with subdomain", "user@mail.example.com", true, "should accept subdomain emails"),
				Entry("Valid email with plus addressing", "user+tag@example.com", true, "should accept plus addressing"),
				Entry("Valid email with dots", "first.last@example.com", true, "should accept dots in local part"),
				Entry("Valid email with numbers", "user123@example.com", true, "should accept numbers"),
				Entry("Valid email with hyphens", "user-name@example.com", true, "should accept hyphens"),
				Entry("Invalid email - no @", "userexample.com", false, "should reject email without @"),
				Entry("Invalid email - multiple @", "user@@example.com", false, "should reject multiple @ symbols"),
				Entry("Invalid email - no domain", "user@", false, "should reject email without domain"),
				Entry("Invalid email - no local part", "@example.com", false, "should reject email without local part"),
				Entry("Invalid email - spaces", "user @example.com", false, "should reject emails with spaces"),
				Entry("Invalid email - special chars", "user<>@example.com", false, "should reject invalid special characters"),
				Entry("Invalid email - consecutive dots", "user..name@example.com", false, "should reject consecutive dots"),
				Entry("Invalid email - starts with dot", ".user@example.com", false, "should reject starting with dot"),
				Entry("Invalid email - ends with dot", "user.@example.com", false, "should reject ending with dot"),
				Entry("Invalid email - too long local", strings.Repeat("a", 65)+"@example.com", false, "should reject overly long local part"),
			)
		})

		Describe("Name Validation Edge Cases", func() {
			DescribeTable("should handle complex name validation scenarios",
				func(name string, shouldSucceed bool, description string) {
					id := createTestUserID("name-edge-case")
					user, err := userService.CreateUser(ctx, id, defaultTestEmail, name)

					if shouldSucceed {
						Expect(err).ToNot(HaveOccurred(), description)
						Expect(user).ToNot(BeNil(), description)
						Expect(user.GetUserName().String()).To(Equal(name), description)
					} else {
						Expect(err).To(HaveOccurred(), description)
						Expect(user).To(BeNil(), description)
						_, isValidationError := errors.AsValidationError(err)
						Expect(isValidationError).To(BeTrue(), description)
					}
				},
				Entry("Valid name with spaces", "John Doe", true, "should accept names with spaces"),
				Entry("Valid name with apostrophe", "O'Connor", true, "should accept apostrophes"),
				Entry("Valid name with hyphen", "Mary-Jane", true, "should accept hyphens"),
				Entry("Valid name with accents", "José", true, "should accept accented characters"),
				Entry("Valid long name", "Christopher Alexander", true, "should accept reasonably long names"),
				Entry("Invalid name - too short", "A", false, "should reject single character names"),
				Entry("Invalid name - empty", "", false, "should reject empty names"),
				Entry("Invalid name - only spaces", "   ", false, "should reject names with only spaces"),
				Entry("Invalid name - only numbers", "123", false, "should reject names with only numbers"),
				Entry("Invalid name - special chars", "John@Doe", false, "should reject invalid special characters"),
				Entry("Invalid name - excessive length", strings.Repeat("John ", 20), false, "should reject excessively long names"),
				Entry("Invalid name - leading/trailing spaces", " John Doe ", false, "should handle names with leading/trailing spaces"),
			)
		})

		Describe("Business Rule Edge Cases", func() {
			Context("duplicate user creation attempts", func() {
				It("should handle rapid concurrent creation attempts", func() {
					email := "concurrent@example.com"
					name := "Concurrent User"

					// Attempt to create the same user multiple times concurrently
					results := make(chan error, 5)

					for i := range 5 {
						go func(index int) {
							id, err := values.NewUserID(fmt.Sprintf("concurrent-user-%d", index))
							if err != nil {
								results <- err
								return
							}
							_, err = userService.CreateUser(ctx, id, email, name)
							results <- err
						}(i)
					}

					// Collect results
					var successCount, errorCount int
					for range 5 {
						err := <-results
						if err == nil {
							successCount++
						} else {
							errorCount++
						}
					}

					// Should have exactly one success and multiple conflicts
					Expect(successCount).To(Equal(1), "should create exactly one user")
					Expect(errorCount).To(Equal(4), "should reject 4 duplicate attempts")
				})
			})

			Context("user lifecycle state transitions", func() {
				It("should maintain data consistency across operations", func() {
					// Create user
					user := createDefaultTestUser("lifecycle-user")
					originalCreated := user.Created

					// Update user multiple times
					for i := range 3 {
						newEmail := fmt.Sprintf("updated%d@example.com", i)
						newName := fmt.Sprintf("Updated User %d", i)

						updatedUser, err := userService.UpdateUser(ctx, user.ID, newEmail, newName)
						Expect(err).ToNot(HaveOccurred())

						// Verify consistency
						Expect(updatedUser.ID).To(Equal(user.ID), "ID should remain constant")
						Expect(updatedUser.Created).To(BeTemporally("~", originalCreated, time.Second), "creation time should not change")
						Expect(updatedUser.Modified).To(BeTemporally(">", updatedUser.Created), "modified should be after created")

						// Update reference for next iteration
						user = updatedUser
					}

					// Verify final state
					finalUser, err := userService.GetUser(ctx, user.ID)
					Expect(err).ToNot(HaveOccurred())
					Expect(finalUser.GetEmail().String()).To(Equal("updated2@example.com"))
					Expect(finalUser.GetUserName().String()).To(Equal("Updated User 2"))
				})
			})

			Context("boundary value testing", func() {
				It("should handle minimum and maximum valid values", func() {
					// Test minimum valid name length (2 characters)
					id1 := createTestUserID("min-name-user")
					user1, err := userService.CreateUser(ctx, id1, "min@example.com", "Jo")
					Expect(err).ToNot(HaveOccurred())
					Expect(user1.GetUserName().String()).To(Equal("Jo"))

					// Test maximum reasonable email length
					longLocalPart := strings.Repeat("a", 60) // 60 chars + @example.com = 71 total
					longEmail := longLocalPart + "@example.com"
					id2 := createTestUserID("long-email-user")
					user2, err := userService.CreateUser(ctx, id2, longEmail, "Long Email User")
					Expect(err).ToNot(HaveOccurred())
					Expect(user2.GetEmail().String()).To(Equal(longEmail))
				})
			})
		})

		Describe("Functional Programming Edge Cases", func() {
			Context("FilterActiveUsers with edge cases", func() {
				It("should handle empty user set", func() {
					activeUsers, err := userService.FilterActiveUsers(ctx)

					Expect(err).ToNot(HaveOccurred())
					Expect(activeUsers).To(BeEmpty())
					Expect(activeUsers).ToNot(BeNil())
				})

				It("should handle large user sets efficiently", func() {
					// Create many users
					const numUsers = 50
					for i := range numUsers {
						createValidTestUser(
							fmt.Sprintf("bulk-user-%d", i),
							fmt.Sprintf("bulk%d@example.com", i),
							fmt.Sprintf("Bulk User %d", i),
						)
					}

					activeUsers, err := userService.FilterActiveUsers(ctx)

					Expect(err).ToNot(HaveOccurred())
					Expect(activeUsers).To(HaveLen(numUsers))

					// Verify all users are considered active (created recently)
					for _, user := range activeUsers {
						Expect(user.Created).To(BeTemporally(">=", time.Now().AddDate(0, 0, -30)))
					}
				})
			})

			Context("GetUserStats with complex scenarios", func() {
				It("should handle mixed domain statistics", func() {
					// Create users with different email domains
					createValidTestUser("user-1", "user1@gmail.com", "Gmail User")
					createValidTestUser("user-2", "user2@gmail.com", "Another Gmail User")
					createValidTestUser("user-3", "user3@yahoo.com", "Yahoo User")
					createValidTestUser("user-4", "user4@outlook.com", "Outlook User")
					createValidTestUser("user-5", "user5@company.com", "Company User")

					stats, err := userService.GetUserStats(ctx)

					Expect(err).ToNot(HaveOccurred())
					Expect(stats["total"]).To(Equal(5))
					Expect(stats["active"]).To(Equal(5))
					Expect(stats["domains"]).To(Equal(4)) // gmail, yahoo, outlook, company
					Expect(stats["avg_days_since_registration"]).To(BeNumerically(">=", 0))
					Expect(stats["avg_days_since_registration"]).To(BeNumerically("<", 1)) // All recent
				})

				It("should calculate accurate averages with precision", func() {
					// Create users and verify statistical calculations
					createValidTestUser("stats-user-1", "stats1@example.com", "Stats User 1")
					createValidTestUser("stats-user-2", "stats2@example.com", "Stats User 2")

					stats, err := userService.GetUserStats(ctx)

					Expect(err).ToNot(HaveOccurred())
					Expect(stats["total"]).To(Equal(2))

					// Verify we have the expected statistics
					Expect(stats).To(HaveKey("total"))
					Expect(stats["total"]).To(BeNumerically(">", 0))
				})
			})
		})
	})
})
