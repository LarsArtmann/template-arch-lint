package values_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/ids"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

var _ = Describe("Value Objects", func() {
	Describe("Email", func() {
		Describe("NewEmail", func() {
			Context("with valid email addresses", func() {
				DescribeTable("should create email successfully",
					func(email, expectedAddress, expectedDomain string) {
						emailVO, err := values.NewEmail(email)
						Expect(err).ToNot(HaveOccurred())
						Expect(emailVO.Value()).To(Equal(expectedAddress))
						Expect(emailVO.Domain()).To(Equal(expectedDomain))
					},
					Entry("simple email", "test@example.com", "test@example.com", "example.com"),
					Entry("email with dots", "first.last@example.com", "first.last@example.com", "example.com"),
					Entry("email with plus", "test+tag@example.com", "test+tag@example.com", "example.com"),
					Entry("email with numbers", "user123@test.org", "user123@test.org", "test.org"),
				)
			})

			Context("with invalid email addresses", func() {
				DescribeTable("should return validation error",
					func(email string) {
						_, err := values.NewEmail(email)
						Expect(err).To(HaveOccurred())
					},
					Entry("empty email", ""),
					Entry("no @ symbol", "testexample.com"),
					Entry("no domain", "test@"),
					Entry("no local part", "@example.com"),
					Entry("spaces", "test @example.com"),
					Entry("multiple @ symbols", "test@@example.com"),
				)
			})
		})

		Describe("Email methods", func() {
			var email values.Email

			BeforeEach(func() {
				var err error
				email, err = values.NewEmail("test@example.com")
				Expect(err).ToNot(HaveOccurred())
			})

			Describe("String", func() {
				It("should return the email address", func() {
					Expect(email.String()).To(Equal("test@example.com"))
				})
			})

			Describe("Address", func() {
				It("should return the email address", func() {
					Expect(email.Value()).To(Equal("test@example.com"))
				})
			})

			Describe("Domain", func() {
				It("should return the domain part", func() {
					Expect(email.Domain()).To(Equal("example.com"))
				})
			})

			Describe("LocalPart", func() {
				It("should return the local part", func() {
					Expect(email.LocalPart()).To(Equal("test"))
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

	Describe("UserName", func() {
		Describe("NewUserName", func() {
			Context("with valid usernames", func() {
				DescribeTable("should create username successfully",
					func(name, expected string) {
						username, err := values.NewUserName(name)
						Expect(err).ToNot(HaveOccurred())
						Expect(username.String()).To(Equal(expected))
					},
					Entry("simple name", "john", "john"),
					Entry("name with spaces", "john doe", "john doe"),
					Entry("name with hyphen", "john-doe", "john-doe"),
					Entry("name with underscore", "john_doe", "john_doe"),
					Entry("name with numbers", "john123", "john123"),
				)
			})

			Context("with invalid usernames", func() {
				DescribeTable("should return validation error",
					func(name string) {
						_, err := values.NewUserName(name)
						Expect(err).To(HaveOccurred())
					},
					Entry("empty name", ""),
					Entry("only spaces", "   "),
					Entry("special characters", "john@doe"),
				)
			})
		})

		Describe("UserName methods", func() {
			var username values.UserName

			BeforeEach(func() {
				var err error
				username, err = values.NewUserName("john doe")
				Expect(err).ToNot(HaveOccurred())
			})

			Describe("String", func() {
				It("should return the username", func() {
					Expect(username.String()).To(Equal("john doe"))
				})
			})

			Describe("IsEmpty", func() {
				It("should return false for valid username", func() {
					Expect(username.IsEmpty()).To(BeFalse())
				})

				It("should return true for empty username", func() {
					emptyUsername := values.UserName{}
					Expect(emptyUsername.IsEmpty()).To(BeTrue())
				})
			})

			Describe("IsReserved", func() {
				It("should return true for reserved names", func() {
					reserved, _ := values.NewUserName("admin")
					Expect(reserved.IsReserved()).To(BeTrue())
				})

				It("should return false for normal names", func() {
					Expect(username.IsReserved()).To(BeFalse())
				})
			})
		})
	})

	Describe("UserID", func() {
		Describe("NewUserID", func() {
			Context("with valid user IDs", func() {
				DescribeTable("should create user ID successfully",
					func(idStr, expected string) {
						userID, err := values.NewUserID(idStr)
						Expect(err).ToNot(HaveOccurred())
						Expect(userID.Get()).To(Equal(expected))
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
				Expect(userID.Get()).To(HavePrefix("user_"))
				Expect(userID.Get()).To(HaveLen(37)) // "user_" + 32 hex chars
			})

			It("should generate unique IDs", func() {
				id1, err := values.GenerateUserID()
				Expect(err).ToNot(HaveOccurred())

				id2, err := values.GenerateUserID()
				Expect(err).ToNot(HaveOccurred())

				Expect(id1.Equal(id2)).To(BeFalse())
			})

			It("should generate IDs that pass validation", func() {
				userID, err := values.GenerateUserID()
				Expect(err).ToNot(HaveOccurred())

				// Test that the generated ID can be parsed back
				parsedID, err := values.NewUserID(userID.Get())
				Expect(err).ToNot(HaveOccurred())
				Expect(parsedID.Equal(userID)).To(BeTrue())
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

			Describe("Get", func() {
				It("should return the underlying value", func() {
					Expect(userID.Get()).To(Equal("test-user-123"))
				})
			})

			Describe("Equal", func() {
				It("should return true for equal user IDs", func() {
					other, err := values.NewUserID("test-user-123")
					Expect(err).ToNot(HaveOccurred())

					Expect(userID.Equal(other)).To(BeTrue())
				})

				It("should return false for different user IDs", func() {
					other, err := values.NewUserID("different-user")
					Expect(err).ToNot(HaveOccurred())

					Expect(userID.Equal(other)).To(BeFalse())
				})
			})

			Describe("IsZero", func() {
				It("should return false for valid user ID", func() {
					Expect(userID.IsZero()).To(BeFalse())
				})

				It("should return true for zero user ID", func() {
					zeroID := ids.UserID{}
					Expect(zeroID.IsZero()).To(BeTrue())
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

			It("should marshal to JSON string", func() {
				data, err := json.Marshal(userID)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(`"test-user-123"`))
			})

			It("should unmarshal from JSON string", func() {
				data := []byte(`"test-user-123"`)
				var unmarshaled values.UserID
				err := json.Unmarshal(data, &unmarshaled)
				Expect(err).ToNot(HaveOccurred())
				Expect(unmarshaled.Get()).To(Equal("test-user-123"))
			})

			It("should handle null in JSON", func() {
				data := []byte(`null`)
				var unmarshaled values.UserID
				err := json.Unmarshal(data, &unmarshaled)
				Expect(err).ToNot(HaveOccurred())
				Expect(unmarshaled.IsZero()).To(BeTrue())
			})

			It("should marshal zero value to null", func() {
				zeroID := ids.UserID{}
				data, err := json.Marshal(zeroID)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal("null"))
			})
		})
	})
})
