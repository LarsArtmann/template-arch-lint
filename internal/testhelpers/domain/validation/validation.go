// Package validation provides comprehensive validation testing utilities.
// These helpers eliminate repetitive validation testing patterns across the codebase.
package validation

import (
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
	"github.com/LarsArtmann/template-arch-lint/internal/testhelpers/base"
)

// ValidationTester provides comprehensive validation testing functionality.
type ValidationTester struct {
	entityName string
}

// NewValidationTester creates a new validation tester for an entity type.
func NewValidationTester(entityName string) *ValidationTester {
	return &ValidationTester{
		entityName: entityName,
	}
}

// TestValidationSuccess verifies that validation passes for valid entity.
func (v *ValidationTester) TestValidationSuccess(entity any, validator func(any) error) {
	err := validator(entity)
	Expect(err).ToNot(HaveOccurred(), "Expected validation to pass for %s", v.entityName)
}

// TestValidationFailure verifies that validation fails for invalid entity.
func (v *ValidationTester) TestValidationFailure(entity any, validator func(any) error, expectedFieldErrors ...string) {
	err := validator(entity)
	Expect(err).To(HaveOccurred(), "Expected validation to fail for %s", v.entityName)

	// Check if it's a validation error
	validationErr, isValidationError := errors.AsValidationError(err)
	Expect(isValidationError).To(BeTrue(), "Expected validation error, got: %v", err)

	// Check expected field errors
	for _, expectedField := range expectedFieldErrors {
		Expect(validationErr.Error()).To(ContainSubstring(expectedField),
			"Expected validation error to mention field '%s', got: %s", expectedField, validationErr.Error())
	}
}

// TestValidationMessage verifies specific validation error message.
func (v *ValidationTester) TestValidationMessage(entity any, validator func(any) error, expectedMessage string) {
	err := validator(entity)
	Expect(err).To(HaveOccurred(), "Expected validation to fail for %s", v.entityName)
	Expect(err.Error()).To(ContainSubstring(expectedMessage),
		"Expected error message to contain '%s', got: %s", expectedMessage, err.Error())
}

// FieldValidationTester provides field-specific validation testing.
type FieldValidationTester struct {
	fieldName string
	tester    *ValidationTester
}

// NewFieldValidationTester creates a field validation tester.
func NewFieldValidationTester(entityName, fieldName string) *FieldValidationTester {
	return &FieldValidationTester{
		fieldName: fieldName,
		tester:    NewValidationTester(entityName),
	}
}

// TestRequired tests that a field is required (empty values should fail).
func (f *FieldValidationTester) TestRequired(createEntityFunc func(fieldValue any) (any, error), validator func(any) error) {
	emptyValues := []any{"", nil}

	for _, emptyValue := range emptyValues {
		entity, err := createEntityFunc(emptyValue)
		if err != nil {
			// Creation itself failed, which is expected for required fields
			base.AssertValidationErrorForField(entity, err, f.fieldName)
		} else {
			// Creation succeeded, validation should fail
			f.tester.TestValidationFailure(entity, validator, f.fieldName)
		}
	}
}

// TestMinLength tests minimum length validation for string fields.
func (f *FieldValidationTester) TestMinLength(createEntityFunc func(string) (any, error), validator func(any) error, minLength int) {
	// Test values below minimum length
	invalidValues := make([]string, 0)
	for i := 0; i < minLength; i++ {
		invalidValue := string(make([]byte, i))
		for j := 0; j < i; j++ {
			invalidValue += "a"
		}
		invalidValues = append(invalidValues, invalidValue)
	}

	for _, invalidValue := range invalidValues {
		entity, err := createEntityFunc(invalidValue)
		if err != nil {
			base.AssertValidationErrorForField(entity, err, f.fieldName)
		} else {
			f.tester.TestValidationFailure(entity, validator, f.fieldName)
		}
	}

	// Test valid length (exactly minimum)
	validValue := ""
	for i := 0; i < minLength; i++ {
		validValue += "a"
	}

	entity, err := createEntityFunc(validValue)
	if err != nil {
		Expect(err).ToNot(HaveOccurred(), "Expected entity creation to succeed with valid length")
	} else {
		f.tester.TestValidationSuccess(entity, validator)
	}
}

// TestMaxLength tests maximum length validation for string fields.
func (f *FieldValidationTester) TestMaxLength(createEntityFunc func(string) (any, error), validator func(any) error, maxLength int) {
	// Test value exceeding maximum length
	invalidValue := ""
	for i := 0; i <= maxLength; i++ {
		invalidValue += "a"
	}

	entity, err := createEntityFunc(invalidValue)
	if err != nil {
		base.AssertValidationErrorForField(entity, err, f.fieldName)
	} else {
		f.tester.TestValidationFailure(entity, validator, f.fieldName)
	}

	// Test valid length (exactly maximum)
	validValue := ""
	for i := 0; i < maxLength; i++ {
		validValue += "a"
	}

	entity, err = createEntityFunc(validValue)
	if err != nil {
		Expect(err).ToNot(HaveOccurred(), "Expected entity creation to succeed with valid length")
	} else {
		f.tester.TestValidationSuccess(entity, validator)
	}
}

