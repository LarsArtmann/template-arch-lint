package entities_test

import (
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	. "github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

var _ = ginkgo.Describe("User Entity", func() {
	ginkgo.Describe("NewUser", func() {
		ginkgo.Context("with valid parameters", func() {
			ginkgo.It("should create a new user successfully", func() {
				// Given
				id, idErr := values.NewUserID("user-123")
				gomega.Expect(idErr).To(gomega.BeNil())
				// Convert to domain values for entities.NewUser
				email := "test@example.com"
				name := "TestUser"

				// When
				user, err := NewUser(id, email, name)

				// Then
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(user).ToNot(gomega.BeNil())
				gomega.Expect(user.ID.Equals(id)).To(gomega.BeTrue())
				gomega.Expect(user.GetEmail().String()).To(gomega.Equal(email))
				gomega.Expect(user.GetUserName().String()).To(gomega.Equal(name))
				gomega.Expect(user.Created).To(gomega.BeTemporally("~", time.Now(), time.Second))
				gomega.Expect(user.Modified).To(gomega.BeTemporally("~", time.Now(), time.Second))
				gomega.Expect(user.Created).To(gomega.Equal(user.Modified))
			})

			ginkgo.It("should set timestamps correctly", func() {
				// Given
				beforeCreation := time.Now()

				// When
				user, err := NewUserFromStrings("user-123", "test@example.com", "TestUser")

				// Then
				afterCreation := time.Now()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(user.Created).To(gomega.BeTemporally(">=", beforeCreation))
				gomega.Expect(user.Created).To(gomega.BeTemporally("<=", afterCreation))
				gomega.Expect(user.Modified).To(gomega.Equal(user.Created))
			})
		})

		ginkgo.Context("with invalid parameters", func() {
			ginkgo.It("should return error when ID is empty", func() {
				// When
				user, err := NewUserFromStrings("", "test@example.com", "TestUser")

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("user ID"))
				gomega.Expect(user).To(gomega.BeNil())
			})

			ginkgo.It("should return error when email is empty", func() {
				// When
				user, err := NewUserFromStrings("user-123", "", "TestUser")

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("email"))
				gomega.Expect(user).To(gomega.BeNil())
			})

			ginkgo.It("should return error when name is empty", func() {
				// When
				user, err := NewUserFromStrings("user-123", "test@example.com", "")

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("name"))
				gomega.Expect(user).To(gomega.BeNil())
			})

			ginkgo.It("should return error when email is invalid", func() {
				// When
				user, err := NewUserFromStrings("user-123", "invalid-email", "TestUser")

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("email"))
				gomega.Expect(user).To(gomega.BeNil())
			})
		})

		ginkgo.Context("edge cases", func() {
			ginkgo.It("should reject whitespace-only inputs for ID", func() {
				// When
				user, err := NewUserFromStrings("   ", "test@example.com", "TestUser")

				// Then - Current implementation validates and rejects whitespace
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(user).To(gomega.BeNil())
			})

			ginkgo.It("should handle very long inputs gracefully", func() {
				// Given
				longString := make([]byte, 1000)
				for i := range longString {
					longString[i] = 'a'
				}
				longStringValue := string(longString)

				// When - This should fail due to validation
				user, err := NewUserFromStrings("user-123", longStringValue+"@example.com", longStringValue)

				// Then - Should fail validation for overly long email/name
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(user).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("NewUserFromStrings", func() {
		ginkgo.It("should create user from string ID", func() {
			// When
			user, err := NewUserFromStrings("user-123", "test@example.com", "TestUser")

			// Then
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(user).ToNot(gomega.BeNil())
			gomega.Expect(user.ID.String()).To(gomega.Equal("user-123"))
		})
	})

	ginkgo.Describe("Validate", func() {
		ginkgo.Context("with a valid user", func() {
			ginkgo.It("should pass validation", func() {
				// Given
				user, err := NewUserFromStrings("user-123", "test@example.com", "TestUser")
				gomega.Expect(err).To(gomega.BeNil())

				// When
				err = user.Validate()

				// Then
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid user state", func() {
			ginkgo.It("should fail validation when ID is empty", func() {
				// Given - Create user with empty ID and zero value objects
				user := &User{
					ID:       values.UserID{}, // Empty UserID
					Created:  time.Now(),
					Modified: time.Now(),
					// email and name are zero values (empty) - will fail validation
				}

				// When
				err := user.Validate()

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("user ID"))
			})

			ginkgo.It("should fail validation when email is empty", func() {
				// Given - Create user with valid ID but zero value email
				userID, _ := values.NewUserID("user-123")
				user := &User{
					ID:       userID,
					Created:  time.Now(),
					Modified: time.Now(),
					// email is zero value (empty) - will fail validation
					// name would need to be set, but we'll test email validation first
				}

				// When
				err := user.Validate()

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("email"))
			})

			ginkgo.It("should fail validation when name is empty", func() {
				// Given - Create user with valid ID and email but zero value name
				userID, _ := values.NewUserID("user-123")
				// Create user via constructor to properly set private fields
				user, err := NewUser(userID, "test@example.com", "") // Empty name will fail validation

				// When
				err = user.Validate()

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("name"))
			})
		})

		ginkgo.Context("with zero-value timestamps", func() {
			ginkgo.It("should still validate successfully (timestamps not validated)", func() {
				// Given - Create user with valid fields but zero timestamps
				userID, _ := values.NewUserID("user-123")
				// Create user via constructor then override timestamps for testing
				user, _ := NewUser(userID, "test@example.com", "TestUser")
				user.Created = time.Time{}  // zero value for testing
				user.Modified = time.Time{} // zero value for testing

				// When
				err := user.Validate()

				// Then - Current implementation doesn't validate timestamps
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("User methods", func() {
		var user *User

		ginkgo.BeforeEach(func() {
			var err error
			user, err = NewUserFromStrings("user-123", "test@example.com", "TestUser")
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.Describe("Email operations", func() {
			ginkgo.It("should get email value object", func() {
				// When
				email := user.GetEmail()

				// Then
				gomega.Expect(email.Value()).To(gomega.Equal("test@example.com"))
				gomega.Expect(email.IsEmpty()).To(gomega.BeFalse())
			})

			ginkgo.It("should set email with validation", func() {
				// When
				err := user.SetEmail("new@example.com")

				// Then
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(user.GetEmail().String()).To(gomega.Equal("new@example.com"))
			})

			ginkgo.It("should reject invalid email", func() {
				// When
				err := user.SetEmail("invalid-email")

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
			})

			ginkgo.It("should get email domain", func() {
				// When
				domain := user.EmailDomain()

				// Then
				gomega.Expect(domain).To(gomega.Equal("example.com"))
			})

			ginkgo.It("should check if email is valid", func() {
				// Then
				gomega.Expect(user.IsEmailValid()).To(gomega.BeTrue())
			})
		})

		ginkgo.Describe("Name operations", func() {
			ginkgo.It("should get username value object", func() {
				// When
				name := user.GetUserName()

				// Then
				gomega.Expect(name.Value()).To(gomega.Equal("TestUser"))
				gomega.Expect(name.IsEmpty()).To(gomega.BeFalse())
			})

			ginkgo.It("should set name with validation", func() {
				// When
				err := user.SetName("NewName")

				// Then
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(user.GetUserName().String()).To(gomega.Equal("NewName"))
			})

			ginkgo.It("should check if name is reserved", func() {
				// When
				isReserved := user.IsNameReserved()

				// Then - "TestUser" should not be reserved
				gomega.Expect(isReserved).To(gomega.BeFalse())
			})
		})
	})

	ginkgo.Describe("UserID value object", func() {
		ginkgo.It("should work as a value object", func() {
			// Given
			id, err := values.NewUserID("test-id")

			// Then
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(id.String()).To(gomega.Equal("test-id"))
			gomega.Expect(id.Value()).To(gomega.Equal("test-id"))
		})

		ginkgo.It("should be comparable", func() {
			// Given
			id1, _ := values.NewUserID("same-id")
			id2, _ := values.NewUserID("same-id")
			id3, _ := values.NewUserID("different-id")

			// Then
			gomega.Expect(id1.Equals(id2)).To(gomega.BeTrue())
			gomega.Expect(id1.Equals(id3)).To(gomega.BeFalse())
		})

		ginkgo.It("should validate format correctly", func() {
			// Valid IDs
			validIDs := []string{"user-123", "test_id", "UserID123", "a1b2c3"}
			for _, validID := range validIDs {
				id, err := values.NewUserID(validID)
				gomega.Expect(err).To(gomega.BeNil(), "Expected %s to be valid", validID)
				gomega.Expect(id.String()).To(gomega.Equal(validID))
			}

			// Invalid IDs
			invalidIDs := []string{"", "   ", "id with spaces", "id@domain", "id#hash"}
			for _, invalidID := range invalidIDs {
				_, err := values.NewUserID(invalidID)
				gomega.Expect(err).To(gomega.HaveOccurred(), "Expected %s to be invalid", invalidID)
			}
		})

		ginkgo.It("should generate unique IDs", func() {
			// When
			id1, err1 := values.GenerateUserID()
			id2, err2 := values.GenerateUserID()

			// Then
			gomega.Expect(err1).To(gomega.BeNil())
			gomega.Expect(err2).To(gomega.BeNil())
			gomega.Expect(id1.Equals(id2)).To(gomega.BeFalse())
			gomega.Expect(id1.IsGenerated()).To(gomega.BeTrue())
			gomega.Expect(id2.IsGenerated()).To(gomega.BeTrue())
		})

		ginkgo.It("should handle empty check", func() {
			// Given
			emptyID := values.UserID{}
			validID, _ := values.NewUserID("test")

			// Then
			gomega.Expect(emptyID.IsEmpty()).To(gomega.BeTrue())
			gomega.Expect(validID.IsEmpty()).To(gomega.BeFalse())
		})
	})
})
