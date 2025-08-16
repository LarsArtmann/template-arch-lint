// Package middleware provides HTTP middleware for cross-cutting concerns.
package middleware

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Security header constants.
const (
	// CORS and timing constants.
	DefaultMaxAge       = 86400    // 24 hours in seconds
	HSTSMaxAgeYear      = 31536000 // 1 year in seconds
	HTTPStatusNoContent = 204

	// CSP directive values.
	CSPSelf             = "'self'"
	CSPSelfUnsafeInline = "'self' 'unsafe-inline'"
	CSPSelfUnsafeEval   = "'self' 'unsafe-inline' 'unsafe-eval'"
	CSPNone             = "'none'"
	CSPDataHTTPS        = "'self' data: https:"
	CSPWebSockets       = "'self' ws: wss:"

	// Header names.
	HeaderOrigin                   = "Origin"
	HeaderAccessControlAllowOrigin = "Access-Control-Allow-Origin"
	HeaderContentSecurityPolicy    = "Content-Security-Policy"
	HeaderStrictTransportSecurity  = "Strict-Transport-Security"
	HeaderXFrameOptions            = "X-Frame-Options"
	HeaderXContentTypeOptions      = "X-Content-Type-Options"

	// Special values.
	CORSWildcard = "*"
	FirstIndex   = 0
	MinLength    = 1

	// Frame options values.
	FrameOptionsDeny = "DENY"
)

// SecurityHeadersConfig contains configuration for security headers middleware.
type SecurityHeadersConfig struct {
	// CORS settings
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int

	// CSP settings
	DefaultSrc string
	ScriptSrc  string
	StyleSrc   string
	ImgSrc     string
	ConnectSrc string
	FontSrc    string
	ObjectSrc  string
	MediaSrc   string
	FrameSrc   string

	// HSTS settings
	HSTSMaxAge            int
	HSTSIncludeSubDomains bool
	HSTSPreload           bool

	// Other security headers
	XFrameOptions       string
	XContentTypeOptions string
	XSSProtection       string
	ReferrerPolicy      string
	PermissionsPolicy   string

	// Configuration flags
	EnableCORS         bool
	EnableCSP          bool
	EnableHSTS         bool
	EnableOtherHeaders bool

	// Logger for middleware logging
	Logger *slog.Logger
}

// DefaultSecurityHeadersConfig returns a default secure configuration.
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		// CORS defaults
		AllowedOrigins: []string{CORSWildcard},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD",
		},
		AllowedHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization",
			"X-Requested-With", "X-Correlation-ID",
		},
		AllowCredentials: false,
		MaxAge:           DefaultMaxAge,

		// CSP defaults - restrictive but functional
		DefaultSrc: CSPSelf,
		ScriptSrc:  CSPSelfUnsafeInline,
		StyleSrc:   CSPSelfUnsafeInline,
		ImgSrc:     CSPDataHTTPS,
		ConnectSrc: CSPSelf,
		FontSrc:    CSPSelf,
		ObjectSrc:  CSPNone,
		MediaSrc:   CSPSelf,
		FrameSrc:   CSPNone,

		// HSTS defaults
		HSTSMaxAge:            HSTSMaxAgeYear,
		HSTSIncludeSubDomains: true,
		HSTSPreload:           false,

		// Other security headers defaults
		XFrameOptions:       FrameOptionsDeny,
		XContentTypeOptions: "nosniff",
		XSSProtection:       "1; mode=block",
		ReferrerPolicy:      "strict-origin-when-cross-origin",
		PermissionsPolicy:   "geolocation=(), microphone=(), camera=()",

		// Enable all security features by default
		EnableCORS:         true,
		EnableCSP:          true,
		EnableHSTS:         true,
		EnableOtherHeaders: true,

		Logger: slog.Default(),
	}
}