// TestPattern tests pattern/format validation for string fields.
func (f *FieldValidationTester) TestPattern(createEntityFunc func(string) (any, error), validator func(any) error, validValues, invalidValues []string) {
	// Test invalid values
	for _, invalidValue := range invalidValues {
		entity, err := createEntityFunc(invalidValue)
		if err != nil {
			base.AssertValidationErrorForField(entity, err, f.fieldName)
		} else {
			f.tester.TestValidationFailure(entity, validator, f.fieldName)
		}
	}

	// Test valid values
	for _, validValue := range validValues {
		entity, err := createEntityFunc(validValue)
		if err != nil {
			Expect(err).ToNot(HaveOccurred(), "Expected entity creation to succeed with valid value: %s", validValue)
		} else {
			f.tester.TestValidationSuccess(entity, validator)
		}
	}
}

// EmailValidationTester provides email-specific validation testing.
type EmailValidationTester struct {
	*FieldValidationTester
}

// NewEmailValidationTester creates an email validation tester.
func NewEmailValidationTester(entityName string) *EmailValidationTester {
	return &EmailValidationTester{
		FieldValidationTester: NewFieldValidationTester(entityName, "email"),
	}
}

// TestEmailValidation tests comprehensive email validation.
func (e *EmailValidationTester) TestEmailValidation(createEntityFunc func(string) (any, error), validator func(any) error) {
	validEmails := []string{
		"test@example.com",
		"user.name@example.com",
		"user+tag@example.co.uk",
		"user123@example123.com",
		"test.email.with+symbol@example.com",
	}

	invalidEmails := []string{
		"",                       // empty
		"invalid",                // no @ symbol
		"@example.com",           // no local part
		"test@",                  // no domain
		"test @example.com",      // space in local part
		"test@ example.com",      // space in domain
		"test@example",           // no TLD
		"test..test@example.com", // double dots
		".test@example.com",      // leading dot
		"test.@example.com",      // trailing dot
		"test@.example.com",      // domain starts with dot
		"test@example.com.",      // domain ends with dot
	}

	e.TestPattern(createEntityFunc, validator, validEmails, invalidEmails)
}

// UserIDValidationTester provides UserID-specific validation testing.
type UserIDValidationTester struct {
	*FieldValidationTester
}

// NewUserIDValidationTester creates a UserID validation tester.
func NewUserIDValidationTester(entityName string) *UserIDValidationTester {
	return &UserIDValidationTester{
		FieldValidationTester: NewFieldValidationTester(entityName, "id"),
	}
}

