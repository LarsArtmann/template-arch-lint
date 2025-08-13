package pages

import (
	"bytes"
	"context"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

var _ = Describe("User Form Page Templates", func() {
	var (
		ctx    context.Context
		buffer *bytes.Buffer
	)

	BeforeEach(func() {
		ctx = context.Background()
		buffer = &bytes.Buffer{}
	})

	// Helper function to create test user
	createTestUser := func(id, email, name string) *entities.User {
		userID, err := values.NewUserID(id)
		Expect(err).To(BeNil())
		user, err := entities.NewUser(userID, email, name)
		Expect(err).To(BeNil())
		return user
	}

	Describe("CreateUserPage component", func() {
		Context("when rendering create user page", func() {
			It("should render complete page with layout and create form", func() {
				// When
				component := CreateUserPage()
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain full HTML document structure (from BaseLayout)
				Expect(html).To(ContainSubstring("<!DOCTYPE html>"))
				Expect(html).To(ContainSubstring("<title>Create User - User Management System</title>"))

				// Should contain navigation (from BaseLayout)
				Expect(html).To(ContainSubstring("User Management"))

				// Should contain create user specific content
				Expect(html).To(ContainSubstring("Create New User"))
				Expect(html).To(ContainSubstring("Add a new user to the system"))

				// Should contain create form
				Expect(html).To(ContainSubstring("hx-post=\"/users\""))
				Expect(html).To(ContainSubstring("Create User"))
			})

			It("should include all external dependencies from layout", func() {
				// When
				component := CreateUserPage()
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should include external dependencies
				Expect(html).To(ContainSubstring("tailwindcss.com"))
				Expect(html).To(ContainSubstring("htmx.org"))
				Expect(html).To(ContainSubstring("heroicons"))
			})
		})
	})

	Describe("CreateUserContent component", func() {
		Context("when rendering create form content", func() {
			It("should render proper page structure and styling", func() {
				// When
				component := CreateUserContent()
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain proper container styling
				Expect(html).To(ContainSubstring("max-w-md mx-auto"))
				Expect(html).To(ContainSubstring("bg-white rounded-lg shadow-sm border border-gray-200 p-6"))

				// Should contain header section
				Expect(html).To(ContainSubstring("mb-6"))
				Expect(html).To(ContainSubstring("text-2xl font-bold text-gray-900"))
			})

			It("should contain proper header information", func() {
				// When
				component := CreateUserContent()
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain create-specific header
				Expect(html).To(ContainSubstring("<h1 class=\"text-2xl font-bold text-gray-900\">Create New User</h1>"))
				Expect(html).To(ContainSubstring("Add a new user to the system"))
			})

			It("should include create form with ID field", func() {
				// When
				component := CreateUserContent()
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain create form attributes
				Expect(html).To(ContainSubstring("hx-post=\"/users\""))

				// Should contain ID field (specific to create mode)
				Expect(html).To(ContainSubstring("User ID"))
				Expect(html).To(ContainSubstring("name=\"id\""))

				// Should contain create button
				Expect(html).To(ContainSubstring("Create User"))

				// Should contain cancel link
				Expect(html).To(ContainSubstring("href=\"/users\""))
				Expect(html).To(ContainSubstring("Cancel"))
			})

			It("should include proper form validation", func() {
				// When
				component := CreateUserContent()
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain required attributes
				Expect(html).To(ContainSubstring("required"))

				// Should contain proper input types
				Expect(html).To(ContainSubstring("type=\"text\""))
				Expect(html).To(ContainSubstring("type=\"email\""))

				// Should contain placeholder text
				Expect(html).To(ContainSubstring("placeholder=\"unique-user-id\""))
				Expect(html).To(ContainSubstring("placeholder=\"John Doe\""))
				Expect(html).To(ContainSubstring("placeholder=\"john@example.com\""))
			})
		})

		Context("when checking accessibility", func() {
			It("should include proper form labels and structure", func() {
				// When
				component := CreateUserContent()
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain proper labels
				Expect(html).To(ContainSubstring("User ID"))
				Expect(html).To(ContainSubstring("Full Name"))
				Expect(html).To(ContainSubstring("Email Address"))

				// Should contain proper form structure
				Expect(html).To(ContainSubstring("space-y-4"))
			})
		})
	})

	Describe("EditUserPage component", func() {
		Context("when rendering edit user page", func() {
			It("should render complete page with layout and edit form", func() {
				// Given
				user := createTestUser("edit-page-user", "editpage@example.com", "Edit Page User")

				// When
				component := EditUserPage(user)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain full HTML document structure (from BaseLayout)
				Expect(html).To(ContainSubstring("<!DOCTYPE html>"))
				Expect(html).To(ContainSubstring("<title>Edit User - User Management System</title>"))

				// Should contain edit user specific content
				Expect(html).To(ContainSubstring("Edit User"))
				Expect(html).To(ContainSubstring("Update user information"))

				// Should contain edit form with user data
				Expect(html).To(ContainSubstring("hx-put=\"/users/edit-page-user\""))
				Expect(html).To(ContainSubstring("value=\"Edit Page User\""))
				Expect(html).To(ContainSubstring("value=\"editpage@example.com\""))

				// Should contain update button
				Expect(html).To(ContainSubstring("Update User"))
			})

			It("should handle user with special characters", func() {
				// Given
				user := createTestUser(
					"special-edit-user",
					"special+test@example.com",
					"Special & Edit <User>",
				)

				// When
				component := EditUserPage(user)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should properly escape special characters in form values
				Expect(html).To(ContainSubstring("value=\"Special &amp; Edit &lt;User&gt;\""))
				Expect(html).To(ContainSubstring("value=\"special+test@example.com\""))
			})
		})

		Context("with nil user", func() {
			It("should handle nil user gracefully", func() {
				// When
				component := EditUserPage(nil)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should still render the page structure
				Expect(html).To(ContainSubstring("Edit User"))
				Expect(html).To(ContainSubstring("Update user information"))
			})
		})
	})

	Describe("EditUserContent component", func() {
		Context("when rendering edit form content", func() {
			It("should render proper page structure for editing", func() {
				// Given
				user := createTestUser("edit-content-user", "editcontent@example.com", "Edit Content User")

				// When
				component := EditUserContent(user)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain proper container styling (same as create)
				Expect(html).To(ContainSubstring("max-w-md mx-auto"))
				Expect(html).To(ContainSubstring("bg-white rounded-lg shadow-sm border border-gray-200 p-6"))

				// Should contain edit-specific header
				Expect(html).To(ContainSubstring("<h1 class=\"text-2xl font-bold text-gray-900\">Edit User</h1>"))
				Expect(html).To(ContainSubstring("Update user information"))
			})

			It("should include edit form with populated values", func() {
				// Given
				user := createTestUser("populated-user", "populated@example.com", "Populated User")

				// When
				component := EditUserContent(user)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain edit form attributes
				Expect(html).To(ContainSubstring("hx-put=\"/users/populated-user\""))

				// Should NOT contain ID field (not editable)
				Expect(html).ToNot(ContainSubstring("User ID"))

				// Should contain populated form values
				Expect(html).To(ContainSubstring("value=\"Populated User\""))
				Expect(html).To(ContainSubstring("value=\"populated@example.com\""))

				// Should contain update button
				Expect(html).To(ContainSubstring("Update User"))
			})

			It("should maintain form structure consistency with create form", func() {
				// Given
				user := createTestUser("consistency-user", "consistency@example.com", "Consistency User")

				// When
				editComponent := EditUserContent(user)
				err := editComponent.Render(ctx, buffer)
				editHTML := buffer.String()

				buffer.Reset()

				createComponent := CreateUserContent()
				err = createComponent.Render(ctx, buffer)
				createHTML := buffer.String()

				// Then
				Expect(err).To(BeNil())

				// Both should have similar structure
				Expect(editHTML).To(ContainSubstring("max-w-md mx-auto"))
				Expect(createHTML).To(ContainSubstring("max-w-md mx-auto"))

				// Both should have proper form styling
				Expect(editHTML).To(ContainSubstring("space-y-4"))
				Expect(createHTML).To(ContainSubstring("space-y-4"))

				// Both should have cancel buttons
				Expect(editHTML).To(ContainSubstring("Cancel"))
				Expect(createHTML).To(ContainSubstring("Cancel"))
			})
		})

		Context("with edge case data", func() {
			It("should handle user with long name gracefully", func() {
				// Given
				longName := strings.Repeat("Very Long Name ", 10)
				user := createTestUser("long-name-user", "long@example.com", longName)

				// When
				component := EditUserContent(user)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain the long name in the form value
				Expect(html).To(ContainSubstring("value=\"" + longName + "\""))
			})

			It("should handle user with special email formats", func() {
				// Given
				user := createTestUser("special-email-user", "test+tag@sub.example.co.uk", "Special Email User")

				// When
				component := EditUserContent(user)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain the special email format
				Expect(html).To(ContainSubstring("value=\"test+tag@sub.example.co.uk\""))
			})
		})
	})

	Describe("Form styling and layout consistency", func() {
		Context("when comparing create and edit forms", func() {
			It("should have consistent styling between create and edit forms", func() {
				// Given
				user := createTestUser("style-user", "style@example.com", "Style User")

				// When - render both forms
				createComponent := CreateUserContent()
				buffer.Reset()
				err := createComponent.Render(ctx, buffer)
				createHTML := buffer.String()

				buffer.Reset()
				editComponent := EditUserContent(user)
				err = editComponent.Render(ctx, buffer)
				editHTML := buffer.String()

				// Then
				Expect(err).To(BeNil())

				// Should have same container styling
				containerClass := "max-w-md mx-auto"
				Expect(createHTML).To(ContainSubstring(containerClass))
				Expect(editHTML).To(ContainSubstring(containerClass))

				// Should have same card styling
				cardClass := "bg-white rounded-lg shadow-sm border border-gray-200 p-6"
				Expect(createHTML).To(ContainSubstring(cardClass))
				Expect(editHTML).To(ContainSubstring(cardClass))

				// Should have same header styling
				headerClass := "text-2xl font-bold text-gray-900"
				Expect(createHTML).To(ContainSubstring(headerClass))
				Expect(editHTML).To(ContainSubstring(headerClass))
			})

			It("should have responsive design classes", func() {
				// When
				component := CreateUserContent()
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain responsive classes
				Expect(html).To(ContainSubstring("mx-auto"))
				Expect(html).To(ContainSubstring("max-w-md"))

				// Form should be responsive-friendly
				Expect(html).To(ContainSubstring("block w-full"))
			})
		})
	})

	Describe("Performance and integration", func() {
		Context("when testing rendering performance", func() {
			It("should render create page efficiently", func() {
				// When
				start := time.Now()
				component := CreateUserPage()
				err := component.Render(ctx, buffer)
				duration := time.Since(start)

				// Then
				Expect(err).To(BeNil())
				Expect(duration).To(BeNumerically("<", 50*time.Millisecond))
			})

			It("should render edit page efficiently", func() {
				// Given
				user := createTestUser("perf-user", "perf@example.com", "Performance User")

				// When
				start := time.Now()
				component := EditUserPage(user)
				err := component.Render(ctx, buffer)
				duration := time.Since(start)

				// Then
				Expect(err).To(BeNil())
				Expect(duration).To(BeNumerically("<", 50*time.Millisecond))
			})
		})

		Context("when testing HTML structure", func() {
			It("should maintain proper HTML nesting", func() {
				// When
				component := CreateUserPage()
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should have proper HTML structure
				Expect(strings.Count(html, "<html")).To(Equal(1))
				Expect(strings.Count(html, "</html>")).To(Equal(1))
				Expect(strings.Count(html, "<form")).To(BeNumerically(">=", 1))
			})

			It("should include proper meta tags from layout", func() {
				// When
				component := CreateUserPage()
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain viewport meta tag
				Expect(html).To(ContainSubstring("name=\"viewport\""))
				Expect(html).To(ContainSubstring("charset=\"UTF-8\""))
			})
		})

		Context("when testing HTMX integration", func() {
			It("should include proper HTMX attributes in create form", func() {
				// When
				component := CreateUserContent()
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain HTMX form attributes
				Expect(html).To(ContainSubstring("hx-post=\"/users\""))
				Expect(html).To(ContainSubstring("hx-target=\"this\""))
				Expect(html).To(ContainSubstring("hx-swap=\"outerHTML\""))
			})

			It("should include proper HTMX attributes in edit form", func() {
				// Given
				user := createTestUser("htmx-user", "htmx@example.com", "HTMX User")

				// When
				component := EditUserContent(user)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain HTMX form attributes
				Expect(html).To(ContainSubstring("hx-put=\"/users/htmx-user\""))
				Expect(html).To(ContainSubstring("hx-target=\"this\""))
				Expect(html).To(ContainSubstring("hx-swap=\"outerHTML\""))
			})
		})
	})
})