// DevelopmentSecurityHeadersConfig returns a development-friendly config.
func DevelopmentSecurityHeadersConfig() SecurityHeadersConfig {
	config := DefaultSecurityHeadersConfig()

	// More permissive CORS for development
	config.AllowedOrigins = []string{CORSWildcard}
	config.AllowCredentials = false

	// More permissive CSP for development
	config.ScriptSrc = CSPSelfUnsafeEval
	config.StyleSrc = CSPSelfUnsafeInline
	config.ConnectSrc = CSPWebSockets

	// Disable HSTS in development (HTTPS not typically used)
	config.EnableHSTS = false

	// Less restrictive frame options for development
	config.XFrameOptions = "SAMEORIGIN"

	return config
}

// ProductionSecurityHeadersConfig returns a production-ready config.
func ProductionSecurityHeadersConfig(
	allowedOrigins []string,
) SecurityHeadersConfig {
	config := DefaultSecurityHeadersConfig()

	// Strict CORS for production
	if len(allowedOrigins) > FirstIndex {
		config.AllowedOrigins = allowedOrigins
	} else {
		config.AllowedOrigins = []string{} // No origins allowed by default
	}
	config.AllowCredentials = true

	// Strict CSP for production
	config.ScriptSrc = CSPSelf
	config.StyleSrc = CSPSelf
	config.ConnectSrc = CSPSelf

	// Enable HSTS with preload for production
	config.EnableHSTS = true
	config.HSTSPreload = true

	// Strict frame options
	config.XFrameOptions = FrameOptionsDeny

	return config
}

// SecurityHeaders returns middleware that adds comprehensive security headers.
func SecurityHeaders(config ...SecurityHeadersConfig) gin.HandlerFunc {
	cfg := DefaultSecurityHeadersConfig()
	if len(config) > FirstIndex {
		cfg = config[FirstIndex]
	}

	return func(c *gin.Context) {
		// Handle preflight requests first
		if c.Request.Method == "OPTIONS" && cfg.EnableCORS {
			setCORSHeaders(c, &cfg)
			c.AbortWithStatus(HTTPStatusNoContent)
			return
		}

		// Set security headers
		if cfg.EnableCORS {
			setCORSHeaders(c, &cfg)
		}

		if cfg.EnableCSP {
			setCSPHeaders(c, &cfg)
		}

		if cfg.EnableHSTS {
			setHSTSHeaders(c, &cfg)
		}

		if cfg.EnableOtherHeaders {
			setOtherSecurityHeaders(c, &cfg)
		}

		cfg.Logger.Debug("Security headers applied",
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"origin", c.Request.Header.Get("Origin"))

		c.Next()
	}
}

