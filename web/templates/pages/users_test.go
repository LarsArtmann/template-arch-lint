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

var _ = Describe("Users Page Templates", func() {
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

	// Helper function to create test stats
	createTestStats := func() map[string]int {
		return map[string]int{
			"total":                       25,
			"active":                      20,
			"domains":                     8,
			"avg_days_since_registration": 45,
		}
	}

	Describe("UsersPage component", func() {
		Context("with complete data", func() {
			It("should render complete users page with layout", func() {
				// Given
				users := []*entities.User{
					createTestUser("user-1", "user1@example.com", "User One"),
					createTestUser("user-2", "user2@example.com", "User Two"),
				}
				stats := createTestStats()

				// When
				component := UsersPage(users, stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain full HTML document structure (from BaseLayout)
				Expect(html).To(ContainSubstring("<!DOCTYPE html>"))
				Expect(html).To(ContainSubstring("<title>Users - User Management System</title>"))

				// Should contain navigation (from BaseLayout)
				Expect(html).To(ContainSubstring("User Management"))

				// Should contain users content
				Expect(html).To(ContainSubstring("User One"))
				Expect(html).To(ContainSubstring("User Two"))

				// Should contain stats
				Expect(html).To(ContainSubstring("25")) // total users
			})

			It("should include all external dependencies from layout", func() {
				// Given
				users := []*entities.User{createTestUser("dep-user", "dep@example.com", "Dep User")}
				stats := createTestStats()

				// When
				component := UsersPage(users, stats)
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

		Context("with empty data", func() {
			It("should render empty state properly", func() {
				// Given
				var users []*entities.User
				stats := map[string]int{
					"total":                       0,
					"active":                      0,
					"domains":                     0,
					"avg_days_since_registration": 0,
				}

				// When
				component := UsersPage(users, stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain empty state message
				Expect(html).To(ContainSubstring("No users found"))
				Expect(html).To(ContainSubstring("Get started by creating a new user"))

				// Should show zero stats
				Expect(strings.Count(html, ">0<")).To(BeNumerically(">=", 4))
			})
		})
	})

	Describe("UsersContent component", func() {
		Context("with complete content structure", func() {
			It("should render all content sections in correct order", func() {
				// Given
				users := []*entities.User{
					createTestUser("content-user-1", "content1@example.com", "Content User One"),
					createTestUser("content-user-2", "content2@example.com", "Content User Two"),
				}
				stats := createTestStats()

				// When
				component := UsersContent(users, stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Find positions of major sections
				headerPos := strings.Index(html, "text-3xl font-bold")
				statsPos := strings.Index(html, "grid-cols-1 md:grid-cols-4")
				searchPos := strings.Index(html, "Search users by name")
				listPos := strings.Index(html, "id=\"user-list\"")

				// Verify correct order
				Expect(headerPos).To(BeNumerically("<", statsPos))
				Expect(statsPos).To(BeNumerically("<", searchPos))
				Expect(searchPos).To(BeNumerically("<", listPos))
			})

			It("should contain proper page header with title and action button", func() {
				// Given
				users := []*entities.User{createTestUser("header-user", "header@example.com", "Header User")}
				stats := createTestStats()

				// When
				component := UsersContent(users, stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain page title
				Expect(html).To(ContainSubstring("<h1 class=\"text-3xl font-bold text-gray-900\">Users</h1>"))
				Expect(html).To(ContainSubstring("Manage your user accounts"))

				// Should contain add user button
				Expect(html).To(ContainSubstring("href=\"/users/new\""))
				Expect(html).To(ContainSubstring("Add New User"))
				Expect(html).To(ContainSubstring("bg-blue-600 hover:bg-blue-700"))
			})

			It("should render statistics grid section", func() {
				// Given
				users := []*entities.User{createTestUser("stats-user", "stats@example.com", "Stats User")}
				stats := map[string]int{
					"total":                       100,
					"active":                      85,
					"domains":                     12,
					"avg_days_since_registration": 60,
				}

				// When
				component := UsersContent(users, stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain all stat values
				Expect(html).To(ContainSubstring("100"))
				Expect(html).To(ContainSubstring("85"))
				Expect(html).To(ContainSubstring("12"))
				Expect(html).To(ContainSubstring("60"))

				// Should contain stat labels
				Expect(html).To(ContainSubstring("Total Users"))
				Expect(html).To(ContainSubstring("Active Users"))
			})

			It("should include complete search and filter section", func() {
				// Given
				users := []*entities.User{createTestUser("filter-user", "filter@example.com", "Filter User")}
				stats := createTestStats()

				// When
				component := UsersContent(users, stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Search input
				Expect(html).To(ContainSubstring("Search users by name or email"))
				Expect(html).To(ContainSubstring("hx-get=\"/users/search\""))
				Expect(html).To(ContainSubstring("hx-trigger=\"keyup changed delay:300ms\""))
				Expect(html).To(ContainSubstring("hx-target=\"#user-list\""))

				// Domain filter
				Expect(html).To(ContainSubstring("All Domains"))
				Expect(html).To(ContainSubstring("gmail.com"))
				Expect(html).To(ContainSubstring("example.com"))
				Expect(html).To(ContainSubstring("company.com"))

				// Active filter
				Expect(html).To(ContainSubstring("All Users"))
				Expect(html).To(ContainSubstring("Active Only"))
				Expect(html).To(ContainSubstring("Inactive Only"))

				// HTMX attributes for filters
				Expect(html).To(ContainSubstring("hx-include=\"[name='domain']\""))
				Expect(html).To(ContainSubstring("hx-include=\"[name='search']\""))
			})

			It("should include users list section with proper ID", func() {
				// Given
				users := []*entities.User{
					createTestUser("list-user-1", "list1@example.com", "List User One"),
					createTestUser("list-user-2", "list2@example.com", "List User Two"),
				}
				stats := createTestStats()

				// When
				component := UsersContent(users, stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain user list container
				Expect(html).To(ContainSubstring("id=\"user-list\""))

				// Should contain users
				Expect(html).To(ContainSubstring("List User One"))
				Expect(html).To(ContainSubstring("List User Two"))
				Expect(html).To(ContainSubstring("list1@example.com"))
			})
		})

		Context("with responsive design elements", func() {
			It("should include proper responsive classes", func() {
				// Given
				users := []*entities.User{createTestUser("resp-user", "resp@example.com", "Responsive User")}
				stats := createTestStats()

				// When
				component := UsersContent(users, stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain responsive classes
				Expect(html).To(ContainSubstring("md:grid-cols-4"))
				Expect(html).To(ContainSubstring("md:flex-row"))
				Expect(html).To(ContainSubstring("md:w-48"))
				Expect(html).To(ContainSubstring("md:w-32"))

				// Should contain flexible layout classes
				Expect(html).To(ContainSubstring("flex-col"))
				Expect(html).To(ContainSubstring("flex-1"))
			})

			It("should include proper spacing and layout classes", func() {
				// Given
				users := []*entities.User{createTestUser("layout-user", "layout@example.com", "Layout User")}
				stats := createTestStats()

				// When
				component := UsersContent(users, stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain spacing classes
				Expect(html).To(ContainSubstring("space-y-6"))
				Expect(html).To(ContainSubstring("gap-4"))
				Expect(html).To(ContainSubstring("gap-6"))

				// Should contain padding and margin classes
				Expect(html).To(ContainSubstring("p-6"))
				Expect(html).To(ContainSubstring("px-4"))
				Expect(html).To(ContainSubstring("py-2"))
			})
		})

		Context("with HTMX integration", func() {
			It("should include all necessary HTMX attributes for search functionality", func() {
				// Given
				users := []*entities.User{createTestUser("htmx-user", "htmx@example.com", "HTMX User")}
				stats := createTestStats()

				// When
				component := UsersContent(users, stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Search input HTMX attributes
				searchHTMX := []string{
					"hx-get=\"/users/search\"",
					"hx-trigger=\"keyup changed delay:300ms\"",
					"hx-target=\"#user-list\"",
					"hx-include=\"[name='domain']\"",
				}

				for _, attr := range searchHTMX {
					Expect(html).To(ContainSubstring(attr))
				}

				// Filter HTMX attributes
				filterHTMX := []string{
					"hx-get=\"/users/search\"",
					"hx-target=\"#user-list\"",
					"hx-include=\"[name='search']\"",
					"hx-include=\"[name='search'],[name='domain']\"",
				}

				for _, attr := range filterHTMX {
					Expect(html).To(ContainSubstring(attr))
				}
			})
		})
	})

	Describe("SearchUsersContent component", func() {
		Context("when rendering search results", func() {
			It("should render filtered users without full page layout", func() {
				// Given
				users := []*entities.User{
					createTestUser("search-user-1", "search1@example.com", "Search User One"),
					createTestUser("search-user-2", "search2@gmail.com", "Search User Two"),
				}

				// When
				component := SearchUsersContent(users)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should NOT contain full page structure
				Expect(html).ToNot(ContainSubstring("<!DOCTYPE html>"))
				Expect(html).ToNot(ContainSubstring("<title>"))

				// Should contain only the users list
				Expect(html).To(ContainSubstring("Search User One"))
				Expect(html).To(ContainSubstring("Search User Two"))
				Expect(html).To(ContainSubstring("search1@example.com"))
				Expect(html).To(ContainSubstring("search2@gmail.com"))

				// Should contain users count
				Expect(html).To(ContainSubstring("Users (2)"))
			})

			It("should handle empty search results", func() {
				// Given
				var users []*entities.User

				// When
				component := SearchUsersContent(users)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain empty state
				Expect(html).To(ContainSubstring("No users found"))
				Expect(html).To(ContainSubstring("Add User"))
			})

			It("should be optimized for HTMX partial updates", func() {
				// Given
				users := []*entities.User{
					createTestUser("partial-user", "partial@example.com", "Partial User"),
				}

				// When
				component := SearchUsersContent(users)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain only the necessary content for replacement
				Expect(html).To(ContainSubstring("bg-white rounded-lg shadow-sm"))
				Expect(html).To(ContainSubstring("Users (1)"))

				// Should not contain extra wrapper elements
				Expect(strings.Count(html, "<div")).To(BeNumerically("<", 20))
			})
		})
	})

	Describe("Integration and performance tests", func() {
		Context("when rendering complete page", func() {
			It("should render efficiently with large datasets", func() {
				// Given
				users := make([]*entities.User, 50)
				for i := 0; i < 50; i++ {
					users[i] = createTestUser(
						"perf-user-"+string(rune('0'+(i%10))),
						"perf"+string(rune('0'+(i%10)))+"@example.com",
						"Performance User "+string(rune('A'+(i%26))),
					)
				}
				stats := map[string]int{
					"total":                       1000,
					"active":                      950,
					"domains":                     25,
					"avg_days_since_registration": 120,
				}

				// When
				start := time.Now()
				component := UsersPage(users, stats)
				err := component.Render(ctx, buffer)
				duration := time.Since(start)

				// Then
				Expect(err).To(BeNil())
				Expect(duration).To(BeNumerically("<", 100*time.Millisecond))

				html := buffer.String()
				Expect(html).To(ContainSubstring("Users (50)"))
				Expect(html).To(ContainSubstring("1000")) // total stat
			})

			It("should maintain HTML structure integrity", func() {
				// Given
				users := []*entities.User{createTestUser("struct-user", "struct@example.com", "Structure User")}
				stats := createTestStats()

				// When
				component := UsersPage(users, stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Basic HTML validation
				Expect(strings.Count(html, "<html")).To(Equal(1))
				Expect(strings.Count(html, "</html>")).To(Equal(1))
				Expect(strings.Count(html, "<head")).To(Equal(1))
				Expect(strings.Count(html, "</head>")).To(Equal(1))
				Expect(strings.Count(html, "<body")).To(Equal(1))
				Expect(strings.Count(html, "</body>")).To(Equal(1))

				// Should contain proper section nesting
				Expect(html).To(ContainSubstring("<main"))
				Expect(html).To(ContainSubstring("</main>"))
			})

			It("should handle special characters and XSS prevention", func() {
				// Given
				users := []*entities.User{
					createTestUser(
						"xss-test-user",
						"xss@example.com",
						"<script>alert('XSS')</script>Test User",
					),
				}
				stats := createTestStats()

				// When
				component := UsersPage(users, stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should properly escape dangerous content
				Expect(html).ToNot(ContainSubstring("<script>alert"))
				Expect(html).To(ContainSubstring("&lt;script&gt;"))
			})
		})

		Context("when testing accessibility features", func() {
			It("should include proper form labels and ARIA attributes", func() {
				// Given
				users := []*entities.User{createTestUser("a11y-user", "a11y@example.com", "A11y User")}
				stats := createTestStats()

				// When
				component := UsersContent(users, stats)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain screen reader labels
				Expect(html).To(ContainSubstring("sr-only"))
				Expect(html).To(ContainSubstring("Search users"))
				Expect(html).To(ContainSubstring("Filter by domain"))

				// Should contain proper input types
				Expect(html).To(ContainSubstring("type=\"text\""))

				// Should contain proper button structure
				Expect(html).To(ContainSubstring("inline-flex items-center"))
			})
		})
	})
})
