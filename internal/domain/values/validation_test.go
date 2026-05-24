package values_test

import (
	"strings"
	"testing"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/ids"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/LarsArtmann/template-arch-lint/pkg/errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestValidation(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "🛡️ Input Validation at Service Boundaries")
}

// usernameValidationTestCase represents a test case for username validation.
type usernameValidationTestCase struct {
	name        string
	input       string
	description string
}

// userIDValidationTestCase represents a test case for user ID validation.
type userIDValidationTestCase struct {
	name        string
	input       string
	description string
}

// invalidUsernameTestCases contains test cases for invalid username validation.
//

var invalidUsernameTestCases = []usernameValidationTestCase{
	{"empty name", "", "empty string"},
	{"only spaces", "   ", "whitespace only"},
	{"single character", "A", "too short"},
	{"only numbers", "123", "numeric only"},
	{"only special chars", "@#$%", "special characters only"},
	{"leading spaces", " John Doe", "leading whitespace"},
	{"trailing spaces", "John Doe ", "trailing whitespace"},
	{"email format", "john@example.com", "email-like format"},
	{"with @ symbol", "John@Doe", "@ symbol in name"},
	{"with hash", "John#Doe", "hash symbol"},
	{"with percent", "John%Doe", "percent symbol"},
	{"with dollar", "John$Doe", "dollar symbol"},
	{"with ampersand", "John&Doe", "ampersand symbol"},
	{"with asterisk", "John*Doe", "asterisk symbol"},
	{"with plus", "John+Doe", "plus symbol"},
	{"with equals", "John=Doe", "equals symbol"},
	{"with brackets", "John[Doe]", "square brackets"},
	{"with parentheses", "John(Doe)", "parentheses"},
	{"with braces", "John{Doe}", "curly braces"},
	{"with pipe", "John|Doe", "pipe symbol"},
	{"with backslash", "John\\Doe", "backslash"},
	{"with forward slash", "John/Doe", "forward slash"},
	{"with semicolon", "John;Doe", "semicolon"},
	{"with colon", "John:Doe", "colon"},
	{"with quotes", "John\"Doe\"", "quotation marks"},
	{"with less than", "John<Doe", "less than symbol"},
	{"with greater than", "John>Doe", "greater than symbol"},
	{"with question mark", "John?Doe", "question mark"},
	{"excessively long", strings.Repeat("John ", 20), "excessively long name"},
	{"tab character", "John\tDoe", "tab character"},
	{"newline character", "John\nDoe", "newline character"},
	{"carriage return", "John\rDoe", "carriage return"},
}

// invalidUserIDTestCases contains test cases for invalid user ID validation.
//

var invalidUserIDTestCases = []userIDValidationTestCase{
	{"empty ID", "", "empty string"},
	{"only spaces", "   ", "whitespace only"},
	{"single character", "a", "too short"},
	{"leading spaces", " user123", "leading whitespace"},
	{"trailing spaces", "user123 ", "trailing whitespace"},
	{"with spaces", "user 123", "spaces in ID"},
	{"with special chars", "user@123", "special characters"},
	{"with hash", "user#123", "hash symbol"},
	{"with percent", "user%123", "percent symbol"},
	{"with ampersand", "user&123", "ampersand symbol"},
	{"with asterisk", "user*123", "asterisk symbol"},
	{"with plus", "user+123", "plus symbol"},
	{"with equals", "user=123", "equals symbol"},
	{"with brackets", "user[123]", "square brackets"},
	{"with parentheses", "user(123)", "parentheses"},
	{"with braces", "user{123}", "curly braces"},
	{"with pipe", "user|123", "pipe symbol"},
	{"with backslash", "user\\123", "backslash"},
	{"with forward slash", "user/123", "forward slash"},
	{"with semicolon", "user;123", "semicolon"},
	{"with colon", "user:123", "colon"},
	{"with quotes", "user\"123\"", "quotation marks"},
	{"with less than", "user<123", "less than symbol"},
	{"with greater than", "user>123", "greater than symbol"},
	{"with question mark", "user?123", "question mark"},
	{"with exclamation", "user!123", "exclamation mark"},
	{"with dot", "user.123", "period/dot"},
	{"with comma", "user,123", "comma"},
	{"excessively long", strings.Repeat("a", 1000), "excessively long ID"},
	{"tab character", "user\t123", "tab character"},
	{"newline character", "user\n123", "newline character"},
	{"carriage return", "user\r123", "carriage return"},
}

