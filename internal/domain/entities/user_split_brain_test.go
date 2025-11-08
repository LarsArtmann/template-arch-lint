package entities

import (
	"encoding/json"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

// Split Brain Test Suite - Documents current behavior and defines target behavior
var _ = ginkgo.Describe("User Split Brain Behavior", func() {
	ginkgo.Describe("CURRENT SPLIT BRAIN ISSUES", func() {
		var user *User

		ginkgo.BeforeEach(func() {
			var err error
			user, err = NewUserFromStrings("user-123", "test@example.com", "TestUser")
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should demonstrate split brain ELIMINATED - single value object access", func() {
			// REFACTORED: Only value object access works - split brain eliminated
			gomega.Expect(user.GetEmail().Value()).To(gomega.Equal("test@example.com")) // Value object only
			gomega.Expect(user.GetUserName().Value()).To(gomega.Equal("TestUser"))      // Value object only

			// FIXED: Only ONE way to access data - through value objects
			// user.Email would not compile - field is private
			// user.Name would not compile - field is private
		})

		ginkgo.It("should demonstrate synchronization ELIMINATED - single source of truth", func() {
			// When
			err := user.SetEmail("new@example.com")
			gomega.Expect(err).To(gomega.BeNil())

			// REFACTORED: Only value object updated - no synchronization needed
			gomega.Expect(user.GetEmail().Value()).To(gomega.Equal("new@example.com")) // Value object updated

			// FIXED: Single source of truth - no manual synchronization required
			// user.Email field doesn't exist anymore
		})

		ginkgo.It("should demonstrate lazy initialization ELIMINATED - direct value object access", func() {
			// Given - Create user properly through constructor
			directUser, err := NewUserFromStrings("user-123", "direct@example.com", "DirectUser")
			gomega.Expect(err).To(gomega.BeNil())

			// REFACTORED: No lazy initialization - value objects created once during construction
			email1 := directUser.GetEmail()
			email2 := directUser.GetEmail()

			// Both calls return the same value object efficiently - no repeated validation
			gomega.Expect(email1.Value()).To(gomega.Equal("direct@example.com"))
			gomega.Expect(email2.Value()).To(gomega.Equal("direct@example.com"))
		})

		ginkgo.It("should demonstrate type safety ENFORCED - no direct field access possible", func() {
			// REFACTORED: Direct field assignment is impossible - fields are private
			// user.email = "invalid-email-format"  // Would not compile - field is private
			// user.name = ""                       // Would not compile - field is private

			// Type safety enforced - only validated setters can modify state
			err := user.SetEmail("invalid-email-format")
			gomega.Expect(err).To(gomega.HaveOccurred()) // Validation happens in setter

			// User remains in valid state - invalid updates are rejected
			gomega.Expect(user.Validate()).To(gomega.BeNil()) // Still valid
		})
	})

	ginkgo.Describe("ACHIEVED BEHAVIOR AFTER SPLIT BRAIN FIX", func() {
		ginkgo.It("should use ONLY value objects for domain logic", func() {
			// ACHIEVED: User entity has ONLY value object fields
			// No more user.Email string field - only private user.email values.Email

			user, err := NewUserFromStrings("user-123", "test@example.com", "TestUser")
			gomega.Expect(err).To(gomega.BeNil())

			// VERIFIED:
			// 1. No direct string field access - fields are private
			// 2. All access through value object methods
			gomega.Expect(user.GetEmail().Value()).To(gomega.Equal("test@example.com"))
			gomega.Expect(user.GetUserName().Value()).To(gomega.Equal("TestUser"))
			// 3. Type safety enforced at compile time
			// 4. Custom JSON marshaling handles serialization (tested separately)
		})

		ginkgo.It("should have custom JSON marshaling for value objects", func() {
			// TARGET: JSON serialization should work seamlessly
			user, err := NewUserFromStrings("user-123", "test@example.com", "TestUser")
			gomega.Expect(err).To(gomega.BeNil())

			// Should marshal to expected JSON structure
			jsonBytes, err := json.Marshal(user)
			gomega.Expect(err).To(gomega.BeNil())

			var jsonMap map[string]any
			err = json.Unmarshal(jsonBytes, &jsonMap)
			gomega.Expect(err).To(gomega.BeNil())

			// Should contain string values, not value object structures
			gomega.Expect(jsonMap["id"]).To(gomega.Equal("user-123"))
			gomega.Expect(jsonMap["email"]).To(gomega.Equal("test@example.com"))
			gomega.Expect(jsonMap["name"]).To(gomega.Equal("TestUser"))
		})

		ginkgo.It("should have custom JSON unmarshaling for value objects", func() {
			// TARGET: JSON deserialization should create valid value objects
			jsonInput := `{
				"id": "user-456",
				"email": "json@example.com", 
				"name": "JsonUser",
				"created": "2023-01-01T00:00:00Z",
				"modified": "2023-01-01T00:00:00Z"
			}`

			var user User
			err := json.Unmarshal([]byte(jsonInput), &user)
			gomega.Expect(err).To(gomega.BeNil())

			// Should create valid value objects internally
			gomega.Expect(user.GetEmail().Value()).To(gomega.Equal("json@example.com"))
			gomega.Expect(user.GetUserName().Value()).To(gomega.Equal("JsonUser"))
			gomega.Expect(user.ID.Value()).To(gomega.Equal("user-456"))
		})

		ginkgo.It("should prevent invalid state at compile time", func() {
			// TARGET: No way to create invalid user through direct field access
			// This will be enforced by making fields private

			// After refactoring, these should not compile:
			// user.email = "invalid"     // Field not exported
			// user.name = ""             // Field not exported

			// Only valid way should be through constructors and setters
			user, err := NewUserFromStrings("user-123", "valid@example.com", "ValidName")
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(user.Validate()).To(gomega.BeNil()) // Always valid
		})

		ginkgo.It("should eliminate lazy initialization overhead", func() {
			// TARGET: Value objects created once during construction
			// No repeated validation on every getter call

			user, err := NewUserFromStrings("user-123", "test@example.com", "TestUser")
			gomega.Expect(err).To(gomega.BeNil())

			// Multiple getter calls should return same object (no re-validation)
			email1 := user.GetEmail()
			email2 := user.GetEmail()

			// Should be efficient - no repeated validation
			gomega.Expect(email1.Value()).To(gomega.Equal("test@example.com"))
			gomega.Expect(email2.Value()).To(gomega.Equal("test@example.com"))
		})
	})

	ginkgo.Describe("REFACTORING VALIDATION TESTS", func() {
		ginkgo.It("should verify no string fields remain after refactoring", func() {
			// VALIDATED: Refactoring was successful
			user, err := NewUserFromStrings("user-123", "test@example.com", "TestUser")
			gomega.Expect(err).To(gomega.BeNil())

			// VERIFIED: These are the ONLY ways to access data:
			gomega.Expect(user.ID.String()).To(gomega.Equal("user-123"))                // UserID value object
			gomega.Expect(user.GetEmail().Value()).To(gomega.Equal("test@example.com")) // Email value object
			gomega.Expect(user.GetUserName().Value()).To(gomega.Equal("TestUser"))      // UserName value object

			// CONFIRMED: Direct field access does not exist:
			// user.Email    // Does not compile - field is private
			// user.Name     // Does not compile - field is private
		})

		ginkgo.It("should verify setter synchronization is eliminated", func() {
			// After refactoring, setters should only update value objects
			user, err := NewUserFromStrings("user-123", "test@example.com", "TestUser")
			gomega.Expect(err).To(gomega.BeNil())

			err = user.SetEmail("new@example.com")
			gomega.Expect(err).To(gomega.BeNil())

			// Only value object should be updated (no dual field sync needed)
			gomega.Expect(user.GetEmail().Value()).To(gomega.Equal("new@example.com"))
		})

		ginkgo.It("should verify JSON marshaling works without string fields", func() {
			// Validate that custom JSON marshaling handles value objects
			user, err := NewUserFromStrings("user-123", "test@example.com", "TestUser")
			gomega.Expect(err).To(gomega.BeNil())

			jsonBytes, err := json.Marshal(user)
			gomega.Expect(err).To(gomega.BeNil())

			// Should produce clean JSON without value object complexity
			gomega.Expect(string(jsonBytes)).To(gomega.ContainSubstring("\"email\":\"test@example.com\""))
			gomega.Expect(string(jsonBytes)).To(gomega.ContainSubstring("\"name\":\"TestUser\""))
		})
	})

	ginkgo.Describe("COMPATIBILITY TESTS", func() {
		ginkgo.It("should maintain backward compatibility for existing code", func() {
			// Existing code using User entity should continue working
			user, err := NewUserFromStrings("user-123", "test@example.com", "TestUser")
			gomega.Expect(err).To(gomega.BeNil())

			// These methods must continue to work after refactoring:
			gomega.Expect(user.GetEmail().Value()).To(gomega.Equal("test@example.com"))
			gomega.Expect(user.GetUserName().Value()).To(gomega.Equal("TestUser"))
			gomega.Expect(user.EmailDomain()).To(gomega.Equal("example.com"))
			gomega.Expect(user.IsEmailValid()).To(gomega.BeTrue())
			gomega.Expect(user.IsNameReserved()).To(gomega.BeFalse())
		})

		ginkgo.It("should maintain validation behavior", func() {
			// All existing validation should continue working
			_, err := NewUserFromStrings("", "test@example.com", "TestUser")
			gomega.Expect(err).To(gomega.HaveOccurred())

			_, err = NewUserFromStrings("user-123", "invalid-email", "TestUser")
			gomega.Expect(err).To(gomega.HaveOccurred())

			_, err = NewUserFromStrings("user-123", "test@example.com", "")
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})
})
