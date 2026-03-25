package services_test

import (
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/onsi/gomega"
)

// CreateTestUserID creates a new UserID from a string, failing on error.
func CreateTestUserID(id string) values.UserID {
	userID, err := values.NewUserID(id)
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())
	return userID
}
