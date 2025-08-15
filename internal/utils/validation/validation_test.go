package validation_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/utils/validation"
)

func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validation Suite")
}

var _ = Describe("StringValidators", func() {
	var (
		validators *validation.StringValidators
		ctx        context.Context
	)

	BeforeEach(func() {
		validators = validation.NewStringValidators()
		ctx = context.Background()
	})

	Describe("NotEmpty", func() {
		It("should pass for non-empty strings", func() {
			validator := validation.ToValidator(validators.NotEmpty("field"))
			err := validator(ctx, "valid")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail for empty strings", func() {
			validator := validation.ToValidator(validators.NotEmpty("field"))
			err := validator(ctx, "")
			Expect(err).To(MatchError("field cannot be empty"))
		})

		It("should fail for whitespace-only strings", func() {
			validator := validation.ToValidator(validators.NotEmpty("field"))
			err := validator(ctx, "   \t\n   ")
			Expect(err).To(MatchError("field cannot be empty"))
		})
	})

	Describe("MinLength", func() {
		It("should pass for strings meeting minimum length", func() {
			validator := validation.ToValidator(validators.MinLength("field", 5))
			err := validator(ctx, "12345")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should pass for strings exceeding minimum length", func() {
			validator := validation.ToValidator(validators.MinLength("field", 5))
			err := validator(ctx, "123456")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail for strings below minimum length", func() {
			validator := validation.ToValidator(validators.MinLength("field", 5))
			err := validator(ctx, "1234")
			Expect(err).To(MatchError("field must be at least 5 characters long"))
		})

		It("should handle Unicode characters correctly", func() {
			validator := validation.ToValidator(validators.MinLength("field", 3))
			err := validator(ctx, "🚀🎯🔥") // 3 Unicode characters
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("MaxLength", func() {
		It("should pass for strings within maximum length", func() {
			validator := validation.ToValidator(validators.MaxLength("field", 5))
			err := validator(ctx, "12345")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail for strings exceeding maximum length", func() {
			validator := validation.ToValidator(validators.MaxLength("field", 5))
			err := validator(ctx, "123456")
			Expect(err).To(MatchError("field must not exceed 5 characters"))
		})
	})

	Describe("LengthRange", func() {
		It("should pass for strings within range", func() {
			validator := validation.ToValidator(validators.LengthRange("field", 3, 8))
			err := validator(ctx, "12345")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail for strings below range", func() {
			validator := validation.ToValidator(validators.LengthRange("field", 3, 8))
			err := validator(ctx, "12")
			Expect(err).To(MatchError("field must be between 3 and 8 characters long"))
		})

		It("should fail for strings above range", func() {
			validator := validation.ToValidator(validators.LengthRange("field", 3, 8))
			err := validator(ctx, "123456789")
			Expect(err).To(MatchError("field must be between 3 and 8 characters long"))
		})
	})

	Describe("Email", func() {
		It("should pass for valid email addresses", func() {
			validator := validation.ToValidator(validators.Email("email"))

			validEmails := []string{
				"test@example.com",
				"user.name+tag@example.co.uk",
				"admin@sub.domain.org",
			}

			for _, email := range validEmails {
				err := validator(ctx, email)
				Expect(err).ToNot(HaveOccurred(), "Expected %s to be valid", email)
			}
		})

		It("should fail for invalid email addresses", func() {
			validator := validation.ToValidator(validators.Email("email"))

			invalidEmails := []string{
				"invalid-email",
				"@example.com",
				"test@",
				"test..test@example.com",
			}

			for _, email := range invalidEmails {
				err := validator(ctx, email)
				Expect(err).To(HaveOccurred(), "Expected %s to be invalid", email)
				Expect(err.Error()).To(ContainSubstring("must be a valid email address"))
			}
		})
	})

	Describe("AlphaNumeric", func() {
		It("should pass for alphanumeric strings", func() {
			validator := validation.ToValidator(validators.AlphaNumeric("field"))

			validValues := []string{
				"abc123",
				"TestValue123",
				"123456",
				"abcdef",
			}

			for _, value := range validValues {
				err := validator(ctx, value)
				Expect(err).ToNot(HaveOccurred(), "Expected %s to be valid", value)
			}
		})

		It("should fail for strings with special characters", func() {
			validator := validation.ToValidator(validators.AlphaNumeric("field"))

			invalidValues := []string{
				"abc-123",
				"test@value",
				"hello world",
				"value!",
			}

			for _, value := range invalidValues {
				err := validator(ctx, value)
				Expect(err).To(HaveOccurred(), "Expected %s to be invalid", value)
				Expect(err.Error()).To(ContainSubstring("must contain only letters and numbers"))
			}
		})
	})
})

var _ = Describe("SanitizationFunctions", func() {
	var sanitizer *validation.SanitizationFunctions

	BeforeEach(func() {
		sanitizer = validation.NewSanitizationFunctions()
	})

	Describe("TrimWhitespace", func() {
		It("should remove leading and trailing whitespace", func() {
			result := sanitizer.TrimWhitespace("   hello world   ")
			Expect(result).To(Equal("hello world"))
		})

		It("should preserve internal whitespace", func() {
			result := sanitizer.TrimWhitespace("  hello   world  ")
			Expect(result).To(Equal("hello   world"))
		})
	})

	Describe("EscapeHTML", func() {
		It("should escape HTML special characters", func() {
			result := sanitizer.EscapeHTML("<script>alert('xss')</script>")
			Expect(result).To(Equal("&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"))
		})

		It("should escape ampersands", func() {
			result := sanitizer.EscapeHTML("Tom & Jerry")
			Expect(result).To(Equal("Tom &amp; Jerry"))
		})
	})

	Describe("RemoveExtraSpaces", func() {
		It("should collapse multiple spaces into single spaces", func() {
			result := sanitizer.RemoveExtraSpaces("hello    world")
			Expect(result).To(Equal("hello world"))
		})

		It("should handle mixed whitespace characters", func() {
			result := sanitizer.RemoveExtraSpaces("hello\t\n   world")
			Expect(result).To(Equal("hello world"))
		})

		It("should trim leading and trailing spaces", func() {
			result := sanitizer.RemoveExtraSpaces("   hello world   ")
			Expect(result).To(Equal("hello world"))
		})
	})

	Describe("Truncate", func() {
		It("should truncate long strings with ellipsis", func() {
			result := sanitizer.Truncate("This is a very long string", 10)
			Expect(result).To(Equal("This is..."))
		})

		It("should not modify short strings", func() {
			result := sanitizer.Truncate("Short", 10)
			Expect(result).To(Equal("Short"))
		})

		It("should handle edge case with maxLen <= 3", func() {
			result := sanitizer.Truncate("Hello", 3)
			Expect(result).To(Equal("Hel"))
		})

		It("should handle Unicode characters correctly", func() {
			result := sanitizer.Truncate("🚀🎯🔥⭐🌟", 3)
			Expect(result).To(Equal("🚀🎯🔥"))
		})
	})

	Describe("StripNonPrintable", func() {
		It("should remove non-printable characters", func() {
			input := "Hello\x00\x01World\x7F"
			result := sanitizer.StripNonPrintable(input)
			Expect(result).To(Equal("HelloWorld"))
		})

		It("should preserve printable characters and spaces", func() {
			input := "Hello World!\t\n"
			result := sanitizer.StripNonPrintable(input)
			Expect(result).To(Equal("Hello World!\t\n"))
		})
	})

	Describe("NormalizeWhitespace", func() {
		It("should convert all whitespace to regular spaces", func() {
			input := "Hello\tWorld\nHow\rare\tyou?"
			result := sanitizer.NormalizeWhitespace(input)
			Expect(result).To(Equal("Hello World How are you?"))
		})

		It("should remove extra spaces after normalization", func() {
			input := "Hello\t\t\nWorld"
			result := sanitizer.NormalizeWhitespace(input)
			Expect(result).To(Equal("Hello World"))
		})
	})
})

var _ = Describe("Chain", func() {
	var (
		validators *validation.StringValidators
		ctx        context.Context
	)

	BeforeEach(func() {
		validators = validation.NewStringValidators()
		ctx = context.Background()
	})

	It("should pass when all validators pass", func() {
		chainedValidator := validation.Chain(
			validation.ToValidator(validators.NotEmpty("field")),
			validation.ToValidator(validators.MinLength("field", 3)),
			validation.ToValidator(validators.MaxLength("field", 10)),
		)

		err := chainedValidator(ctx, "hello")
		Expect(err).ToNot(HaveOccurred())
	})

	It("should fail on first failing validator", func() {
		chainedValidator := validation.Chain(
			validation.ToValidator(validators.NotEmpty("field")),
			validation.ToValidator(validators.MinLength("field", 10)), // This will fail
			validation.ToValidator(validators.MaxLength("field", 20)),
		)

		err := chainedValidator(ctx, "short")
		Expect(err).To(MatchError("field must be at least 10 characters long"))
	})
})

var _ = Describe("ValidationResult", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("should create valid result when no errors", func() {
		validators := validation.NewStringValidators()
		validator := validation.ToValidator(validators.NotEmpty("field"))

		result := validation.ValidateWithResult(ctx, "valid", validator)

		Expect(result.IsValid()).To(BeTrue())
		Expect(result.HasErrors()).To(BeFalse())
		Expect(result.FirstError()).To(BeEmpty())
		Expect(result.AllErrors()).To(BeEmpty())
	})

	It("should create invalid result with errors", func() {
		validators := validation.NewStringValidators()
		validator1 := validation.ToValidator(validators.NotEmpty("field1"))
		validator2 := validation.ToValidator(validators.MinLength("field2", 5))

		result := validation.ValidateWithResult(ctx, "", validator1, validator2)

		Expect(result.IsValid()).To(BeFalse())
		Expect(result.HasErrors()).To(BeTrue())
		Expect(result.FirstError()).To(Equal("field1 cannot be empty"))
		Expect(result.AllErrors()).To(ContainSubstring("field1 cannot be empty"))
		Expect(result.AllErrors()).To(ContainSubstring("field2 must be at least 5 characters long"))
	})
})

var _ = Describe("PrebuiltValidators", func() {
	var (
		prebuilt *validation.PrebuiltValidators
		ctx      context.Context
	)

	BeforeEach(func() {
		prebuilt = validation.NewPrebuiltValidators()
		ctx = context.Background()
	})

	Describe("Username", func() {
		It("should pass for valid usernames", func() {
			validator := validation.ToValidator(prebuilt.Username("username"))

			validUsernames := []string{
				"john_doe",
				"user123",
				"test-user",
				"MyUsername",
			}

			for _, username := range validUsernames {
				err := validator(ctx, username)
				Expect(err).ToNot(HaveOccurred(), "Expected %s to be valid", username)
			}
		})

		It("should fail for invalid usernames", func() {
			validator := validation.ToValidator(prebuilt.Username("username"))

			invalidUsernames := []string{
				"us", // too short
				"this_username_is_way_too_long_to_be_valid", // too long
				"user@name", // invalid character
				"user name", // space not allowed
				"",          // empty
			}

			for _, username := range invalidUsernames {
				err := validator(ctx, username)
				Expect(err).To(HaveOccurred(), "Expected %s to be invalid", username)
			}
		})
	})

	Describe("SimplePassword", func() {
		It("should pass for valid simple passwords", func() {
			validator := validation.ToValidator(prebuilt.SimplePassword("password"))

			validPasswords := []string{
				"password123",
				"myPass123",
				"Test123456",
			}

			for _, password := range validPasswords {
				err := validator(ctx, password)
				Expect(err).ToNot(HaveOccurred(), "Expected %s to be valid", password)
			}
		})

		It("should fail for invalid simple passwords", func() {
			validator := validation.ToValidator(prebuilt.SimplePassword("password"))

			invalidPasswords := []string{
				"short1",      // too short
				"onlyletters", // no numbers
				"12345678",    // no letters
				"",            // empty
			}

			for _, password := range invalidPasswords {
				err := validator(ctx, password)
				Expect(err).To(HaveOccurred(), "Expected %s to be invalid", password)
			}
		})
	})

	Describe("UUID", func() {
		It("should pass for valid UUIDs", func() {
			validator := validation.ToValidator(prebuilt.UUID("id"))

			validUUIDs := []string{
				"123e4567-e89b-42d3-a456-426614174000",
				"550e8400-e29b-41d4-a716-446655440000",
			}

			for _, uuid := range validUUIDs {
				err := validator(ctx, uuid)
				Expect(err).ToNot(HaveOccurred(), "Expected %s to be valid", uuid)
			}
		})

		It("should fail for invalid UUIDs", func() {
			validator := validation.ToValidator(prebuilt.UUID("id"))

			invalidUUIDs := []string{
				"123e4567-e89b-12d3-a456-426614174000", // wrong version
				"123e4567-e89b-42d3-a456-42661417400",  // too short
				"not-a-uuid",
				"",
			}

			for _, uuid := range invalidUUIDs {
				err := validator(ctx, uuid)
				Expect(err).To(HaveOccurred(), "Expected %s to be invalid", uuid)
			}
		})
	})
})
