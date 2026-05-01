package ids_test

import (
	"encoding/json"
	"testing"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/ids"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIDs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IDs Suite")
}

var _ = Describe("UserID", func() {
	Describe("NewUserID", func() {
		It("creates valid UserID", func() {
			userID, err := ids.NewUserID("user-123")
			Expect(err).ToNot(HaveOccurred())
			Expect(userID.Get()).To(Equal("user-123"))
		})

		It("returns error for empty string", func() {
			_, err := ids.NewUserID("")
			Expect(err).To(HaveOccurred())
		})

		It("returns error for whitespace", func() {
			_, err := ids.NewUserID("  user-123  ")
			Expect(err).To(HaveOccurred())
		})

		It("returns error for invalid characters", func() {
			_, err := ids.NewUserID("user@123")
			Expect(err).To(HaveOccurred())
		})

		It("returns error for too short ID", func() {
			_, err := ids.NewUserID("a")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GenerateUserID", func() {
		It("generates unique IDs", func() {
			id1, err := ids.GenerateUserID()
			Expect(err).ToNot(HaveOccurred())

			id2, err := ids.GenerateUserID()
			Expect(err).ToNot(HaveOccurred())

			Expect(id1.Equal(id2)).To(BeFalse())
		})

		It("generates IDs with correct prefix", func() {
			userID, err := ids.GenerateUserID()
			Expect(err).ToNot(HaveOccurred())
			Expect(userID.Get()).To(HavePrefix("user_"))
		})
	})

	Describe("Equal", func() {
		DescribeTable(
			"should return expected result",
			func(id1Str, id2Str string, expected bool) {
				id1, _ := ids.NewUserID(id1Str)
				id2, _ := ids.NewUserID(id2Str)
				Expect(id1.Equal(id2)).To(Equal(expected))
			},
			Entry("same value returns true", "user-123", "user-123", true),
			Entry("different values return false", "user-123", "user-456", false),
		)
	})

	Describe("IsZero", func() {
		It("returns true for zero value", func() {
			var zero ids.UserID
			Expect(zero.IsZero()).To(BeTrue())
		})

		It("returns false for non-zero value", func() {
			userID, _ := ids.NewUserID("user-123")
			Expect(userID.IsZero()).To(BeFalse())
		})
	})

	Describe("JSON Serialization", func() {
		It("marshals to string", func() {
			userID, _ := ids.NewUserID("user-123")
			data, err := json.Marshal(userID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"user-123"`))
		})

		It("unmarshals from string", func() {
			var userID ids.UserID

			err := json.Unmarshal([]byte(`"user-123"`), &userID)
			Expect(err).ToNot(HaveOccurred())
			Expect(userID.Get()).To(Equal("user-123"))
		})

		It("marshals zero value to null", func() {
			var zero ids.UserID

			data, err := json.Marshal(zero)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("null"))
		})
	})
})

var _ = Describe("SessionID", func() {
	Describe("NewSessionID", func() {
		It("creates valid SessionID", func() {
			sessionID, err := ids.NewSessionID("sess-123")
			Expect(err).ToNot(HaveOccurred())
			Expect(sessionID.Get()).To(Equal("sess-123"))
		})
	})

	Describe("Type Safety", func() {
		It("prevents mixing UserID and SessionID", func() {
			// This should not compile if uncommented:
			// userID, _ := ids.NewUserID("user-123")
			// sessionID, _ := ids.NewSessionID("sess-456")
			// _ = userID.Equal(sessionID) // Compile error: type mismatch

			// Correct usage:
			userID1, _ := ids.NewUserID("user-123")
			userID2, _ := ids.NewUserID("user-123")
			Expect(userID1.Equal(userID2)).To(BeTrue())
		})
	})
})
