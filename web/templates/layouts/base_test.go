package layouts

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"github.com/a-h/templ"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BaseLayout Template", func() {
	var (
		ctx           context.Context
		buffer        *bytes.Buffer
		sampleContent templ.Component
	)

	BeforeEach(func() {
		ctx = context.Background()
		buffer = &bytes.Buffer{}
		// Create a simple content component for testing
		sampleContent = templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			_, err := w.Write([]byte("<p>Sample content for testing</p>"))
			return err
		})
	})

	Describe("BaseLayout rendering", func() {
		Context("with standard parameters", func() {
			It("should render complete HTML document structure", func() {
				// Given
				title := "Test Page"

				// When
				component := BaseLayout(title, sampleContent)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain proper HTML5 structure
				Expect(html).To(ContainSubstring("<!DOCTYPE html>"))
				Expect(html).To(ContainSubstring("<html lang=\"en\">"))
				Expect(html).To(ContainSubstring("<head>"))
				Expect(html).To(ContainSubstring("<body"))
				Expect(html).To(ContainSubstring("</html>"))

				// Should contain proper title
				Expect(html).To(ContainSubstring("<title>Test Page - User Management System</title>"))

				// Should contain content
				Expect(html).To(ContainSubstring("<p>Sample content for testing</p>"))
			})

			It("should include all required external dependencies", func() {
				// Given
				title := "Dependencies Test"

				// When
				component := BaseLayout(title, sampleContent)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should include TailwindCSS
				Expect(html).To(ContainSubstring("https://cdn.tailwindcss.com"))

				// Should include HTMX
				Expect(html).To(ContainSubstring("https://unpkg.com/htmx.org@2.0.3"))

				// Should include Heroicons
				Expect(html).To(ContainSubstring("heroicons"))
			})

			It("should include proper meta tags", func() {
				// Given
				title := "Meta Test"

				// When
				component := BaseLayout(title, sampleContent)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain proper meta tags
				Expect(html).To(ContainSubstring("<meta charset=\"UTF-8\"/>"))
				Expect(html).To(ContainSubstring("name=\"viewport\" content=\"width=device-width, initial-scale=1.0\""))
			})

			It("should include navigation component", func() {
				// Given
				title := "Navigation Test"

				// When
				component := BaseLayout(title, sampleContent)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain navigation elements
				Expect(html).To(ContainSubstring("User Management"))
				Expect(html).To(ContainSubstring("href=\"/users\""))
				Expect(html).To(ContainSubstring("href=\"/users/new\""))
			})

			It("should include HTMX configuration and JavaScript", func() {
				// Given
				title := "JavaScript Test"

				// When
				component := BaseLayout(title, sampleContent)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain HTMX configuration
				Expect(html).To(ContainSubstring("htmx.config.defaultSwapStyle"))
				Expect(html).To(ContainSubstring("htmx:beforeRequest"))
				Expect(html).To(ContainSubstring("htmx:afterRequest"))

				// Should contain toast notification system
				Expect(html).To(ContainSubstring("window.showToast"))
				Expect(html).To(ContainSubstring("toast-container"))
			})

			It("should include loading indicator", func() {
				// Given
				title := "Loading Test"

				// When
				component := BaseLayout(title, sampleContent)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain loading indicator
				Expect(html).To(ContainSubstring("id=\"loading-indicator\""))
				Expect(html).To(ContainSubstring("Loading..."))
				Expect(html).To(ContainSubstring("animate-spin"))
			})
		})

		Context("with edge cases", func() {
			It("should handle empty title gracefully", func() {
				// Given
				title := ""

				// When
				component := BaseLayout(title, sampleContent)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()
				Expect(html).To(ContainSubstring("<title> - User Management System</title>"))
			})

			It("should handle special characters in title", func() {
				// Given
				title := "Test & Special <Characters> \"Quotes\""

				// When
				component := BaseLayout(title, sampleContent)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()
				// Should be properly escaped
				Expect(html).To(ContainSubstring("Test &amp; Special &lt;Characters&gt; &#34;Quotes&#34; - User Management System"))
			})

			It("should handle nil content component", func() {
				// Given
				title := "Nil Content Test"
				var nilContent templ.Component

				// When
				component := BaseLayout(title, nilContent)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()
				Expect(html).To(ContainSubstring("<main class=\"container mx-auto px-4 py-8\">"))
			})

			It("should handle very long title", func() {
				// Given
				title := strings.Repeat("Very Long Title ", 50)

				// When
				component := BaseLayout(title, sampleContent)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()
				Expect(html).To(ContainSubstring("<title>" + title + " - User Management System</title>"))
			})
		})
	})

	Describe("Navigation component", func() {
		Context("when rendered independently", func() {
			It("should render navigation structure correctly", func() {
				// Given
				buffer := &bytes.Buffer{}

				// When
				component := Navigation()
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain navigation structure
				Expect(html).To(ContainSubstring("<nav class=\"bg-white shadow-sm border-b border-gray-200\">"))
				Expect(html).To(ContainSubstring("User Management"))

				// Should contain navigation links
				Expect(html).To(ContainSubstring("href=\"/users\""))
				Expect(html).To(ContainSubstring("href=\"/users/new\""))
				Expect(html).To(ContainSubstring("Users"))
				Expect(html).To(ContainSubstring("Add User"))

				// Should contain mobile menu
				Expect(html).To(ContainSubstring("id=\"mobile-menu\""))
				Expect(html).To(ContainSubstring("toggleMobileMenu()"))
			})

			It("should include responsive classes", func() {
				// Given
				buffer := &bytes.Buffer{}

				// When
				component := Navigation()
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain responsive classes
				Expect(html).To(ContainSubstring("hidden md:flex"))
				Expect(html).To(ContainSubstring("md:hidden"))
				Expect(html).To(ContainSubstring("hidden md:hidden"))
			})

			It("should contain proper accessibility attributes", func() {
				// Given
				buffer := &bytes.Buffer{}

				// When
				component := Navigation()
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain ARIA attributes and proper button structure
				Expect(html).To(ContainSubstring("type=\"button\""))
				Expect(html).To(ContainSubstring("focus:outline-none"))
			})
		})
	})

	Describe("Performance and structure", func() {
		Context("when rendering with complex content", func() {
			It("should render efficiently with multiple components", func() {
				// Given
				complexContent := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
					for i := 0; i < 100; i++ {
						_, err := w.Write([]byte("<div class=\"test-item\">Item " + string(rune('0'+i%10)) + "</div>"))
						if err != nil {
							return err
						}
					}
					return nil
				})

				// When
				start := time.Now()
				component := BaseLayout("Performance Test", complexContent)
				err := component.Render(ctx, buffer)
				duration := time.Since(start)

				// Then
				Expect(err).To(BeNil())
				Expect(duration).To(BeNumerically("<", 100*time.Millisecond))

				html := buffer.String()
				Expect(strings.Count(html, "test-item")).To(Equal(100))
			})
		})

		Context("HTML structure validation", func() {
			It("should produce valid HTML structure", func() {
				// Given
				title := "HTML Validation Test"

				// When
				component := BaseLayout(title, sampleContent)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Basic HTML structure validation
				Expect(strings.Count(html, "<html")).To(Equal(1))
				Expect(strings.Count(html, "</html>")).To(Equal(1))
				Expect(strings.Count(html, "<head")).To(Equal(1))
				Expect(strings.Count(html, "</head>")).To(Equal(1))
				Expect(strings.Count(html, "<body")).To(Equal(1))
				Expect(strings.Count(html, "</body>")).To(Equal(1))

				// Should have proper nesting
				htmlIndex := strings.Index(html, "<html")
				headIndex := strings.Index(html, "<head")
				bodyIndex := strings.Index(html, "<body")

				Expect(headIndex).To(BeNumerically(">", htmlIndex))
				Expect(bodyIndex).To(BeNumerically(">", headIndex))
			})
		})
	})

	Describe("CSS and styling", func() {
		Context("when checking for styling classes", func() {
			It("should include all necessary Tailwind classes", func() {
				// Given
				title := "Styling Test"

				// When
				component := BaseLayout(title, sampleContent)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain layout classes
				Expect(html).To(ContainSubstring("bg-gray-50"))
				Expect(html).To(ContainSubstring("min-h-screen"))
				Expect(html).To(ContainSubstring("container"))
				Expect(html).To(ContainSubstring("mx-auto"))

				// Should contain component-specific classes
				Expect(html).To(ContainSubstring("shadow-sm"))
				Expect(html).To(ContainSubstring("border-b"))
				Expect(html).To(ContainSubstring("transition-colors"))
			})

			It("should include custom CSS for HTMX", func() {
				// Given
				title := "HTMX CSS Test"

				// When
				component := BaseLayout(title, sampleContent)
				err := component.Render(ctx, buffer)

				// Then
				Expect(err).To(BeNil())
				html := buffer.String()

				// Should contain HTMX-specific styles
				Expect(html).To(ContainSubstring(".loading"))
				Expect(html).To(ContainSubstring(".htmx-indicator"))
				Expect(html).To(ContainSubstring(".htmx-request"))
			})
		})
	})
})