// TestUserIDValidation tests comprehensive UserID validation.
func (u *UserIDValidationTester) TestUserIDValidation(createEntityFunc func(string) (any, error), validator func(any) error) {
	validUserIDs := []string{
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

	invalidUserIDs := []string{
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

	u.TestPattern(createEntityFunc, validator, validUserIDs, invalidUserIDs)
}

// NameValidationTester provides name-specific validation testing.
type NameValidationTester struct {
	*FieldValidationTester
}

// NewNameValidationTester creates a name validation tester.
func NewNameValidationTester(entityName string) *NameValidationTester {
	return &NameValidationTester{
		FieldValidationTester: NewFieldValidationTester(entityName, "name"),
	}
}

// TestNameValidation tests comprehensive name validation.
func (n *NameValidationTester) TestNameValidation(createEntityFunc func(string) (any, error), validator func(any) error) {
	validNames := []string{
		"John Doe",
		"Jane Smith",
		"Mary Jane Watson",
		"Jean-Pierre",
		"O'Connor",
		"van der Berg",
		"Test User",
		"Single",
		"Very Long Name With Many Words",
	}

	invalidNames := []string{
		"",        // empty
		"A",       // too short
		"123",     // only numbers
		"   ",     // whitespace only
		"!@#$",    // special characters only
		"Test123", // contains numbers (depending on validation rules)
	}

	n.TestPattern(createEntityFunc, validator, validNames, invalidNames)

	// Test minimum length (assuming names must be at least 2 characters)
	n.TestMinLength(createEntityFunc, validator, 2)
}

// ValidationScenarioTester provides comprehensive scenario-based validation testing.
type ValidationScenarioTester struct {
	scenarios map[string]ValidationScenario
}

// ValidationScenario defines a complete validation test scenario.
type ValidationScenario struct {
	Name                string
	Description         string
	CreateEntity        func() (any, error)
	ShouldPass          bool
	ExpectedFieldErrors []string
	ExpectedMessage     string
}

// NewValidationScenarioTester creates a validation scenario tester.
func NewValidationScenarioTester() *ValidationScenarioTester {
	return &ValidationScenarioTester{
		scenarios: make(map[string]ValidationScenario),
	}
}

// AddScenario adds a validation scenario to test.
func (v *ValidationScenarioTester) AddScenario(scenario ValidationScenario) {
	v.scenarios[scenario.Name] = scenario
}

// RunScenario executes a specific validation scenario.
func (v *ValidationScenarioTester) RunScenario(scenarioName string, validator func(any) error) {
	scenario, exists := v.scenarios[scenarioName]
	Expect(exists).To(BeTrue(), "Validation scenario '%s' not found", scenarioName)

	entity, err := scenario.CreateEntity()

	if scenario.ShouldPass {
		v.assertSuccessScenario(&scenario, entity, err, validator)
	} else {
		v.assertFailureScenario(&scenario, entity, err, validator)
	}
}

// RunAllScenarios executes all validation scenarios.
func (v *ValidationScenarioTester) RunAllScenarios(validator func(any) error) {
	for scenarioName := range v.scenarios {
		v.RunScenario(scenarioName, validator)
	}
}

// GetScenarioNames returns all available scenario names.
func (v *ValidationScenarioTester) GetScenarioNames() []string {
	names := make([]string, 0, len(v.scenarios))
	for name := range v.scenarios {
		names = append(names, name)
	}
	return names
}

func (v *ValidationScenarioTester) assertSuccessScenario(scenario *ValidationScenario, entity any, err error, validator func(any) error) {
	// Entity creation and validation should both succeed
	Expect(err).ToNot(HaveOccurred(), "Scenario '%s': Expected entity creation to succeed", scenario.Name)

	validationErr := validator(entity)
	Expect(validationErr).ToNot(HaveOccurred(), "Scenario '%s': Expected validation to pass", scenario.Name)
}

func (v *ValidationScenarioTester) assertFailureScenario(scenario *ValidationScenario, entity any, err error, validator func(any) error) {
	// Either entity creation or validation should fail
	if err != nil {
		v.assertCreationFailure(scenario, entity, err)
	} else {
		v.assertValidationFailure(scenario, entity, validator)
	}

	v.checkExpectedMessage(scenario, entity, err, validator)
}

func (v *ValidationScenarioTester) assertCreationFailure(scenario *ValidationScenario, entity any, err error) {
	// Entity creation failed - check if it's a validation error
	base.AssertValidationError(entity, err)

	// Check expected field errors
	for _, expectedField := range scenario.ExpectedFieldErrors {
		Expect(err.Error()).To(ContainSubstring(expectedField))
	}
}

func (v *ValidationScenarioTester) assertValidationFailure(scenario *ValidationScenario, entity any, validator func(any) error) {
	// Entity creation succeeded, validation should fail
	validationErr := validator(entity)
	base.AssertValidationError(entity, validationErr)

	// Check expected field errors
	for _, expectedField := range scenario.ExpectedFieldErrors {
		Expect(validationErr.Error()).To(ContainSubstring(expectedField))
	}
}

func (v *ValidationScenarioTester) checkExpectedMessage(scenario *ValidationScenario, entity any, err error, validator func(any) error) {
	// Check expected message if provided
	if scenario.ExpectedMessage == "" {
		return
	}

	if err != nil {
		Expect(err.Error()).To(ContainSubstring(scenario.ExpectedMessage))
	} else {
		validationErr := validator(entity)
		Expect(validationErr.Error()).To(ContainSubstring(scenario.ExpectedMessage))
	}
}

// Convenience functions for common validation testing patterns

// ValidateRequiredField tests that a field is required using various empty values.
func ValidateRequiredField(fieldName string, createInvalidEntity func() (any, error), validator func(any) error) {
	tester := NewFieldValidationTester("entity", fieldName)

	// Test with function that returns the invalid entity
	createEntityFunc := func(_ any) (any, error) {
		return createInvalidEntity()
	}

	tester.TestRequired(createEntityFunc, validator)
}

// ValidateEmailField tests email field validation comprehensively.
func ValidateEmailField(createEntityWithEmail func(string) (any, error), validator func(any) error) {
	tester := NewEmailValidationTester("entity")
	tester.TestEmailValidation(createEntityWithEmail, validator)
}

// ValidateUserIDField tests UserID field validation comprehensively.
func ValidateUserIDField(createEntityWithUserID func(string) (any, error), validator func(any) error) {
	tester := NewUserIDValidationTester("entity")
	tester.TestUserIDValidation(createEntityWithUserID, validator)
}

// ValidateNameField tests name field validation comprehensively.
func ValidateNameField(createEntityWithName func(string) (any, error), validator func(any) error) {
	tester := NewNameValidationTester("entity")
	tester.TestNameValidation(createEntityWithName, validator)
}

// ValidateSuccessfulCreation validates that an entity can be created successfully.
func ValidateSuccessfulCreation(entity any, err error) {
	base.AssertSuccess(entity, err)
}

// ValidateCreationFailure validates that entity creation fails with validation error.
func ValidateCreationFailure(entity any, err error, expectedFieldErrors ...string) {
	tester := NewValidationTester("entity")
	validationFunc := func(any) error { return err }
	tester.TestValidationFailure(entity, validationFunc, expectedFieldErrors...)
}
