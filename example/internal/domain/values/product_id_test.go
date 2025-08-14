package values_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint-example/internal/domain/values"
)

func TestProductID(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ProductID Value Object Suite")
}

var _ = Describe("ProductID", func() {
	Describe("NewProductID", func() {
		Context("with valid input", func() {
			It("should create a valid ProductID", func() {
				id, err := values.NewProductID("prod-123")

				Expect(err).NotTo(HaveOccurred())
				Expect(id.String()).To(Equal("prod-123"))
				Expect(id.IsEmpty()).To(BeFalse())
			})
		})

		Context("with invalid input", func() {
			It("should reject empty ID", func() {
				_, err := values.NewProductID("")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cannot be empty"))
			})

			It("should reject ID with spaces", func() {
				_, err := values.NewProductID("prod 123")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cannot contain spaces"))
			})

			It("should reject very long ID", func() {
				longID := string(make([]byte, 51)) // 51 characters
				for i := range longID {
					longID = longID[:i] + "a" + longID[i+1:]
				}

				_, err := values.NewProductID(longID)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cannot exceed 50 characters"))
			})
		})
	})

	Describe("Equals", func() {
		It("should return true for equal IDs", func() {
			id1, _ := values.NewProductID("prod-123")
			id2, _ := values.NewProductID("prod-123")

			Expect(id1.Equals(id2)).To(BeTrue())
		})

		It("should return false for different IDs", func() {
			id1, _ := values.NewProductID("prod-123")
			id2, _ := values.NewProductID("prod-456")

			Expect(id1.Equals(id2)).To(BeFalse())
		})
	})

	Describe("IsEmpty", func() {
		It("should return false for valid ID", func() {
			id, _ := values.NewProductID("prod-123")

			Expect(id.IsEmpty()).To(BeFalse())
		})

		It("should return true for zero value", func() {
			var id values.ProductID

			Expect(id.IsEmpty()).To(BeTrue())
		})
	})
})
