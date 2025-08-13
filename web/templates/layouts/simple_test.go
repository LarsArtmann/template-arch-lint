package layouts

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestBaseLayoutSimple(t *testing.T) {
	ctx := context.Background()
	buffer := &bytes.Buffer{}

	// Create simple content component
	content := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte("<p>Test content</p>"))
		return err
	})

	component := BaseLayout("Test Title", content)
	err := component.Render(ctx, buffer)

	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	html := buffer.String()
	t.Logf("Rendered HTML: %s", html)

	if !strings.Contains(html, "Test Title") {
		t.Errorf("Expected HTML to contain 'Test Title', but it didn't")
	}

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Errorf("Expected HTML to contain DOCTYPE, but it didn't. Got: %s", html[:100])
	}

	if !strings.Contains(html, "<p>Test content</p>") {
		t.Errorf("Expected HTML to contain test content, but it didn't")
	}
}
