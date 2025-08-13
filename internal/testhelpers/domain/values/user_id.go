// Package values provides test helpers for domain value objects.
// These helpers eliminate repetitive value object creation and validation patterns.
package values

import (
	"fmt"

	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/LarsArtmann/template-arch-lint/internal/testhelpers/base"
)

// UserIDBuilder provides a fluent API for creating test UserID value objects.
type UserIDBuilder struct {
	value         string
	useDefaults   bool
	shouldBeValid bool
	invalidType   string
}

// NewUserIDBuilder creates a new UserIDBuilder.
func NewUserIDBuilder() *UserIDBuilder {
	return &UserIDBuilder{}
}

// WithValue sets the UserID value.
func (b *UserIDBuilder) WithValue(value string) *UserIDBuilder {
	b.value = value
	return b
}

// WithDefaults sets a commonly used default value.
func (b *UserIDBuilder) WithDefaults() *UserIDBuilder {
	b.useDefaults = true
	return b
}

// WithValidData ensures the UserID has valid format.
func (b *UserIDBuilder) WithValidData() *UserIDBuilder {
	b.shouldBeValid = true
	b.useDefaults = true
	return b
}

// WithInvalidData sets invalid data for validation testing.
func (b *UserIDBuilder) WithInvalidData(invalidType string) *UserIDBuilder {
	b.invalidType = invalidType
	b.shouldBeValid = false
	return b
}

// Reset clears all configuration.
func (b *UserIDBuilder) Reset() *UserIDBuilder {
	return &UserIDBuilder{}
}

// Clone creates a copy of the builder.
func (b *UserIDBuilder) Clone() *UserIDBuilder {
	return &UserIDBuilder{
		value:         b.value,
		useDefaults:   b.useDefaults,
		shouldBeValid: b.shouldBeValid,
		invalidType:   b.invalidType,
	}
}

// Build creates the UserID with current configuration.
func (b *UserIDBuilder) Build() values.UserID {
	userID, err := b.BuildWithError()
	if b.shouldBeValid {
		Expect(err).ToNot(HaveOccurred())
	}
	return userID
}

// BuildWithError creates the UserID and returns any construction error.
func (b *UserIDBuilder) BuildWithError() (values.UserID, error) {
	b.applyDefaults()
	b.applyInvalidData()

	return values.NewUserID(b.value)
}

// applyDefaults sets default value if requested.
func (b *UserIDBuilder) applyDefaults() {
	if !b.useDefaults {
		return
	}

	if b.value == "" {
		b.value = "test-user-123"
	}
}

// applyInvalidData sets invalid data based on type.
func (b *UserIDBuilder) applyInvalidData() {
	if b.invalidType == "" {
		return
	}

	switch b.invalidType {
	case "empty":
		b.value = ""
	case "spaces":
		b.value = "invalid id with spaces"
	case "special_chars":
		b.value = "invalid@id#hash"
	case "whitespace":
		b.value = "   "
	case "unicode":
		b.value = "invalid—id"
	default:
		b.value = b.invalidType // Use the invalid type as the value
	}
}

// UserIDTestHelper provides UserID-specific testing utilities.
type UserIDTestHelper struct {
	builder *UserIDBuilder
}

// NewUserIDTestHelper creates a new UserID test helper.
func NewUserIDTestHelper() *UserIDTestHelper {
	return &UserIDTestHelper{
		builder: NewUserIDBuilder(),
	}
}

// CreateValidUserID creates a valid UserID for testing.
func (h *UserIDTestHelper) CreateValidUserID() values.UserID {
	return h.builder.Reset().WithValidData().Build()
}

// CreateValidUserIDWithValue creates a valid UserID with specific value.
func (h *UserIDTestHelper) CreateValidUserIDWithValue(value string) values.UserID {
	return h.builder.Reset().WithValue(value).Build()
}

// CreateInvalidUserID creates an invalid UserID for validation testing.
func (h *UserIDTestHelper) CreateInvalidUserID(invalidType string) (values.UserID, error) {
	return h.builder.Reset().WithInvalidData(invalidType).BuildWithError()
}

// AssertUserIDValidationError verifies UserID validation fails.
func (h *UserIDTestHelper) AssertUserIDValidationError(invalidType string) {
	userID, err := h.CreateInvalidUserID(invalidType)
	base.AssertValidationError(userID, err)
}

