package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDefaultSecurityHeadersConfig(t *testing.T) {
	config := DefaultSecurityHeadersConfig()

	assert.Equal(t, []string{"*"}, config.AllowedOrigins)
	assert.Contains(t, config.AllowedMethods, "GET")
	assert.Contains(t, config.AllowedMethods, "POST")
	assert.Contains(t, config.AllowedHeaders, "Content-Type")
	assert.Equal(t, "'self'", config.DefaultSrc)
	assert.Equal(t, "'self' 'unsafe-inline'", config.ScriptSrc)
	assert.Equal(t, 31536000, config.HSTSMaxAge)
	assert.True(t, config.HSTSIncludeSubDomains)
	assert.Equal(t, "DENY", config.XFrameOptions)
	assert.Equal(t, "nosniff", config.XContentTypeOptions)
	assert.True(t, config.EnableCORS)
	assert.True(t, config.EnableCSP)
	assert.True(t, config.EnableHSTS)
	assert.True(t, config.EnableOtherHeaders)
}

func TestDevelopmentSecurityHeadersConfig(t *testing.T) {
	config := DevelopmentSecurityHeadersConfig()

	assert.Equal(t, []string{"*"}, config.AllowedOrigins)
	assert.Contains(t, config.ScriptSrc, "'unsafe-eval'")
	assert.Contains(t, config.ConnectSrc, "ws:")
	assert.False(t, config.EnableHSTS)
	assert.Equal(t, "SAMEORIGIN", config.XFrameOptions)
}

func TestProductionSecurityHeadersConfig(t *testing.T) {
	allowedOrigins := []string{"https://example.com", "https://app.example.com"}
	config := ProductionSecurityHeadersConfig(allowedOrigins)

	assert.Equal(t, allowedOrigins, config.AllowedOrigins)
	assert.Equal(t, "'self'", config.ScriptSrc)
	assert.Equal(t, "'self'", config.StyleSrc)
	assert.True(t, config.EnableHSTS)
	assert.True(t, config.HSTSPreload)
	assert.Equal(t, "DENY", config.XFrameOptions)
}

func TestProductionSecurityHeadersConfigNoOrigins(t *testing.T) {
	config := ProductionSecurityHeadersConfig(nil)

	assert.Empty(t, config.AllowedOrigins)
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		config         SecurityHeadersConfig
		requestMethod  string
		requestOrigin  string
		expectedStatus int
		checkHeaders   func(t *testing.T, headers http.Header)
	}{
		{
			name:           "default config with GET request",
			config:         DefaultSecurityHeadersConfig(),
			requestMethod:  "GET",
			requestOrigin:  "https://example.com",
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, headers http.Header) {
				assert.Equal(t, "*", headers.Get("Access-Control-Allow-Origin"))
				assert.Contains(t, headers.Get("Content-Security-Policy"), "default-src 'self'")
				assert.Contains(t, headers.Get("Strict-Transport-Security"), "max-age=31536000")
				assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
				assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
				assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
			},
		},
		{
			name:           "production config with allowed origin",
			config:         ProductionSecurityHeadersConfig([]string{"https://example.com"}),
			requestMethod:  "GET",
			requestOrigin:  "https://example.com",
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, headers http.Header) {
				assert.Equal(t, "https://example.com", headers.Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "'self'", getCSPDirective(headers.Get("Content-Security-Policy"), "script-src"))
			},
		},
		{
			name:           "production config with disallowed origin",
			config:         ProductionSecurityHeadersConfig([]string{"https://example.com"}),
			requestMethod:  "GET",
			requestOrigin:  "https://malicious.com",
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, headers http.Header) {
				assert.Empty(t, headers.Get("Access-Control-Allow-Origin"))
			},
		},
		{
			name:           "OPTIONS preflight request",
			config:         DefaultSecurityHeadersConfig(),
			requestMethod:  "OPTIONS",
			requestOrigin:  "https://example.com",
			expectedStatus: http.StatusNoContent,
			checkHeaders: func(t *testing.T, headers http.Header) {
				assert.Equal(t, "*", headers.Get("Access-Control-Allow-Origin"))
				assert.Contains(t, headers.Get("Access-Control-Allow-Methods"), "GET")
				assert.Contains(t, headers.Get("Access-Control-Allow-Headers"), "Content-Type")
			},
		},
		{
			name: "disabled CORS",
			config: func() SecurityHeadersConfig {
				cfg := DefaultSecurityHeadersConfig()
				cfg.EnableCORS = false
				return cfg
			}(),
			requestMethod:  "GET",
			requestOrigin:  "https://example.com",
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, headers http.Header) {
				assert.Empty(t, headers.Get("Access-Control-Allow-Origin"))
			},
		},
		{
			name: "disabled CSP",
			config: func() SecurityHeadersConfig {
				cfg := DefaultSecurityHeadersConfig()
				cfg.EnableCSP = false
				return cfg
			}(),
			requestMethod:  "GET",
			requestOrigin:  "https://example.com",
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, headers http.Header) {
				assert.Empty(t, headers.Get("Content-Security-Policy"))
			},
		},
		{
			name: "disabled HSTS",
			config: func() SecurityHeadersConfig {
				cfg := DefaultSecurityHeadersConfig()
				cfg.EnableHSTS = false
				return cfg
			}(),
			requestMethod:  "GET",
			requestOrigin:  "https://example.com",
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, headers http.Header) {
				assert.Empty(t, headers.Get("Strict-Transport-Security"))
			},
		},
		{
			name: "disabled other security headers",
			config: func() SecurityHeadersConfig {
				cfg := DefaultSecurityHeadersConfig()
				cfg.EnableOtherHeaders = false
				return cfg
			}(),
			requestMethod:  "GET",
			requestOrigin:  "https://example.com",
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, headers http.Header) {
				assert.Empty(t, headers.Get("X-Frame-Options"))
				assert.Empty(t, headers.Get("X-Content-Type-Options"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(SecurityHeaders(tt.config))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "test"})
			})
			router.POST("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "test"})
			})

			req := httptest.NewRequest(tt.requestMethod, "/test", nil)
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkHeaders(t, w.Header())
		})
	}
}

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{
			name:           "exact match",
			origin:         "https://example.com",
			allowedOrigins: []string{"https://example.com", "https://other.com"},
			expected:       true,
		},
		{
			name:           "wildcard match",
			origin:         "https://example.com",
			allowedOrigins: []string{"*"},
			expected:       true,
		},
		{
			name:           "subdomain wildcard match",
			origin:         "https://api.example.com",
			allowedOrigins: []string{"*.example.com"},
			expected:       true,
		},
		{
			name:           "exact domain from wildcard",
			origin:         "https://example.com",
			allowedOrigins: []string{"*.example.com"},
			expected:       true,
		},
		{
			name:           "no match",
			origin:         "https://malicious.com",
			allowedOrigins: []string{"https://example.com"},
			expected:       false,
		},
		{
			name:           "empty origin",
			origin:         "",
			allowedOrigins: []string{"https://example.com"},
			expected:       false,
		},
		{
			name:           "subdomain no wildcard",
			origin:         "https://api.example.com",
			allowedOrigins: []string{"https://example.com"},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOriginAllowed(tt.origin, tt.allowedOrigins)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWithSecurityHeaders(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	tests := []struct {
		name           string
		environment    string
		allowedOrigins []string
		expectedCORS   bool
		expectedHSTS   bool
	}{
		{
			name:           "production environment",
			environment:    "production",
			allowedOrigins: []string{"https://example.com"},
			expectedCORS:   true,
			expectedHSTS:   true,
		},
		{
			name:           "development environment",
			environment:    "development",
			allowedOrigins: []string{"*"},
			expectedCORS:   true,
			expectedHSTS:   false,
		},
		{
			name:           "test environment",
			environment:    "test",
			allowedOrigins: []string{"*"},
			expectedCORS:   true,
			expectedHSTS:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(WithSecurityHeaders(tt.environment, tt.allowedOrigins, logger))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "test"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", "https://example.com")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			if tt.expectedCORS {
				assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
			}

			if tt.expectedHSTS {
				assert.NotEmpty(t, w.Header().Get("Strict-Transport-Security"))
			} else {
				assert.Empty(t, w.Header().Get("Strict-Transport-Security"))
			}
		})
	}
}

func TestWithCORSOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	router := gin.New()
	router.Use(WithCORSOnly([]string{"https://example.com"}, logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Content-Security-Policy"))
	assert.Empty(t, w.Header().Get("Strict-Transport-Security"))
	assert.Empty(t, w.Header().Get("X-Frame-Options"))
}

func TestWithStrictSecurity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	router := gin.New()
	router.Use(WithStrictSecurity([]string{"https://example.com"}, logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "script-src 'self'")
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "includeSubDomains")
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

func TestHSTSHeaderGeneration(t *testing.T) {
	tests := []struct {
		name              string
		maxAge            int
		includeSubDomains bool
		preload           bool
		expectedHeader    string
	}{
		{
			name:              "basic HSTS",
			maxAge:            31536000,
			includeSubDomains: false,
			preload:           false,
			expectedHeader:    "max-age=31536000",
		},
		{
			name:              "HSTS with subdomains",
			maxAge:            31536000,
			includeSubDomains: true,
			preload:           false,
			expectedHeader:    "max-age=31536000; includeSubDomains",
		},
		{
			name:              "HSTS with preload",
			maxAge:            31536000,
			includeSubDomains: true,
			preload:           true,
			expectedHeader:    "max-age=31536000; includeSubDomains; preload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			config := DefaultSecurityHeadersConfig()
			config.HSTSMaxAge = tt.maxAge
			config.HSTSIncludeSubDomains = tt.includeSubDomains
			config.HSTSPreload = tt.preload

			router := gin.New()
			router.Use(SecurityHeaders(config))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "test"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.expectedHeader, w.Header().Get("Strict-Transport-Security"))
		})
	}
}

func TestCSPHeaderGeneration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := DefaultSecurityHeadersConfig()
	config.ScriptSrc = "'self' 'unsafe-inline'"
	config.StyleSrc = "'self' 'unsafe-inline'"
	config.ImgSrc = "'self' data: https:"

	router := gin.New()
	router.Use(SecurityHeaders(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	csp := w.Header().Get("Content-Security-Policy")
	assert.Contains(t, csp, "default-src 'self'")
	assert.Contains(t, csp, "script-src 'self' 'unsafe-inline'")
	assert.Contains(t, csp, "style-src 'self' 'unsafe-inline'")
	assert.Contains(t, csp, "img-src 'self' data: https:")
	assert.Contains(t, csp, "object-src 'none'")
}

func TestCorrelationIDExposure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Access-Control-Expose-Headers"), "X-Correlation-ID")
}

// Helper function to extract CSP directive values
func getCSPDirective(csp, directive string) string {
	// Simple parser for CSP directives in tests
	if csp == "" {
		return ""
	}

	// This is a simplified parser for test purposes
	// In a real implementation, you might want a more robust parser
	directives := map[string]string{}
	parts := strings.Split(csp, "; ")
	for _, part := range parts {
		if idx := strings.Index(part, " "); idx > 0 {
			key := part[:idx]
			value := part[idx+1:]
			directives[key] = value
		}
	}

	return directives[directive]
}
