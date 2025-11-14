package values_test

import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/LarsArtmann/template-arch-lint/pkg/errors"
)

func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "🛡️ Input Validation Testing Suite - Service Boundary Protection")
}

var _ = Describe("🛡️ Input Validation at Service Boundaries", func() {
	Describe("📧 Email Validation", func() {
		Context("with valid email addresses", func() {
			DescribeTable("should accept valid email formats",
				func(emailStr string, description string) {
					email, err := values.NewEmail(emailStr)
					Expect(err).ToNot(HaveOccurred(), description)
					Expect(email.String()).To(Equal(emailStr), description)
				},
				Entry("standard email", "user@example.com", "basic email format"),
				Entry("email with subdomain", "user@mail.example.com", "subdomain support"),
				Entry("email with plus addressing", "user+tag@example.com", "plus addressing support"),
				Entry("email with dots in local part", "first.last@example.com", "dots in local part"),
				Entry("email with numbers", "user123@example.com", "numeric characters"),
				Entry("email with hyphens", "user-name@example.com", "hyphen support"),
				Entry("email with underscore", "user_name@example.com", "underscore support"),
				Entry("short domain", "user@a.co", "minimal domain"),
				Entry("long local part", strings.Repeat("a", 60)+"@example.com", "maximum reasonable local part"),
				Entry("multiple subdomains", "user@mail.support.example.com", "multiple subdomain levels"),
				Entry("international domain", "user@example.co.uk", "country code domains"),
				Entry("numeric domain", "user@123.456.789.012", "numeric IP-like domain"),
			)
		})

		Context("with invalid email addresses", func() {
			DescribeTable("should reject invalid email formats",
				func(emailStr string, description string) {
					email, err := values.NewEmail(emailStr)
					Expect(email.String()).To(BeEmpty(), description)
					Expect(err).To(HaveOccurred(), description)

					_, isValidationError := errors.AsValidationError(err)
					Expect(isValidationError).To(BeTrue(), "should be validation error: %s", description)
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
				Entry("consecutive dots in local", "user..name@example.com", "consecutive dots in local part"),
				Entry("starting dot in local", ".user@example.com", "starting dot in local part"),
				Entry("ending dot in local", "user.@example.com", "ending dot in local part"),
				Entry("consecutive dots in domain", "user@example..com", "consecutive dots in domain"),
				Entry("starting dot in domain", "user@.example.com", "starting dot in domain"),
				Entry("ending dot in domain", "user@example.com.", "ending dot in domain"),
				Entry("invalid characters", "user<>@example.com", "invalid special characters"),
				Entry("brackets", "user[name]@example.com", "square brackets"),
				Entry("quotes", "user\"name\"@example.com", "quotation marks"),
				Entry("backslash", "user\\name@example.com", "backslash character"),
				Entry("pipe character", "user|name@example.com", "pipe character"),
				Entry("too long local part", strings.Repeat("a", 65)+"@example.com", "excessively long local part"),
				Entry("too long domain", "user@"+strings.Repeat("a", 250)+".com", "excessively long domain"),
				Entry("too long overall", strings.Repeat("a", 200)+"@"+strings.Repeat("b", 200)+".com", "excessively long overall"),
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
			DescribeTable("should accept valid name formats",
				func(nameStr string, description string) {
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

		Context("with invalid user names", func() {
			DescribeTable("should reject invalid name formats",
				func(nameStr string, description string) {
					userName, err := values.NewUserName(nameStr)
					Expect(userName.String()).To(BeEmpty(), description)
					Expect(err).To(HaveOccurred(), description)

					_, isValidationError := errors.AsValidationError(err)
					Expect(isValidationError).To(BeTrue(), "should be validation error: %s", description)
				},
				Entry("empty name", "", "empty string"),
				Entry("only spaces", "   ", "whitespace only"),
				Entry("single character", "A", "too short"),
				Entry("only numbers", "123", "numeric only"),
				Entry("only special chars", "@#$%", "special characters only"),
				Entry("leading spaces", " John Doe", "leading whitespace"),
				Entry("trailing spaces", "John Doe ", "trailing whitespace"),
				Entry("email format", "john@example.com", "email-like format"),
				Entry("with @ symbol", "John@Doe", "@ symbol in name"),
				Entry("with hash", "John#Doe", "hash symbol"),
				Entry("with percent", "John%Doe", "percent symbol"),
				Entry("with dollar", "John$Doe", "dollar symbol"),
				Entry("with ampersand", "John&Doe", "ampersand symbol"),
				Entry("with asterisk", "John*Doe", "asterisk symbol"),
				Entry("with plus", "John+Doe", "plus symbol"),
				Entry("with equals", "John=Doe", "equals symbol"),
				Entry("with brackets", "John[Doe]", "square brackets"),
				Entry("with parentheses", "John(Doe)", "parentheses"),
				Entry("with braces", "John{Doe}", "curly braces"),
				Entry("with pipe", "John|Doe", "pipe symbol"),
				Entry("with backslash", "John\\Doe", "backslash"),
				Entry("with forward slash", "John/Doe", "forward slash"),
				Entry("with semicolon", "John;Doe", "semicolon"),
				Entry("with colon", "John:Doe", "colon"),
				Entry("with quotes", "John\"Doe\"", "quotation marks"),
				Entry("with less than", "John<Doe", "less than symbol"),
				Entry("with greater than", "John>Doe", "greater than symbol"),
				Entry("with question mark", "John?Doe", "question mark"),
				Entry("excessively long", strings.Repeat("John ", 20), "excessively long name"),
				Entry("tab character", "John\tDoe", "tab character"),
				Entry("newline character", "John\nDoe", "newline character"),
				Entry("carriage return", "John\rDoe", "carriage return"),
			)
		})

		Context("edge cases and boundary conditions", func() {
			It("should accept minimum valid length (2 characters)", func() {
				userName, err := values.NewUserName("Jo")
				Expect(err).ToNot(HaveOccurred())
				Expect(userName.String()).To(Equal("Jo"))
			})

			It("should handle names with mixed valid characters", func() {
				userName, err := values.NewUserName("Mary-Jane O'Connor")
				Expect(err).ToNot(HaveOccurred())
				Expect(userName.String()).To(Equal("Mary-Jane O'Connor"))
			})

			It("should handle names with periods", func() {
				userName, err := values.NewUserName("Dr. John Doe")
				Expect(err).ToNot(HaveOccurred())
				Expect(userName.String()).To(Equal("Dr. John Doe"))
			})

			It("should handle names with commas", func() {
				userName, err := values.NewUserName("Doe, John")
				Expect(err).ToNot(HaveOccurred())
				Expect(userName.String()).To(Equal("Doe, John"))
			})
		})
	})

	Describe("🆔 UserID Validation", func() {
		Context("with valid user IDs", func() {
			DescribeTable("should accept valid ID formats",
				func(idStr string, description string) {
					userID, err := values.NewUserID(idStr)
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

		Context("with invalid user IDs", func() {
			DescribeTable("should reject invalid ID formats",
				func(idStr string, description string) {
					userID, err := values.NewUserID(idStr)
					Expect(userID.String()).To(BeEmpty(), description)
					Expect(err).To(HaveOccurred(), description)

					_, isValidationError := errors.AsValidationError(err)
					Expect(isValidationError).To(BeTrue(), "should be validation error: %s", description)
				},
				Entry("empty ID", "", "empty string"),
				Entry("only spaces", "   ", "whitespace only"),
				Entry("single character", "a", "too short"),
				Entry("leading spaces", " user123", "leading whitespace"),
				Entry("trailing spaces", "user123 ", "trailing whitespace"),
				Entry("with spaces", "user 123", "spaces in ID"),
				Entry("with special chars", "user@123", "special characters"),
				Entry("with hash", "user#123", "hash symbol"),
				Entry("with percent", "user%123", "percent symbol"),
				Entry("with ampersand", "user&123", "ampersand symbol"),
				Entry("with asterisk", "user*123", "asterisk symbol"),
				Entry("with plus", "user+123", "plus symbol"),
				Entry("with equals", "user=123", "equals symbol"),
				Entry("with brackets", "user[123]", "square brackets"),
				Entry("with parentheses", "user(123)", "parentheses"),
				Entry("with braces", "user{123}", "curly braces"),
				Entry("with pipe", "user|123", "pipe symbol"),
				Entry("with backslash", "user\\123", "backslash"),
				Entry("with forward slash", "user/123", "forward slash"),
				Entry("with semicolon", "user;123", "semicolon"),
				Entry("with colon", "user:123", "colon"),
				Entry("with quotes", "user\"123\"", "quotation marks"),
				Entry("with less than", "user<123", "less than symbol"),
				Entry("with greater than", "user>123", "greater than symbol"),
				Entry("with question mark", "user?123", "question mark"),
				Entry("with exclamation", "user!123", "exclamation mark"),
				Entry("with dot", "user.123", "period/dot"),
				Entry("with comma", "user,123", "comma"),
				Entry("excessively long", strings.Repeat("a", 1000), "excessively long ID"),
				Entry("tab character", "user\t123", "tab character"),
				Entry("newline character", "user\n123", "newline character"),
				Entry("carriage return", "user\r123", "carriage return"),
			)
		})

		Context("edge cases and boundary conditions", func() {
			It("should handle exactly 2 character ID (minimum boundary)", func() {
				userID, err := values.NewUserID("ab")
				Expect(err).ToNot(HaveOccurred())
				Expect(userID.String()).To(Equal("ab"))
			})

			It("should handle reasonable maximum length", func() {
				longID := strings.Repeat("a", 100)
				userID, err := values.NewUserID(longID)
				Expect(err).ToNot(HaveOccurred())
				Expect(userID.String()).To(Equal(longID))
			})

			It("should preserve case in IDs", func() {
				userID, err := values.NewUserID("UsErId123")
				Expect(err).ToNot(HaveOccurred())
				Expect(userID.String()).To(Equal("UsErId123"))
			})
		})
	})

	Describe("🔄 Cross-Validation Integration", func() {
		Context("when creating complete user data", func() {
			It("should validate all components together", func() {
				// Valid combination
				userID, err := values.NewUserID("valid-user-123")
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
				validID, err := values.NewUserID("valid-user-123")
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
				invalidID, err := values.NewUserID("")
				Expect(err).To(HaveOccurred())
				Expect(invalidID.String()).To(BeEmpty())
			})
		})
	})

	Describe("🛡️ Security Validation", func() {
		Context("injection attack prevention", func() {
			It("should reject SQL injection attempts", func() {
				maliciousInputs := []string{
					"'; DROP TABLE users; --",
					"admin'--",
					"admin' OR '1'='1",
					"' UNION SELECT * FROM users --",
					"'; INSERT INTO users VALUES ('hacker'); --",
				}

				for _, maliciousInput := range maliciousInputs {
					// Try in email
					email, err := values.NewEmail(maliciousInput)
					Expect(err).To(HaveOccurred(), "should reject SQL injection in email: %s", maliciousInput)
					Expect(email.String()).To(BeEmpty())

					// Try in name
					name, err := values.NewUserName(maliciousInput)
					Expect(err).To(HaveOccurred(), "should reject SQL injection in name: %s", maliciousInput)
					Expect(name.String()).To(BeEmpty())

					// Try in ID
					id, err := values.NewUserID(maliciousInput)
					Expect(err).To(HaveOccurred(), "should reject SQL injection in ID: %s", maliciousInput)
					Expect(id.String()).To(BeEmpty())
				}
			})

			It("should reject XSS attempts", func() {
				xssInputs := []string{
					"<script>alert('xss')</script>",
					"javascript:alert('xss')",
					"<img src=x onerror=alert('xss')>",
					"<svg onload=alert('xss')>",
					"&lt;script&gt;alert('xss')&lt;/script&gt;",
				}

				for _, xssInput := range xssInputs {
					// Try in email (will likely fail format validation anyway)
					email, err := values.NewEmail(xssInput)
					Expect(err).To(HaveOccurred(), "should reject XSS in email: %s", xssInput)
					Expect(email.String()).To(BeEmpty())

					// Try in name
					name, err := values.NewUserName(xssInput)
					Expect(err).To(HaveOccurred(), "should reject XSS in name: %s", xssInput)
					Expect(name.String()).To(BeEmpty())

					// Try in ID
					id, err := values.NewUserID(xssInput)
					Expect(err).To(HaveOccurred(), "should reject XSS in ID: %s", xssInput)
					Expect(id.String()).To(BeEmpty())
				}
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
					id, err := values.NewUserID(pathInput)
					Expect(err).To(HaveOccurred(), "should reject path traversal in ID: %s", pathInput)
					Expect(id.String()).To(BeEmpty())
				}
			})
		})
	})
})
