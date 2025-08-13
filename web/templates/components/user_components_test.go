package components

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

var _ = Describe("User Components Templates", func() {
	var (
		ctx    context.Context
		buffer *bytes.Buffer
	)

	BeforeEach(func() {
		ctx = context.Background()
		buffer = &bytes.Buffer{}
	})

	// Helper function to create test users
	createTestUser := func(id, email, name string) *entities.User {
		userID, err := values.NewUserID(id)
		Expect(err).To(BeNil())
		user, err := entities.NewUser(userID, email, name)
		Expect(err).To(BeNil())
		return user
	}

	Describe("StatsGrid component", func() {
		Context("with valid statistics", func() {
			It("should render all stat cards with correct data", func() {
				// Given
				stats := map[string]int{
					"total":                       42,
					"active":                      38,
					"domains":                     5,
					"avg_days_since_registration": 30,
				}

				// When
				component := StatsGrid(stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain grid structure
				Expect(html).To(ContainSubstring("grid grid-cols-1 md:grid-cols-4 gap-6"))

				// Should contain all stat values
				Expect(html).To(ContainSubstring("42")) // total users
				Expect(html).To(ContainSubstring("38")) // active users
				Expect(html).To(ContainSubstring("5"))  // domains
				Expect(html).To(ContainSubstring("30")) // avg days

				// Should contain stat labels
				Expect(html).To(ContainSubstring("Total Users"))
				Expect(html).To(ContainSubstring("Active Users"))
				Expect(html).To(ContainSubstring("Domains"))
				Expect(html).To(ContainSubstring("Avg Days"))

				// Should contain proper styling classes
				Expect(html).To(ContainSubstring("bg-white rounded-lg shadow-sm"))
				Expect(html).To(ContainSubstring("text-blue-600"))
				Expect(html).To(ContainSubstring("text-green-600"))
				Expect(html).To(ContainSubstring("text-purple-600"))
				Expect(html).To(ContainSubstring("text-yellow-600"))
			})

			It("should handle zero values gracefully", func() {
				// Given
				stats := map[string]int{
					"total":                       0,
					"active":                      0,
					"domains":                     0,
					"avg_days_since_registration": 0,
				}

				// When
				component := StatsGrid(stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain zero values
				Expect(strings.Count(html, ">0<")).To(Equal(4))
			})

			It("should include proper SVG icons", func() {
				// Given
				stats := map[string]int{
					"total":                       1,
					"active":                      1,
					"domains":                     1,
					"avg_days_since_registration": 1,
				}

				// When
				component := StatsGrid(stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain SVG icons with proper attributes
				Expect(strings.Count(html, "<svg")).To(Equal(4))
				Expect(html).To(ContainSubstring("stroke=\"currentColor\""))
				Expect(html).To(ContainSubstring("viewBox=\"0 0 24 24\""))
			})
		})

		Context("with missing statistics", func() {
			It("should handle missing keys gracefully", func() {
				// Given
				stats := map[string]int{
					"total": 10,
					// missing other keys
				}

				// When
				component := StatsGrid(stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain 10 for total and 0 for others
				Expect(html).To(ContainSubstring("10"))
				Expect(strings.Count(html, ">0<")).To(Equal(3))
			})
		})
	})

	Describe("UsersList component", func() {
		Context("with empty user list", func() {
			It("should render empty state with proper message", func() {
				// Given
				var users []*entities.User

				// When
				component := UsersList(users)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain empty state message
				Expect(html).To(ContainSubstring("No users found"))
				Expect(html).To(ContainSubstring("Get started by creating a new user"))
				Expect(html).To(ContainSubstring("href=\"/users/new\""))
				Expect(html).To(ContainSubstring("Add User"))

				// Should contain empty state styling
				Expect(html).To(ContainSubstring("p-12 text-center"))
				Expect(html).To(ContainSubstring("text-gray-400"))
			})
		})

		Context("with user list", func() {
			It("should render users table with header and count", func() {
				// Given
				users := []*entities.User{
					createTestUser("user-1", "user1@example.com", "User One"),
					createTestUser("user-2", "user2@example.com", "User Two"),
					createTestUser("user-3", "user3@example.com", "User Three"),
				}

				// When
				component := UsersList(users)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain table structure
				Expect(html).To(ContainSubstring("min-w-full divide-y divide-gray-200"))

				// Should contain user count in header
				Expect(html).To(ContainSubstring("Users (3)"))
				Expect(html).To(ContainSubstring("Click on a user to edit"))

				// Should contain user rows
				Expect(html).To(ContainSubstring("User One"))
				Expect(html).To(ContainSubstring("User Two"))
				Expect(html).To(ContainSubstring("User Three"))
				Expect(html).To(ContainSubstring("user1@example.com"))
				Expect(html).To(ContainSubstring("user2@example.com"))
				Expect(html).To(ContainSubstring("user3@example.com"))
			})

			It("should handle single user correctly", func() {
				// Given
				users := []*entities.User{
					createTestUser("single-user", "single@example.com", "Single User"),
				}

				// When
				component := UsersList(users)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain correct count
				Expect(html).To(ContainSubstring("Users (1)"))
				Expect(html).To(ContainSubstring("Single User"))
			})
		})
	})

	Describe("UserRow component", func() {
		Context("with valid user data", func() {
			It("should render complete user row with all information", func() {
				// Given
				user := createTestUser("test-user-123", "test@example.com", "Test User")

				// When
				component := UserRow(user)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain user ID as element ID
				Expect(html).To(ContainSubstring("id=\"user-test-user-123\""))

				// Should contain user information
				Expect(html).To(ContainSubstring("Test User"))
				Expect(html).To(ContainSubstring("test@example.com"))
				Expect(html).To(ContainSubstring("ID: test-user-123"))

				// Should contain avatar with first letter
				Expect(html).To(ContainSubstring(">T<"))

				// Should contain status badge
				Expect(html).To(ContainSubstring("Active"))
				Expect(html).To(ContainSubstring("bg-green-100 text-green-800"))

				// Should contain formatted creation date
				today := time.Now().Format("Jan 2, 2006")
				Expect(html).To(ContainSubstring(today))
			})

			It("should include all action buttons with proper HTMX attributes", func() {
				// Given
				user := createTestUser("action-user", "action@example.com", "Action User")

				// When
				component := UserRow(user)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Edit button
				Expect(html).To(ContainSubstring("Edit"))
				Expect(html).To(ContainSubstring("hx-get=\"/users/action-user/edit-inline\""))
				Expect(html).To(ContainSubstring("hx-target=\"#user-action-user\""))

				// Delete button
				Expect(html).To(ContainSubstring("Delete"))
				Expect(html).To(ContainSubstring("hx-delete=\"/users/action-user\""))
				Expect(html).To(ContainSubstring("hx-confirm=\"Are you sure you want to delete this user?\""))

				// View button
				Expect(html).To(ContainSubstring("View"))
				Expect(html).To(ContainSubstring("href=\"/users/action-user\""))
			})

			It("should handle special characters in user data", func() {
				// Given
				user := createTestUser("special-user", "test+special@example.com", "Test & Special <User>")

				// When
				component := UserRow(user)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should properly escape special characters
				Expect(html).To(ContainSubstring("Test &amp; Special &lt;User&gt;"))
				Expect(html).To(ContainSubstring("test+special@example.com"))

				// Avatar should use first character
				Expect(html).To(ContainSubstring(">T<"))
			})
		})

		Context("with edge cases", func() {
			It("should handle empty name gracefully", func() {
				// This would normally fail validation, but testing the template
				user := &entities.User{
					Name:  "",
					Email: "empty@example.com",
				}
				userID, _ := values.NewUserID("empty-name")
				user.ID = userID
				user.Created = time.Now()

				// When
				component := UserRow(user)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				// Template should still render even with empty data
			})
		})
	})

	Describe("UserEditRow component", func() {
		Context("with valid user data", func() {
			It("should render inline edit form with proper structure", func() {
				// Given
				user := createTestUser("edit-user", "edit@example.com", "Edit User")

				// When
				component := UserEditRow(user)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain form with HTMX attributes
				Expect(html).To(ContainSubstring("hx-put=\"/users/edit-user\""))
				Expect(html).To(ContainSubstring("hx-target=\"#user-edit-user\""))
				Expect(html).To(ContainSubstring("hx-swap=\"outerHTML\""))

				// Should contain input fields with current values
				Expect(html).To(ContainSubstring("value=\"Edit User\""))
				Expect(html).To(ContainSubstring("value=\"edit@example.com\""))

				// Should contain proper form labels and IDs
				Expect(html).To(ContainSubstring("id=\"name-edit-user\""))
				Expect(html).To(ContainSubstring("id=\"email-edit-user\""))

				// Should contain action buttons
				Expect(html).To(ContainSubstring(">Save</button>"))
				Expect(html).To(ContainSubstring(">Cancel</button>"))

				// Cancel button should have HTMX attributes
				Expect(html).To(ContainSubstring("hx-get=\"/users/edit-user/cancel-edit\""))
			})

			It("should have distinct styling for edit mode", func() {
				// Given
				user := createTestUser("style-user", "style@example.com", "Style User")

				// When
				component := UserEditRow(user)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should have edit mode styling
				Expect(html).To(ContainSubstring("bg-blue-50 border-l-4 border-blue-400"))
				Expect(html).To(ContainSubstring("focus:ring-blue-500"))
			})

			It("should include proper input validation", func() {
				// Given
				user := createTestUser("valid-user", "valid@example.com", "Valid User")

				// When
				component := UserEditRow(user)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain required attributes
				Expect(html).To(ContainSubstring("required"))
				Expect(html).To(ContainSubstring("type=\"email\""))
				Expect(html).To(ContainSubstring("placeholder=\"user@example.com\""))
			})
		})
	})

	Describe("UserForm component", func() {
		Context("in create mode", func() {
			It("should render create form with ID field", func() {
				// Given
				mode := "create"

				// When
				component := UserForm(nil, mode)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain create form attributes
				Expect(html).To(ContainSubstring("hx-post=\"/users\""))

				// Should contain ID field for create mode
				Expect(html).To(ContainSubstring("User ID"))
				Expect(html).To(ContainSubstring("name=\"id\""))
				Expect(html).To(ContainSubstring("unique-user-id"))

				// Should contain create button
				Expect(html).To(ContainSubstring("Create User"))
			})
		})

		Context("in edit mode", func() {
			It("should render edit form with populated fields", func() {
				// Given
				user := createTestUser("form-user", "form@example.com", "Form User")
				mode := "edit"

				// When
				component := UserForm(user, mode)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain edit form attributes
				Expect(html).To(ContainSubstring("hx-put=\"/users/form-user\""))

				// Should NOT contain ID field in edit mode
				Expect(html).ToNot(ContainSubstring("User ID"))

				// Should contain populated values
				Expect(html).To(ContainSubstring("value=\"Form User\""))
				Expect(html).To(ContainSubstring("value=\"form@example.com\""))

				// Should contain update button
				Expect(html).To(ContainSubstring("Update User"))
			})
		})

		Context("with form structure", func() {
			It("should include proper form validation and styling", func() {
				// Given
				mode := "create"

				// When
				component := UserForm(nil, mode)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain proper form structure
				Expect(html).To(ContainSubstring("space-y-4"))

				// Should contain proper labels
				Expect(html).To(ContainSubstring("Full Name"))
				Expect(html).To(ContainSubstring("Email Address"))

				// Should contain proper input types
				Expect(html).To(ContainSubstring("type=\"text\""))
				Expect(html).To(ContainSubstring("type=\"email\""))

				// Should contain validation
				Expect(html).To(ContainSubstring("required"))

				// Should contain cancel link
				Expect(html).To(ContainSubstring("href=\"/users\""))
				Expect(html).To(ContainSubstring("Cancel"))
			})
		})
	})

	Describe("UserFormSuccess component", func() {
		Context("with successful action", func() {
			It("should render success message with user details", func() {
				// Given
				user := createTestUser("success-user", "success@example.com", "Success User")
				action := "created"

				// When
				component := UserFormSuccess(user, action)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain success styling
				Expect(html).To(ContainSubstring("bg-green-50"))
				Expect(html).To(ContainSubstring("text-green-800"))

				// Should contain success message
				Expect(html).To(ContainSubstring("User created successfully!"))
				Expect(html).To(ContainSubstring("Success User (success@example.com) has been created"))

				// Should contain back link
				Expect(html).To(ContainSubstring("← Back to Users List"))
				Expect(html).To(ContainSubstring("href=\"/users\""))

				// Should contain success icon
				Expect(html).To(ContainSubstring("<svg"))
				Expect(html).To(ContainSubstring("text-green-400"))
			})

			It("should handle different action types", func() {
				// Given
				user := createTestUser("updated-user", "updated@example.com", "Updated User")
				action := "updated"

				// When
				component := UserFormSuccess(user, action)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain update-specific messaging
				Expect(html).To(ContainSubstring("User updated successfully!"))
				Expect(html).To(ContainSubstring("has been updated"))
			})
		})
	})

	Describe("ErrorMessage component", func() {
		Context("with error details", func() {
			It("should render error message with details", func() {
				// Given
				message := "Something went wrong"
				details := "Invalid email format provided"

				// When
				component := ErrorMessage(message, details)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain error styling
				Expect(html).To(ContainSubstring("bg-red-50"))
				Expect(html).To(ContainSubstring("text-red-800"))
				Expect(html).To(ContainSubstring("text-red-700"))

				// Should contain error message and details
				Expect(html).To(ContainSubstring("Something went wrong"))
				Expect(html).To(ContainSubstring("Invalid email format provided"))

				// Should contain error icon
				Expect(html).To(ContainSubstring("<svg"))
				Expect(html).To(ContainSubstring("text-red-400"))
			})

			It("should handle empty details gracefully", func() {
				// Given
				message := "Error without details"
				details := ""

				// When
				component := ErrorMessage(message, details)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain main message only
				Expect(html).To(ContainSubstring("Error without details"))

				// Should not render details section when empty
				Expect(strings.Count(html, "text-xs text-red-600")).To(Equal(0))
			})
		})
	})

	Describe("Template performance and structure", func() {
		Context("when rendering multiple components", func() {
			It("should render efficiently with large user lists", func() {
				// Given
				users := make([]*entities.User, 100)
				for i := 0; i < 100; i++ {
					users[i] = createTestUser(
						"user-"+string(rune('A'+i%26)),
						"user"+string(rune('0'+(i%10)))+"@example.com",
						"User "+string(rune('A'+i%26)),
					)
				}

				// When
				start := time.Now()
				component := UsersList(users)
				err := component.Render(ctx, buffer)
				duration := time.Since(start)

				// Then
				Expect(err).To(BeNil())
				Expect(duration).To(BeNumerically("<", 50*time.Millisecond))

				html := buffer.String()
				Expect(html).To(ContainSubstring("Users (100)"))
			})

			It("should properly escape all user data", func() {
				// Given
				user := createTestUser(
					"xss-test",
					"test@example.com",
					"<script>alert('xss')</script>",
				)

				// When
				component := UserRow(user)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should properly escape script tags
				Expect(html).ToNot(ContainSubstring("<script>"))
				Expect(html).To(ContainSubstring("&lt;script&gt;"))
			})
		})
	})
})