// runInvalidUsernameTests runs all invalid username test cases as subtests.
func runInvalidUsernameTests() {
	for _, testCase := range invalidUsernameTestCases {
		It("should reject "+testCase.name, func() {
			userName, err := values.NewUserName(testCase.input)
			Expect(userName.String()).To(BeEmpty(), testCase.description)
			Expect(err).To(HaveOccurred(), testCase.description)

			_, isValidationError := errors.AsValidationError(err)
			Expect(
				isValidationError,
			).To(BeTrue(), "should be validation error: %s", testCase.description)
		})
	}
}

// runInvalidUserIDTests runs all invalid user ID test cases as subtests.
func runInvalidUserIDTests() {
	for _, testCase := range invalidUserIDTestCases {
		It("should reject "+testCase.name, func() {
			userID, err := ids.NewUserID(testCase.input)
			Expect(userID.String()).To(BeEmpty(), testCase.description)
			Expect(err).To(HaveOccurred(), testCase.description)

			_, isValidationError := errors.AsValidationError(err)
			Expect(
				isValidationError,
			).To(BeTrue(), "should be validation error: %s", testCase.description)
		})
	}
}

var _ = Describe("🛡️ Input Validation at Service Boundaries", func() {
	Describe("📧 Email Validation", func() {
		Context("with valid email addresses", func() {
			DescribeTable(
				"should accept valid email formats",
				func(emailStr, description string) {
					email, err := values.NewEmail(emailStr)
					Expect(err).ToNot(HaveOccurred(), description)
					Expect(email.String()).To(Equal(emailStr), description)
				},
				Entry("standard email", "user@example.com", "basic email format"),
				Entry("email with subdomain", "user@mail.example.com", "subdomain support"),
				Entry(
					"email with plus addressing",
					"user+tag@example.com",
					"plus addressing support",
				),
				Entry(
					"email with dots in local part",
					"first.last@example.com",
					"dots in local part",
				),
				Entry("email with numbers", "user123@example.com", "numeric characters"),
				Entry("email with hyphens", "user-name@example.com", "hyphen support"),
				Entry("email with underscore", "user_name@example.com", "underscore support"),
				Entry("short domain", "user@a.co", "minimal domain"),
				Entry(
					"long local part",
					strings.Repeat("a", 60)+"@example.com",
					"maximum reasonable local part",
				),
				Entry(
					"multiple subdomains",
					"user@mail.support.example.com",
					"multiple subdomain levels",
				),
				Entry("international domain", "user@example.co.uk", "country code domains"),
				Entry("numeric domain", "user@123.456.789.012", "numeric IP-like domain"),
			)
		})

		Context("with invalid email addresses", func() {
			DescribeTable(
				"should reject invalid email formats",
				func(emailStr, description string) {
					email, err := values.NewEmail(emailStr)
					Expect(email.String()).To(BeEmpty(), description)
					Expect(err).To(HaveOccurred(), description)

					_, isValidationError := errors.AsValidationError(err)
					Expect(
						isValidationError,
					).To(BeTrue(), "should be validation error: %s", description)
				},
				Entry("empty email", "", "empty string"),
				Entry("only spaces", "   ", "whitespace only"),
				Entry("no @ symbol", "userexample.com", "missing @ symbol"),
				Entry("multiple @ symbols", "user@@example.com", "multiple @ symbols"),
				Entry("@ at start", "@example.com", "@ at beginning"),
				Entry("@ at end", "user@", "@ at end"),
				Entry("consecutive @", "user@@example.com", "consecutive @ symbols"),
				Entry("no local part", "@example.com", "missing local part"),
				Entry("no domain", "user@", "missing domain"),
				Entry("space in email", "user name@example.com", "space in local part"),
				Entry("space in domain", "user@example .com", "space in domain"),
				Entry("leading space", " user@example.com", "leading whitespace"),
				Entry("trailing space", "user@example.com ", "trailing whitespace"),
				Entry(
					"consecutive dots in local",
					"user..name@example.com",
					"consecutive dots in local part",
				),
				Entry("starting dot in local", ".user@example.com", "starting dot in local part"),
				Entry("ending dot in local", "user.@example.com", "ending dot in local part"),
				Entry(
					"consecutive dots in domain",
					"user@example..com",
					"consecutive dots in domain",
				),
				Entry("starting dot in domain", "user@.example.com", "starting dot in domain"),
				Entry("ending dot in domain", "user@example.com.", "ending dot in domain"),
				Entry("invalid characters", "user<>@example.com", "invalid special characters"),
				Entry("brackets", "user[name]@example.com", "square brackets"),
				Entry("quotes", "user\"name\"@example.com", "quotation marks"),
				Entry("backslash", "user\\name@example.com", "backslash character"),
				Entry("pipe character", "user|name@example.com", "pipe character"),
				Entry(
					"too long local part",
					strings.Repeat("a", 65)+"@example.com",
					"excessively long local part",
				),
				Entry(
					"too long domain",
					"user@"+strings.Repeat("a", 250)+".com",
					"excessively long domain",
				),
				Entry(
					"too long overall",
					strings.Repeat("a", 200)+"@"+strings.Repeat("b", 200)+".com",
					"excessively long overall",
				),
				Entry("unicode in local", "üser@example.com", "unicode in local part"),
				Entry("unicode in domain", "user@exämple.com", "unicode in domain"),
				Entry("tab character", "user\t@example.com", "tab character"),
				Entry("newline character", "user\n@example.com", "newline character"),
				Entry("carriage return", "user\r@example.com", "carriage return"),
			)
		})

		Context("edge cases and boundary conditions", func() {
			It("should handle exactly 64 character local part (boundary)", func() {
				localPart := strings.Repeat("a", 64)
				email, err := values.NewEmail(localPart + "@example.com")
				Expect(err).ToNot(HaveOccurred())
				Expect(email.String()).To(Equal(localPart + "@example.com"))
			})

			It("should reject 65 character local part (over boundary)", func() {
				localPart := strings.Repeat("a", 65)
				email, err := values.NewEmail(localPart + "@example.com")
				Expect(email.String()).To(BeEmpty())
				Expect(err).To(HaveOccurred())
			})

			It("should handle case sensitivity correctly", func() {
				// Email addresses should preserve case but validation should be case-insensitive for domains
				email1, err := values.NewEmail("User@Example.COM")
				Expect(err).ToNot(HaveOccurred())
				Expect(email1.String()).To(Equal("User@Example.COM"))

				email2, err := values.NewEmail("USER@EXAMPLE.COM")
				Expect(err).ToNot(HaveOccurred())
				Expect(email2.String()).To(Equal("USER@EXAMPLE.COM"))
			})
		})
	})

	Describe("👤 UserName Validation", func() {
		Context("with valid user names", func() {
			DescribeTable(
				"should accept valid name formats",
				func(nameStr, description string) {
					userName, err := values.NewUserName(nameStr)
					Expect(err).ToNot(HaveOccurred(), description)
					Expect(userName.String()).To(Equal(nameStr), description)
				},
				Entry("simple name", "John", "basic single name"),
				Entry("full name", "John Doe", "first and last name"),
				Entry("name with middle", "John Michael Doe", "first, middle, last name"),
				Entry("hyphenated name", "Mary-Jane", "hyphenated name"),
				Entry("apostrophe name", "O'Connor", "name with apostrophe"),
				Entry("accented characters", "José", "name with accents"),
				Entry("multiple spaces", "John   Doe", "multiple spaces between names"),
				Entry("long name", "Christopher Alexander", "reasonably long name"),
				Entry("name with Jr", "John Doe Jr", "name with suffix"),
				Entry("name with Roman numerals", "John Doe III", "name with Roman numerals"),
				Entry("single letter middle", "John A Doe", "single letter middle name"),
				Entry("multiple surnames", "John van der Berg", "complex surname"),
			)
		})

		Context("with invalid user names", runInvalidUsernameTests)

		Context("edge cases and boundary conditions", func() {
			DescribeTable(
				"should handle valid name edge cases",
				func(name string) {
					userName, err := values.NewUserName(name)
					Expect(err).ToNot(HaveOccurred())
					Expect(userName.String()).To(Equal(name))
				},
				Entry("minimum valid length (2 characters)", "Jo"),
				Entry("names with mixed valid characters", "Mary-Jane O'Connor"),
				Entry("names with periods", "Dr. John Doe"),
				Entry("names with commas", "Doe, John"),
			)
		})
	})

	Describe("🆔 UserID Validation", func() {
		Context("with valid user IDs", func() {
			DescribeTable(
				"should accept valid ID formats",
				func(idStr, description string) {
					userID, err := ids.NewUserID(idStr)
					Expect(err).ToNot(HaveOccurred(), description)
					Expect(userID.String()).To(Equal(idStr), description)
				},
				Entry("simple ID", "user123", "basic alphanumeric ID"),
				Entry("UUID-like", "550e8400-e29b-41d4-a716-446655440000", "UUID format"),
				Entry("with hyphens", "user-123-test", "hyphenated ID"),
				Entry("with underscores", "user_123_test", "underscore ID"),
				Entry("mixed case", "User123Test", "mixed case ID"),
				Entry("numbers only", "123456", "numeric ID"),
				Entry("letters only", "useridtest", "alphabetic ID"),
				Entry("minimum length", "ab", "minimum valid length"),
				Entry("reasonable length", "user-id-with-reasonable-length", "longer ID"),
			)
		})

		Context("with invalid user IDs", runInvalidUserIDTests)

		Context("edge cases and boundary conditions", func() {
			DescribeTable(
				"should handle valid ID edge cases",
				func(id string) {
					userID, err := ids.NewUserID(id)
					Expect(err).ToNot(HaveOccurred())
					Expect(userID.String()).To(Equal(id))
				},
				Entry("exactly 2 character ID (minimum boundary)", "ab"),
				Entry("reasonable maximum length", strings.Repeat("a", 100)),
				Entry("preserve case in IDs", "UsErId123"),
			)
		})
	})

	Describe("🔄 Cross-Validation Integration", func() {
		Context("when creating complete user data", func() {
			It("should validate all components together", func() {
				// Valid combination
				userID, err := ids.NewUserID("valid-user-123")
				Expect(err).ToNot(HaveOccurred())

				email, err := values.NewEmail("valid@example.com")
				Expect(err).ToNot(HaveOccurred())

				userName, err := values.NewUserName("John Doe")
				Expect(err).ToNot(HaveOccurred())

				// All should be valid
				Expect(userID.String()).To(Equal("valid-user-123"))
				Expect(email.String()).To(Equal("valid@example.com"))
				Expect(userName.String()).To(Equal("John Doe"))
			})

			It("should catch any invalid component in the set", func() {
				// Test that each validation is independent
				validID, err := ids.NewUserID("valid-user-123")
				Expect(err).ToNot(HaveOccurred())
				Expect(validID.String()).To(Equal("valid-user-123"))

				// Invalid email should fail
				invalidEmail, err := values.NewEmail("invalid-email")
				Expect(err).To(HaveOccurred())
				Expect(invalidEmail.String()).To(BeEmpty())

				// Invalid name should fail
				invalidName, err := values.NewUserName("")
				Expect(err).To(HaveOccurred())
				Expect(invalidName.String()).To(BeEmpty())

				// Invalid ID should fail
				invalidID, err := ids.NewUserID("")
				Expect(err).To(HaveOccurred())
				Expect(invalidID.String()).To(BeEmpty())
			})
		})
	})

	Describe("🛡️ Security Validation", func() {
		// Helper function to test that malicious inputs are rejected by all value types.
		expectSecurityInputRejection := func(inputs []string, attackType string) {
			for _, input := range inputs {
				// Try in email
				email, err := values.NewEmail(input)
				Expect(err).To(HaveOccurred(), "should reject %s in email: %s", attackType, input)
				Expect(email.String()).To(BeEmpty())

				// Try in name
				name, err := values.NewUserName(input)
				Expect(err).To(HaveOccurred(), "should reject %s in name: %s", attackType, input)
				Expect(name.String()).To(BeEmpty())

				// Try in ID
				id, err := ids.NewUserID(input)
				Expect(err).To(HaveOccurred(), "should reject %s in ID: %s", attackType, input)
				Expect(id.String()).To(BeEmpty())
			}
		}

		Context("injection attack prevention", func() {
			It("should reject SQL injection attempts", func() {
				maliciousInputs := []string{
					"'; DROP TABLE users; --",
					"admin'--",
					"admin' OR '1'='1",
					"' UNION SELECT * FROM users --",
					"'; INSERT INTO users VALUES ('hacker'); --",
				}

				expectSecurityInputRejection(maliciousInputs, "SQL injection")
			})

			It("should reject XSS attempts", func() {
				xssInputs := []string{
					"<script>alert('xss')</script>",
					"javascript:alert('xss')",
					"<img src=x onerror=alert('xss')>",
					"<svg onload=alert('xss')>",
					"&lt;script&gt;alert('xss')&lt;/script&gt;",
				}

				expectSecurityInputRejection(xssInputs, "XSS")
			})

			It("should reject path traversal attempts", func() {
				pathTraversalInputs := []string{
					"../../../etc/passwd",
					"..\\..\\..\\windows\\system32",
					"....//....//....//etc/passwd",
					"%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
				}

				for _, pathInput := range pathTraversalInputs {
					// Try in ID (most likely to be used in file paths)
					id, err := ids.NewUserID(pathInput)
					Expect(
						err,
					).To(HaveOccurred(), "should reject path traversal in ID: %s", pathInput)
					Expect(id.String()).To(BeEmpty())
				}
			})
		})
	})
})
