// Profiling middleware provides secure pprof endpoints for performance analysis
package middleware

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

// ProfilingHandler wraps the standard pprof handlers for Gin
type ProfilingHandler struct {
	handler http.HandlerFunc
}

// NewProfilingHandler creates a new profiling handler
func NewProfilingHandler(handler http.HandlerFunc) *ProfilingHandler {
	return &ProfilingHandler{handler: handler}
}

// ServeHTTP implements the http.Handler interface
func (p *ProfilingHandler) ServeHTTP(ctx *gin.Context) {
	p.handler.ServeHTTP(ctx.Writer, ctx.Request)
}

// ProfilingAuthMiddleware provides basic authentication for profiling endpoints
// In production, this should be replaced with proper authentication
func ProfilingAuthMiddleware(username, password string) gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		username: password,
	})
}

// IPWhitelistMiddleware restricts profiling access to specific IPs
func IPWhitelistMiddleware(allowedIPs []string) gin.HandlerFunc {
	allowedIPMap := make(map[string]bool)
	for _, ip := range allowedIPs {
		allowedIPMap[ip] = true
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		// Always allow localhost for development
		if clientIP == "127.0.0.1" || clientIP == "::1" {
			c.Next()
			return
		}
		
		// Check if IP is in whitelist
		if !allowedIPMap[clientIP] {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		
		c.Next()
	})
}

// SetupProfilingRoutes sets up secure profiling endpoints
func SetupProfilingRoutes(router *gin.Engine, enableProfiling bool, authUsername, authPassword string, allowedIPs []string) {
	if !enableProfiling {
		return
	}

	// Create profiling route group with security middleware
	profilingGroup := router.Group("/debug/pprof")
	
	// Add IP whitelist middleware
	if len(allowedIPs) > 0 {
		profilingGroup.Use(IPWhitelistMiddleware(allowedIPs))
	}
	
	// Add basic auth if credentials provided
	if authUsername != "" && authPassword != "" {
		profilingGroup.Use(ProfilingAuthMiddleware(authUsername, authPassword))
	}

	// Register pprof endpoints
	profilingGroup.GET("/", gin.WrapH(http.HandlerFunc(pprof.Index)))
	profilingGroup.GET("/cmdline", gin.WrapH(http.HandlerFunc(pprof.Cmdline)))
	profilingGroup.GET("/profile", gin.WrapH(http.HandlerFunc(pprof.Profile)))
	profilingGroup.POST("/symbol", gin.WrapH(http.HandlerFunc(pprof.Symbol)))
	profilingGroup.GET("/symbol", gin.WrapH(http.HandlerFunc(pprof.Symbol)))
	profilingGroup.GET("/trace", gin.WrapH(http.HandlerFunc(pprof.Trace)))
	
	// Individual profiling endpoints
	profilingGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
	profilingGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
	profilingGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
	profilingGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
	profilingGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
	profilingGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
}

// ProfilingConfig holds configuration for profiling endpoints
type ProfilingConfig struct {
	Enabled        bool     `json:"enabled"`
	AuthUsername   string   `json:"auth_username"`
	AuthPassword   string   `json:"auth_password"`
	AllowedIPs     []string `json:"allowed_ips"`
	EnableSampling bool     `json:"enable_sampling"`
	SampleRate     int      `json:"sample_rate"` // Sample rate for CPU profiling
}

// DefaultProfilingConfig returns default profiling configuration
func DefaultProfilingConfig() *ProfilingConfig {
	return &ProfilingConfig{
		Enabled:        false,
		AuthUsername:   "",
		AuthPassword:   "",
		AllowedIPs:     []string{"127.0.0.1", "::1"},
		EnableSampling: true,
		SampleRate:     100, // 100Hz sampling rate
	}
}