// AssertValidUserIDCreation verifies successful UserID creation.
func (h *UserIDTestHelper) AssertValidUserIDCreation(userID values.UserID, expectedValue string) {
	base.AssertSuccess(userID, nil)
	Expect(userID.String()).To(Equal(expectedValue))
	Expect(userID.Value()).To(Equal(expectedValue))
	Expect(userID.IsEmpty()).To(BeFalse())
}

// Convenience functions for common UserID testing patterns

// DefaultTestUserID creates a UserID with standard test value.
func DefaultTestUserID() values.UserID {
	return NewUserIDBuilder().WithDefaults().Build()
}

// TestUserIDFromString creates a UserID from string, failing fast on errors.
func TestUserIDFromString(value string) values.UserID {
	userID, err := values.NewUserID(value)
	Expect(err).ToNot(HaveOccurred())
	return userID
}

// TestUserIDSequence creates multiple UserIDs with sequential numbering.
func TestUserIDSequence(count int, prefix string) []values.UserID {
	if prefix == "" {
		prefix = "test-user"
	}

	userIDs := make([]values.UserID, count)
	for i := 0; i < count; i++ {
		value := fmt.Sprintf("%s-%d", prefix, i+1)
		userIDs[i] = TestUserIDFromString(value)
	}
	return userIDs
}

// GenerateTestUserID creates a unique generated UserID for testing.
func GenerateTestUserID() values.UserID {
	userID, err := values.GenerateUserID()
	Expect(err).ToNot(HaveOccurred())
	return userID
}

// ValidateUserIDEquals is a test helper that verifies UserID equality.
func ValidateUserIDEquals(actual, expected values.UserID) {
	Expect(actual.Equals(expected)).To(BeTrue())
	Expect(actual.String()).To(Equal(expected.String()))
}

// ValidateUserIDNotEquals is a test helper that verifies UserID inequality.
func ValidateUserIDNotEquals(actual, notExpected values.UserID) {
	Expect(actual.Equals(notExpected)).To(BeFalse())
}

// ValidateUserIDFormat is a test helper that verifies UserID format compliance.
func ValidateUserIDFormat(userID values.UserID, expectedValue string) {
	Expect(userID.String()).To(Equal(expectedValue))
	Expect(userID.Value()).To(Equal(expectedValue))
	Expect(userID.IsEmpty()).To(BeFalse())
}

// ValidateUserIDValidationError is a test helper that verifies UserID validation error.
func ValidateUserIDValidationError(userID values.UserID, err error) {
	base.AssertValidationError(userID, err)
}

// CommonInvalidUserIDValues returns a list of commonly invalid UserID values for testing.
func CommonInvalidUserIDValues() []string {
	return []string{
		"",               // empty
		"   ",            // whitespace only
		"id with spaces", // spaces
		"id@domain",      // email-like
		"id#hash",        // special characters
		"invalid—id",     // unicode dash
		"id<script>",     // HTML-like
		"id|pipe",        // pipe character
		"id\"quote",      // quote character
		"id*asterisk",    // asterisk
		"id?question",    // question mark
	}
}

// ValidateAllInvalidUserIDValues tests all common invalid UserID values.
func ValidateAllInvalidUserIDValues() {
	for _, invalidValue := range CommonInvalidUserIDValues() {
		_, err := values.NewUserID(invalidValue)
		Expect(err).To(HaveOccurred(), fmt.Sprintf("Expected %q to be invalid", invalidValue))
	}
}

// CommonValidUserIDValues returns a list of commonly valid UserID values for testing.
func CommonValidUserIDValues() []string {
	return []string{
		"user-123",
		"test_id",
		"UserID123",
		"a1b2c3",
		"simple",
		"user_test_123",
		"ID-with-dashes",
		"ID_with_underscores",
		"mixedCaseID",
		"ALL_CAPS_ID",
	}
}

// ValidateAllValidUserIDValues tests all common valid UserID values.
func ValidateAllValidUserIDValues() {
	for _, validValue := range CommonValidUserIDValues() {
		userID, err := values.NewUserID(validValue)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Expected %q to be valid", validValue))
		Expect(userID.String()).To(Equal(validValue))
	}
}
