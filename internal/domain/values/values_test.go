package values_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

// TestValues is commented out to avoid Ginkgo conflict with TestValidation in validation_test.go
// Both files define their own test suites, but Ginkgo only allows one RunSpecs per package
// The validation_test.go contains the comprehensive test suite that covers all validation cases
// func TestValues(t *testing.T) {
//	RegisterFailHandler(Fail)
//	RunSpecs(t, "Values Suite")
// }

var _ = Describe("Email", func() {
	Describe("NewEmail", func() {
		Context("with valid email addresses", func() {
			DescribeTable("should create email successfully",
				func(emailStr string, expectedNormalized string) {
					email, err := values.NewEmail(emailStr)
					Expect(err).ToNot(HaveOccurred())
					Expect(email.String()).To(Equal(expectedNormalized))
				},
				Entry("simple email", "test@example.com", "test@example.com"),
				Entry("email with subdomain", "user@mail.example.com", "user@mail.example.com"),
				Entry("email with numbers", "user123@example123.com", "user123@example123.com"),
				Entry("email with special chars", "user.name+tag@example.com", "user.name+tag@example.com"),
				Entry("uppercase email", "TEST@EXAMPLE.COM", "TEST@EXAMPLE.COM"),
			)
		})

		Context("with invalid email addresses", func() {
			DescribeTable("should return validation error",
				func(emailStr string) {
					_, err := values.NewEmail(emailStr)
					Expect(err).To(HaveOccurred())
				},
				Entry("empty email", ""),
				Entry("email without @", "testexample.com"),
				Entry("email without domain", "test@"),
				Entry("email without local part", "@example.com"),
				Entry("email with spaces", "test @example.com"),
				Entry("email with leading whitespace", "  test@example.com"),
				Entry("email with trailing whitespace", "test@example.com  "),
				Entry("email with multiple @", "test@example@com"),
				Entry("email without TLD", "test@example"),
				Entry("email starting with dot", ".test@example.com"),
				Entry("email ending with dot", "test@example.com."),
				Entry("email with consecutive dots", "test..user@example.com"),
				Entry("too short email", "a@b"),
				Entry("local part too long", "verylonglocapartthatexceedsthemaximumlengthof64charactersallowedwhichshouldfail@example.com"),
			)
		})

		Context("with edge cases", func() {
			It("should handle minimum valid email", func() {
				email, err := values.NewEmail("a@b.co")
				Expect(err).ToNot(HaveOccurred())
				Expect(email.String()).To(Equal("a@b.co"))
			})

			It("should reject email that's too long", func() {
				// Create an email longer than 254 characters
				longLocal := "verylonglocapartthatexceedsthemaximumlengthof64charactersallowedverylonglocapartthatexceedsthemaximumlengthof64charactersallowedverylonglocapartthatexceedsthemaximumlengthof64charactersallowedverylonglocapartthatexceedsthemaximumlengthof64charactersallowed"
				longEmail := longLocal + "@example.com"

				_, err := values.NewEmail(longEmail)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("too long"))
			})
		})
	})

	Describe("Email methods", func() {
		var email values.Email

		BeforeEach(func() {
			var err error
			email, err = values.NewEmail("user.name@example.com")
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("String", func() {
			It("should return the string representation", func() {
				Expect(email.String()).To(Equal("user.name@example.com"))
			})
		})

		Describe("Value", func() {
			It("should return the email value", func() {
				Expect(email.Value()).To(Equal("user.name@example.com"))
			})
		})

		Describe("Domain", func() {
			It("should return the domain part", func() {
				Expect(email.Domain()).To(Equal("example.com"))
			})
		})

		Describe("LocalPart", func() {
			It("should return the local part", func() {
				Expect(email.LocalPart()).To(Equal("user.name"))
			})
		})

		Describe("Equals", func() {
			It("should return true for equal emails", func() {
				other, err := values.NewEmail("user.name@example.com")
				Expect(err).ToNot(HaveOccurred())

				Expect(email.Equals(other)).To(BeTrue())
			})

			It("should return false for different emails", func() {
				other, err := values.NewEmail("different@example.com")
				Expect(err).ToNot(HaveOccurred())

				Expect(email.Equals(other)).To(BeFalse())
			})

			It("should handle case insensitive comparison", func() {
				upper, err := values.NewEmail("USER.NAME@EXAMPLE.COM")
				Expect(err).ToNot(HaveOccurred())

				Expect(email.Equals(upper)).To(BeTrue())
			})
		})

		Describe("IsEmpty", func() {
			It("should return false for valid email", func() {
				Expect(email.IsEmpty()).To(BeFalse())
			})

			It("should return true for empty email", func() {
				emptyEmail := values.Email{}
				Expect(emptyEmail.IsEmpty()).To(BeTrue())
			})
		})
	})
})

var _ = Describe("UserID", func() {
	Describe("NewUserID", func() {
		Context("with valid user IDs", func() {
			DescribeTable("should create user ID successfully",
				func(idStr string, expected string) {
					userID, err := values.NewUserID(idStr)
					Expect(err).ToNot(HaveOccurred())
					Expect(userID.String()).To(Equal(expected))
				},
				Entry("simple ID", "user123", "user123"),
				Entry("ID with hyphens", "user-123", "user-123"),
				Entry("ID with underscores", "user_123", "user_123"),
				Entry("mixed case", "User123", "User123"),
				Entry("numeric ID", "123456", "123456"),
			)
		})

		Context("with invalid user IDs", func() {
			DescribeTable("should return validation error",
				func(idStr string) {
					_, err := values.NewUserID(idStr)
					Expect(err).To(HaveOccurred())
				},
				Entry("empty ID", ""),
				Entry("ID with spaces", "user 123"),
				Entry("ID with tab", "user\t123"),
				Entry("ID with newline", "user\n123"),
				Entry("ID with carriage return", "user\r123"),
				Entry("ID with special chars", "user@123"),
				Entry("ID with dots", "user.123"),
				Entry("ID with slashes", "user/123"),
				Entry("ID with backslashes", "user\\123"),
				Entry("ID with leading/trailing spaces", "  user123  "),
			)
		})

		Context("with edge cases", func() {
			It("should reject single character ID (too short)", func() {
				_, err := values.NewUserID("a")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("too short"))
			})

			It("should reject ID that's too long", func() {
				// Create an ID longer than 100 characters
				longID := "verylonguserthatexceedsthemaximumlengthof100charactersallowedinthesystemforidentificationwhichshouldfail"
				_, err := values.NewUserID(longID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("too long"))
			})

			It("should reject ID with only whitespace after trimming", func() {
				_, err := values.NewUserID("   ")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GenerateUserID", func() {
		It("should generate a valid user ID", func() {
			userID, err := values.GenerateUserID()
			Expect(err).ToNot(HaveOccurred())
			Expect(userID.String()).To(HavePrefix("user_"))
			Expect(userID.String()).To(HaveLen(37)) // "user_" + 32 hex chars
		})

		It("should generate unique IDs", func() {
			id1, err := values.GenerateUserID()
			Expect(err).ToNot(HaveOccurred())

			id2, err := values.GenerateUserID()
			Expect(err).ToNot(HaveOccurred())

			Expect(id1.Equals(id2)).To(BeFalse())
		})

		It("should generate IDs that pass validation", func() {
			userID, err := values.GenerateUserID()
			Expect(err).ToNot(HaveOccurred())

			// Test that the generated ID can be parsed back
			parsedID, err := values.NewUserID(userID.String())
			Expect(err).ToNot(HaveOccurred())
			Expect(parsedID.Equals(userID)).To(BeTrue())
		})
	})

	Describe("UserID methods", func() {
		var userID values.UserID

		BeforeEach(func() {
			var err error
			userID, err = values.NewUserID("test-user-123")
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("String", func() {
			It("should return the string representation", func() {
				Expect(userID.String()).To(Equal("test-user-123"))
			})
		})

		Describe("StringValue", func() {
			It("should return the user ID value", func() {
				Expect(userID.StringValue()).To(Equal("test-user-123"))
			})
		})

		Describe("Equals", func() {
			It("should return true for equal user IDs", func() {
				other, err := values.NewUserID("test-user-123")
				Expect(err).ToNot(HaveOccurred())

				Expect(userID.Equals(other)).To(BeTrue())
			})

			It("should return false for different user IDs", func() {
				other, err := values.NewUserID("different-user")
				Expect(err).ToNot(HaveOccurred())

				Expect(userID.Equals(other)).To(BeFalse())
			})
		})

		Describe("IsEmpty", func() {
			It("should return false for valid user ID", func() {
				Expect(userID.IsEmpty()).To(BeFalse())
			})

			It("should return true for empty user ID", func() {
				emptyID := values.UserID{}
				Expect(emptyID.IsEmpty()).To(BeTrue())
			})
		})

		Describe("IsGenerated", func() {
			It("should return false for manual user ID", func() {
				Expect(userID.IsGenerated()).To(BeFalse())
			})

			It("should return true for generated user ID", func() {
				generated, err := values.GenerateUserID()
				Expect(err).ToNot(HaveOccurred())
				Expect(generated.IsGenerated()).To(BeTrue())
			})
		})
	})

	Describe("JSON marshaling", func() {
		var userID values.UserID

		BeforeEach(func() {
			var err error
			userID, err = values.NewUserID("test-user-123")
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("MarshalJSON", func() {
			It("should marshal to JSON string", func() {
				data, err := json.Marshal(userID)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(`"test-user-123"`))
			})
		})

		Describe("UnmarshalJSON", func() {
			It("should unmarshal from JSON string", func() {
				data := []byte(`"test-user-123"`)
				var unmarshaled values.UserID

				err := json.Unmarshal(data, &unmarshaled)
				Expect(err).ToNot(HaveOccurred())
				Expect(unmarshaled.Equals(userID)).To(BeTrue())
			})

			It("should return error for invalid JSON", func() {
				data := []byte(`invalid-json`)
				var unmarshaled values.UserID

				err := json.Unmarshal(data, &unmarshaled)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for invalid user ID in JSON", func() {
				data := []byte(`"invalid user id"`)
				var unmarshaled values.UserID

				err := json.Unmarshal(data, &unmarshaled)
				Expect(err).To(HaveOccurred())
			})
		})

		Describe("round trip", func() {
			It("should preserve value through marshal/unmarshal", func() {
				data, err := json.Marshal(userID)
				Expect(err).ToNot(HaveOccurred())

				var unmarshaled values.UserID
				err = json.Unmarshal(data, &unmarshaled)
				Expect(err).ToNot(HaveOccurred())

				Expect(unmarshaled.Equals(userID)).To(BeTrue())
			})
		})
	})

	Describe("Database scanning", func() {
		var userID values.UserID

		BeforeEach(func() {
			var err error
			userID, err = values.NewUserID("test-user-123")
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("Value", func() {
			It("should return driver value", func() {
				value, err := userID.Value()
				Expect(err).ToNot(HaveOccurred())
				Expect(value).To(Equal("test-user-123"))
			})
		})

		Describe("Scan", func() {
			It("should scan from string value", func() {
				var scanned values.UserID
				err := scanned.Scan("test-user-123")
				Expect(err).ToNot(HaveOccurred())
				Expect(scanned.Equals(userID)).To(BeTrue())
			})

			It("should scan from byte slice", func() {
				var scanned values.UserID
				err := scanned.Scan([]byte("test-user-123"))
				Expect(err).ToNot(HaveOccurred())
				Expect(scanned.Equals(userID)).To(BeTrue())
			})

			It("should handle nil value", func() {
				var scanned values.UserID
				err := scanned.Scan(nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(scanned.IsEmpty()).To(BeTrue())
			})

			It("should return error for unsupported type", func() {
				var scanned values.UserID
				err := scanned.Scan(123)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cannot scan"))
			})

			It("should return error for invalid user ID", func() {
				var scanned values.UserID
				err := scanned.Scan("invalid user id")
				Expect(err).To(HaveOccurred())
			})
		})

		Describe("round trip", func() {
			It("should preserve value through Value/Scan", func() {
				value, err := userID.Value()
				Expect(err).ToNot(HaveOccurred())

				var scanned values.UserID
				err = scanned.Scan(value)
				Expect(err).ToNot(HaveOccurred())

				Expect(scanned.Equals(userID)).To(BeTrue())
			})
		})
	})
})
