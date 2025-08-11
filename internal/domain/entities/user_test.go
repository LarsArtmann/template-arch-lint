package entities

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

var _ = Describe("User Entity", func() {
	Describe("NewUser", func() {
		Context("with valid parameters", func() {
			It("should create a new user successfully", func() {
				// Given
				id, idErr := values.NewUserID("user-123")
				Expect(idErr).To(BeNil())
				email := "test@example.com"
				name := "TestUser"

				// When
				user, err := NewUser(id, email, name)

				// Then
				Expect(err).To(BeNil())
				Expect(user).ToNot(BeNil())
				Expect(user.ID.Equals(id)).To(BeTrue())
				Expect(user.Email).To(Equal(email))
				Expect(user.Name).To(Equal(name))
				Expect(user.Created).To(BeTemporally("~", time.Now(), time.Second))
				Expect(user.Modified).To(BeTemporally("~", time.Now(), time.Second))
				Expect(user.Created).To(Equal(user.Modified))
			})

			It("should set timestamps correctly", func() {
				// Given
				beforeCreation := time.Now()

				// When
				user, err := NewUserFromStrings("user-123", "test@example.com", "TestUser")

				// Then
				afterCreation := time.Now()
				Expect(err).To(BeNil())
				Expect(user.Created).To(BeTemporally(">=", beforeCreation))
				Expect(user.Created).To(BeTemporally("<=", afterCreation))
				Expect(user.Modified).To(Equal(user.Created))
			})
		})

		Context("with invalid parameters", func() {
			It("should return error when ID is empty", func() {
				// When
				user, err := NewUserFromStrings("", "test@example.com", "TestUser")

				// Then
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("user ID"))
				Expect(user).To(BeNil())
			})

			It("should return error when email is empty", func() {
				// When
				user, err := NewUserFromStrings("user-123", "", "TestUser")

				// Then
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("email"))
				Expect(user).To(BeNil())
			})

			It("should return error when name is empty", func() {
				// When
				user, err := NewUserFromStrings("user-123", "test@example.com", "")

				// Then
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("name"))
				Expect(user).To(BeNil())
			})

			It("should return error when email is invalid", func() {
				// When
				user, err := NewUserFromStrings("user-123", "invalid-email", "TestUser")

				// Then
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("email"))
				Expect(user).To(BeNil())
			})
		})

		Context("edge cases", func() {
			It("should reject whitespace-only inputs for ID", func() {
				// When
				user, err := NewUserFromStrings("   ", "test@example.com", "TestUser")

				// Then - Current implementation validates and rejects whitespace
				Expect(err).To(HaveOccurred())
				Expect(user).To(BeNil())
			})

			It("should handle very long inputs gracefully", func() {
				// Given
				longString := make([]byte, 1000)
				for i := range longString {
					longString[i] = 'a'
				}
				longStringValue := string(longString)

				// When - This should fail due to validation
				user, err := NewUserFromStrings("user-123", longStringValue+"@example.com", longStringValue)

				// Then - Should fail validation for overly long email/name
				Expect(err).To(HaveOccurred())
				Expect(user).To(BeNil())
			})
		})
	})

	Describe("NewUserFromStrings", func() {
		It("should create user from string ID", func() {
			// When
			user, err := NewUserFromStrings("user-123", "test@example.com", "TestUser")

			// Then
			Expect(err).To(BeNil())
			Expect(user).ToNot(BeNil())
			Expect(user.ID.String()).To(Equal("user-123"))
		})
	})

	Describe("Validate", func() {
		Context("with a valid user", func() {
			It("should pass validation", func() {
				// Given
				user, err := NewUserFromStrings("user-123", "test@example.com", "TestUser")
				Expect(err).To(BeNil())

				// When
				err = user.Validate()

				// Then
				Expect(err).To(BeNil())
			})
		})

		Context("with invalid user state", func() {
			It("should fail validation when ID is empty", func() {
				// Given - Create user with empty ID using struct literal (bypassing validation)
				user := &User{
					ID:       values.UserID{}, // Empty UserID
					Email:    "test@example.com",
					Name:     "TestUser",
					Created:  time.Now(),
					Modified: time.Now(),
				}

				// When
				err := user.Validate()

				// Then
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("user ID"))
			})

			It("should fail validation when email is empty", func() {
				// Given
				userID, _ := values.NewUserID("user-123")
				user := &User{
					ID:       userID,
					Email:    "", // Empty email
					Name:     "TestUser",
					Created:  time.Now(),
					Modified: time.Now(),
				}

				// When
				err := user.Validate()

				// Then
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("email"))
			})

			It("should fail validation when name is empty", func() {
				// Given
				userID, _ := values.NewUserID("user-123")
				user := &User{
					ID:       userID,
					Email:    "test@example.com",
					Name:     "", // Empty name
					Created:  time.Now(),
					Modified: time.Now(),
				}

				// When
				err := user.Validate()

				// Then
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("name"))
			})
		})

		Context("with zero-value timestamps", func() {
			It("should still validate successfully (timestamps not validated)", func() {
				// Given
				userID, _ := values.NewUserID("user-123")
				user := &User{
					ID:       userID,
					Email:    "test@example.com",
					Name:     "TestUser",
					Created:  time.Time{}, // zero value
					Modified: time.Time{}, // zero value
				}

				// When
				err := user.Validate()

				// Then - Current implementation doesn't validate timestamps
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("User methods", func() {
		var user *User

		BeforeEach(func() {
			var err error
			user, err = NewUserFromStrings("user-123", "test@example.com", "TestUser")
			Expect(err).To(BeNil())
		})

		Describe("Email operations", func() {
			It("should get email value object", func() {
				// When
				email := user.GetEmail()

				// Then
				Expect(email.Value()).To(Equal("test@example.com"))
				Expect(email.IsEmpty()).To(BeFalse())
			})

			It("should set email with validation", func() {
				// When
				err := user.SetEmail("new@example.com")

				// Then
				Expect(err).To(BeNil())
				Expect(user.Email).To(Equal("new@example.com"))
			})

			It("should reject invalid email", func() {
				// When
				err := user.SetEmail("invalid-email")

				// Then
				Expect(err).To(HaveOccurred())
			})

			It("should get email domain", func() {
				// When
				domain := user.EmailDomain()

				// Then
				Expect(domain).To(Equal("example.com"))
			})

			It("should check if email is valid", func() {
				// Then
				Expect(user.IsEmailValid()).To(BeTrue())
			})
		})

		Describe("Name operations", func() {
			It("should get username value object", func() {
				// When
				name := user.GetUserName()

				// Then
				Expect(name.Value()).To(Equal("TestUser"))
				Expect(name.IsEmpty()).To(BeFalse())
			})

			It("should set name with validation", func() {
				// When
				err := user.SetName("NewName")

				// Then
				Expect(err).To(BeNil())
				Expect(user.Name).To(Equal("NewName"))
			})

			It("should check if name is reserved", func() {
				// When
				isReserved := user.IsNameReserved()

				// Then - "TestUser" should not be reserved
				Expect(isReserved).To(BeFalse())
			})
		})
	})

	Describe("UserID value object", func() {
		It("should work as a value object", func() {
			// Given
			id, err := values.NewUserID("test-id")

			// Then
			Expect(err).To(BeNil())
			Expect(id.String()).To(Equal("test-id"))
			Expect(id.Value()).To(Equal("test-id"))
		})

		It("should be comparable", func() {
			// Given
			id1, _ := values.NewUserID("same-id")
			id2, _ := values.NewUserID("same-id")
			id3, _ := values.NewUserID("different-id")

			// Then
			Expect(id1.Equals(id2)).To(BeTrue())
			Expect(id1.Equals(id3)).To(BeFalse())
		})

		It("should validate format correctly", func() {
			// Valid IDs
			validIDs := []string{"user-123", "test_id", "UserID123", "a1b2c3"}
			for _, validID := range validIDs {
				id, err := values.NewUserID(validID)
				Expect(err).To(BeNil(), "Expected %s to be valid", validID)
				Expect(id.String()).To(Equal(validID))
			}

			// Invalid IDs
			invalidIDs := []string{"", "   ", "id with spaces", "id@domain", "id#hash"}
			for _, invalidID := range invalidIDs {
				_, err := values.NewUserID(invalidID)
				Expect(err).To(HaveOccurred(), "Expected %s to be invalid", invalidID)
			}
		})

		It("should generate unique IDs", func() {
			// When
			id1, err1 := values.GenerateUserID()
			id2, err2 := values.GenerateUserID()

			// Then
			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(id1.Equals(id2)).To(BeFalse())
			Expect(id1.IsGenerated()).To(BeTrue())
			Expect(id2.IsGenerated()).To(BeTrue())
		})

		It("should handle empty check", func() {
			// Given
			emptyID := values.UserID{}
			validID, _ := values.NewUserID("test")

			// Then
			Expect(emptyID.IsEmpty()).To(BeTrue())
			Expect(validID.IsEmpty()).To(BeFalse())
		})
	})
})
