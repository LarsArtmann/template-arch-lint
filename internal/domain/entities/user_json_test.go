package entities

import (
	"encoding/json"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

// JSON Marshaling Test Suite - Verifies custom JSON implementation works correctly.
var _ = ginkgo.Describe("User JSON Marshaling", func() {
	var user *User

	ginkgo.BeforeEach(func() {
		var err error
		user, err = NewUserFromStrings("user-123", "test@example.com", "TestUser")
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		// Set specific timestamps for predictable testing
		user.Created = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		user.Modified = time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC)
	})

	ginkgo.Describe("MarshalJSON", func() {
		ginkgo.It("should marshal User to correct JSON structure", func() {
			// When
			jsonBytes, err := json.Marshal(user)

			// Then
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(jsonBytes).ToNot(gomega.BeEmpty())

			// Parse back to verify structure
			var jsonMap map[string]any
			err = json.Unmarshal(jsonBytes, &jsonMap)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			// Verify all expected fields are present with correct types
			gomega.Expect(jsonMap["id"]).To(gomega.Equal("user-123"))
			gomega.Expect(jsonMap["email"]).To(gomega.Equal("test@example.com"))
			gomega.Expect(jsonMap["name"]).To(gomega.Equal("TestUser"))
			gomega.Expect(jsonMap["created"]).To(gomega.Equal("2023-01-01T12:00:00Z"))
			gomega.Expect(jsonMap["modified"]).To(gomega.Equal("2023-01-02T12:00:00Z"))
		})

		ginkgo.It("should marshal to clean JSON without value object complexity", func() {
			// When
			jsonBytes, err := json.Marshal(user)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			jsonString := string(jsonBytes)

			// Then - should not contain internal value object structure
			gomega.Expect(jsonString).To(gomega.ContainSubstring("\"email\":\"test@example.com\""))
			gomega.Expect(jsonString).To(gomega.ContainSubstring("\"name\":\"TestUser\""))
			gomega.Expect(jsonString).ToNot(gomega.ContainSubstring("emailVO"))
			gomega.Expect(jsonString).ToNot(gomega.ContainSubstring("nameVO"))
			gomega.Expect(jsonString).ToNot(gomega.ContainSubstring("value"))
		})

		ginkgo.It("should handle special characters in email and name", func() {
			// Given
			specialUser, err := NewUserFromStrings("user-456", "test+special@sub.domain.com", "José María")
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			// When
			jsonBytes, err := json.Marshal(specialUser)

			// Then
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			var jsonMap map[string]any
			err = json.Unmarshal(jsonBytes, &jsonMap)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			gomega.Expect(jsonMap["email"]).To(gomega.Equal("test+special@sub.domain.com"))
			gomega.Expect(jsonMap["name"]).To(gomega.Equal("José María"))
		})
	})

	ginkgo.Describe("UnmarshalJSON", func() {
		ginkgo.It("should unmarshal valid JSON to User with value objects", func() {
			// Given
			jsonInput := `{
				"id": "user-789",
				"email": "json@example.com",
				"name": "JsonUser",
				"created": "2023-03-01T10:00:00Z",
				"modified": "2023-03-02T11:00:00Z"
			}`

			// When
			var unmarshaledUser User
			err := json.Unmarshal([]byte(jsonInput), &unmarshaledUser)

			// Then
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			// Verify value objects were created correctly
			gomega.Expect(unmarshaledUser.ID.String()).To(gomega.Equal("user-789"))
			gomega.Expect(unmarshaledUser.GetEmail().String()).To(gomega.Equal("json@example.com"))
			gomega.Expect(unmarshaledUser.GetUserName().String()).To(gomega.Equal("JsonUser"))

			// Verify timestamps
			expectedCreated := time.Date(2023, 3, 1, 10, 0, 0, 0, time.UTC)
			expectedModified := time.Date(2023, 3, 2, 11, 0, 0, 0, time.UTC)
			gomega.Expect(unmarshaledUser.Created).To(gomega.Equal(expectedCreated))
			gomega.Expect(unmarshaledUser.Modified).To(gomega.Equal(expectedModified))
		})

		ginkgo.It("should return validation error for invalid email in JSON", func() {
			// Given - JSON with invalid email
			jsonInput := `{
				"id": "user-invalid",
				"email": "not-an-email",
				"name": "InvalidUser",
				"created": "2023-03-01T10:00:00Z",
				"modified": "2023-03-02T11:00:00Z"
			}`

			// When
			var unmarshaledUser User
			err := json.Unmarshal([]byte(jsonInput), &unmarshaledUser)

			// Then
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("email"))
		})

		ginkgo.It("should return validation error for invalid user ID in JSON", func() {
			// Given - JSON with empty user ID
			jsonInput := `{
				"id": "",
				"email": "valid@example.com",
				"name": "ValidUser",
				"created": "2023-03-01T10:00:00Z",
				"modified": "2023-03-02T11:00:00Z"
			}`

			// When
			var unmarshaledUser User
			err := json.Unmarshal([]byte(jsonInput), &unmarshaledUser)

			// Then
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("user ID"))
		})

		ginkgo.It("should return validation error for invalid name in JSON", func() {
			// Given - JSON with empty name
			jsonInput := `{
				"id": "user-valid",
				"email": "valid@example.com",
				"name": "",
				"created": "2023-03-01T10:00:00Z",
				"modified": "2023-03-02T11:00:00Z"
			}`

			// When
			var unmarshaledUser User
			err := json.Unmarshal([]byte(jsonInput), &unmarshaledUser)

			// Then
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("name"))
		})

		ginkgo.It("should handle malformed JSON gracefully", func() {
			// Given - Malformed JSON
			jsonInput := `{
				"id": "user-malformed",
				"email": "test@example.com"
				// Missing comma, invalid JSON
			}`

			// When
			var unmarshaledUser User
			err := json.Unmarshal([]byte(jsonInput), &unmarshaledUser)

			// Then
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})

	ginkgo.Describe("Round-trip JSON serialization", func() {
		ginkgo.It("should maintain data integrity through marshal/unmarshal cycle", func() {
			// Given - Original user
			original := user

			// When - Marshal then unmarshal
			jsonBytes, err := json.Marshal(original)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			var roundtrip User
			err = json.Unmarshal(jsonBytes, &roundtrip)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			// Then - Data should be identical
			gomega.Expect(roundtrip.ID.String()).To(gomega.Equal(original.ID.String()))
			gomega.Expect(roundtrip.GetEmail().String()).To(gomega.Equal(original.GetEmail().String()))
			gomega.Expect(roundtrip.GetUserName().String()).To(gomega.Equal(original.GetUserName().String()))
			gomega.Expect(roundtrip.Created).To(gomega.Equal(original.Created))
			gomega.Expect(roundtrip.Modified).To(gomega.Equal(original.Modified))

			// Verify value object functionality still works
			gomega.Expect(roundtrip.EmailDomain()).To(gomega.Equal("example.com"))
			gomega.Expect(roundtrip.IsEmailValid()).To(gomega.BeTrue())
			gomega.Expect(roundtrip.IsNameReserved()).To(gomega.BeFalse())
		})

		ginkgo.It("should maintain validation capabilities after unmarshaling", func() {
			// Given - User from JSON
			jsonInput := `{
				"id": "roundtrip-user",
				"email": "roundtrip@example.com",
				"name": "RoundtripUser",
				"created": "2023-03-01T10:00:00Z",
				"modified": "2023-03-02T11:00:00Z"
			}`

			var unmarshaledUser User
			err := json.Unmarshal([]byte(jsonInput), &unmarshaledUser)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			// When - Use domain methods
			domain := unmarshaledUser.EmailDomain()
			isValid := unmarshaledUser.IsEmailValid()

			// Then - All functionality should work
			gomega.Expect(domain).To(gomega.Equal("example.com"))
			gomega.Expect(isValid).To(gomega.BeTrue())

			// Validation should also work
			err = unmarshaledUser.Validate()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		})
	})
})