// setCORSHeaders sets Cross-Origin Resource Sharing headers.
func setCORSHeaders(c *gin.Context, cfg *SecurityHeadersConfig) {
	origin := c.Request.Header.Get(HeaderOrigin)

	// Check for wildcard first
	if len(cfg.AllowedOrigins) == MinLength && cfg.AllowedOrigins[FirstIndex] == CORSWildcard {
		c.Header(HeaderAccessControlAllowOrigin, CORSWildcard)
	} else if isOriginAllowed(origin, cfg.AllowedOrigins) {
		c.Header(HeaderAccessControlAllowOrigin, origin)
	}

	// Set other CORS headers
	c.Header("Access-Control-Allow-Methods",
		strings.Join(cfg.AllowedMethods, ", "))
	c.Header("Access-Control-Allow-Headers",
		strings.Join(cfg.AllowedHeaders, ", "))

	if cfg.AllowCredentials {
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	if cfg.MaxAge > FirstIndex {
		c.Header("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
	}

	// Expose correlation ID for client debugging
	c.Header("Access-Control-Expose-Headers", "X-Correlation-ID")
}

// setCSPHeaders sets Content Security Policy headers.
func setCSPHeaders(c *gin.Context, cfg *SecurityHeadersConfig) {
	cspDirectives := []string{
		"default-src " + cfg.DefaultSrc,
		"script-src " + cfg.ScriptSrc,
		"style-src " + cfg.StyleSrc,
		"img-src " + cfg.ImgSrc,
		"connect-src " + cfg.ConnectSrc,
		"font-src " + cfg.FontSrc,
		"object-src " + cfg.ObjectSrc,
		"media-src " + cfg.MediaSrc,
		"frame-src " + cfg.FrameSrc,
	}

	csp := strings.Join(cspDirectives, "; ")
	c.Header(HeaderContentSecurityPolicy, csp)

	// Also set the report-only header for monitoring in production
	c.Header("Content-Security-Policy-Report-Only", csp)
}

// setHSTSHeaders sets HTTP Strict Transport Security headers.
func setHSTSHeaders(c *gin.Context, cfg *SecurityHeadersConfig) {
	hstsValue := "max-age=" + strconv.Itoa(cfg.HSTSMaxAge)

	if cfg.HSTSIncludeSubDomains {
		hstsValue += "; includeSubDomains"
	}

	if cfg.HSTSPreload {
		hstsValue += "; preload"
	}

	c.Header(HeaderStrictTransportSecurity, hstsValue)
}

// setOtherSecurityHeaders sets various other security headers.
func setOtherSecurityHeaders(c *gin.Context, cfg *SecurityHeadersConfig) {
	c.Header(HeaderXFrameOptions, cfg.XFrameOptions)
	c.Header(HeaderXContentTypeOptions, cfg.XContentTypeOptions)
	c.Header("X-XSS-Protection", cfg.XSSProtection)
	c.Header("Referrer-Policy", cfg.ReferrerPolicy)
	c.Header("Permissions-Policy", cfg.PermissionsPolicy)

	// Prevent MIME type sniffing
	c.Header("X-Download-Options", "noopen")

	// Prevent clickjacking
	c.Header("X-Permitted-Cross-Domain-Policies", "none")

	// Remove server information
	c.Header("Server", "")
}

// isOriginAllowed checks if the given origin is allowed.
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == CORSWildcard || allowed == origin {
			return true
		}

		// Support wildcard subdomains (e.g., "*.example.com")
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*.")
			// Extract just the domain part from the origin URL
			originDomain := extractDomainFromOrigin(origin)
			if strings.HasSuffix(originDomain, "."+domain) || originDomain == domain {
				return true
			}
		}
	}

	return false
}

// extractDomainFromOrigin extracts the domain from an origin URL.
func extractDomainFromOrigin(origin string) string {
	// Remove protocol (http:// or https://)
	if strings.HasPrefix(origin, "http://") {
		origin = strings.TrimPrefix(origin, "http://")
	} else if strings.HasPrefix(origin, "https://") {
		origin = strings.TrimPrefix(origin, "https://")
	}

	// Remove port if present
	if idx := strings.Index(origin, ":"); idx > FirstIndex {
		origin = origin[:idx]
	}

	// Remove path if present
	if idx := strings.Index(origin, "/"); idx > FirstIndex {
		origin = origin[:idx]
	}

	return origin
}

// WithSecurityHeaders creates security headers middleware with env defaults.
func WithSecurityHeaders(
	environment string,
	allowedOrigins []string,
	logger *slog.Logger,
) gin.HandlerFunc {
	var config SecurityHeadersConfig

	switch environment {
	case "production":
		config = ProductionSecurityHeadersConfig(allowedOrigins)
	case "development", "debug":
		config = DevelopmentSecurityHeadersConfig()
	default:
		config = DefaultSecurityHeadersConfig()
	}

	config.Logger = logger

	return SecurityHeaders(config)
}

// WithCORSOnly is a convenience function for CORS-only middleware.
func WithCORSOnly(allowedOrigins []string, logger *slog.Logger) gin.HandlerFunc {
	config := DefaultSecurityHeadersConfig()
	config.AllowedOrigins = allowedOrigins
	config.EnableCSP = false
	config.EnableHSTS = false
	config.EnableOtherHeaders = false
	config.Logger = logger

	return SecurityHeaders(config)
}

// WithStrictSecurity is a convenience function for maximum security middleware.
func WithStrictSecurity(allowedOrigins []string, logger *slog.Logger) gin.HandlerFunc {
	config := ProductionSecurityHeadersConfig(allowedOrigins)
	config.Logger = logger

	// Extra strict settings
	config.XFrameOptions = FrameOptionsDeny
	config.ScriptSrc = CSPSelf
	config.StyleSrc = CSPSelf
	config.ObjectSrc = CSPNone
	config.FrameSrc = CSPNone

	return SecurityHeaders(config)
}
