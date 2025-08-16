// Package middleware provides HTTP middleware for cross-cutting concerns.
package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitType represents different types of rate limits.
type RateLimitType string

const (
	// GeneralRateLimit for general API endpoints - 100 requests per minute.
	GeneralRateLimit RateLimitType = "general"
	// AuthRateLimit for authentication endpoints - 10 requests per minute.
	AuthRateLimit RateLimitType = "auth"

	// Rate limiting constants.
	defaultGeneralRequestsPerMinute = 100.0
	defaultGeneralBurst             = 10
	defaultAuthRequestsPerMinute    = 10.0
	defaultAuthBurst                = 3
	defaultCleanupIntervalMinutes   = 5
	zeroValue                       = 0
	secondsPerMinute                = 60.0
	retryAfterSeconds               = 60
	base10                          = 10
	notFoundIndex                   = -1

	// Logging field constants.
	clientIDField  = "client_id"
	limitTypeField = "limit_type"
)

// RateLimitConfig holds configuration for rate limiting.
type RateLimitConfig struct {
	// GeneralRPS is requests per second for general endpoints
	GeneralRPS float64
	// GeneralBurst is burst size for general endpoints
	GeneralBurst int
	// AuthRPS is requests per second for auth endpoints
	AuthRPS float64
	// AuthBurst is burst size for auth endpoints
	AuthBurst int
	// Logger for middleware logging
	Logger *slog.Logger
	// CleanupInterval for cleaning up expired limiters
	CleanupInterval time.Duration
}

// DefaultRateLimitConfig returns default rate limiting configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		GeneralRPS:      defaultGeneralRequestsPerMinute / secondsPerMinute,
		GeneralBurst:    defaultGeneralBurst,
		AuthRPS:         defaultAuthRequestsPerMinute / secondsPerMinute,
		AuthBurst:       defaultAuthBurst,
		Logger:          slog.Default(),
		CleanupInterval: defaultCleanupIntervalMinutes * time.Minute,
	}
}

// clientLimiter holds rate limiter and last access time for a client.
type clientLimiter struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// RateLimiterStore manages rate limiters for different clients.
type RateLimiterStore struct {
	mu           sync.RWMutex
	limiters     map[string]*clientLimiter
	generalRPS   float64
	generalBurst int
	authRPS      float64
	authBurst    int
	logger       *slog.Logger
}

// NewRateLimiterStore creates a new rate limiter store.
func NewRateLimiterStore(config RateLimitConfig) *RateLimiterStore {
	store := &RateLimiterStore{
		limiters:     make(map[string]*clientLimiter),
		generalRPS:   config.GeneralRPS,
		generalBurst: config.GeneralBurst,
		authRPS:      config.AuthRPS,
		authBurst:    config.AuthBurst,
		logger:       config.Logger,
	}

	// Start cleanup goroutine only if cleanup interval is positive
	cleanupInterval := config.CleanupInterval
	if cleanupInterval <= zeroValue {
		cleanupInterval = defaultCleanupIntervalMinutes * time.Minute
	}
	go store.cleanupExpiredLimiters(cleanupInterval)

	return store
}

// getLimiter retrieves or creates a rate limiter for a client.
func (store *RateLimiterStore) getLimiter(clientID string, limitType RateLimitType) *rate.Limiter {
	store.mu.Lock()
	defer store.mu.Unlock()

	key := fmt.Sprintf("%s:%s", clientID, limitType)

	limiterEntry, exists := store.limiters[key]
	if !exists {
		var limiter *rate.Limiter
		switch limitType {
		case AuthRateLimit:
			limiter = rate.NewLimiter(rate.Limit(store.authRPS), store.authBurst)
		default:
			limiter = rate.NewLimiter(rate.Limit(store.generalRPS), store.generalBurst)
		}

		limiterEntry = &clientLimiter{
			limiter:    limiter,
			lastAccess: time.Now(),
		}
		store.limiters[key] = limiterEntry

		store.logger.Debug("Created new rate limiter",
			clientIDField, clientID,
			limitTypeField, limitType,
			"key", key)
	} else {
		limiterEntry.lastAccess = time.Now()
	}

	return limiterEntry.limiter
}

// cleanupExpiredLimiters removes unused rate limiters to prevent memory leaks.
func (store *RateLimiterStore) cleanupExpiredLimiters(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		store.performCleanup()
	}
}

// performCleanup removes expired limiters to reduce complexity.
func (store *RateLimiterStore) performCleanup() {
	store.mu.Lock()
	defer store.mu.Unlock()

	expiredKeys := store.findExpiredKeys()
	store.removeExpiredKeys(expiredKeys)
	store.logCleanupResult(expiredKeys)
}

// findExpiredKeys identifies rate limiters that haven't been accessed recently.
func (store *RateLimiterStore) findExpiredKeys() []string {
	now := time.Now()
	expiredKeys := make([]string, zeroValue)

	for key, limiterEntry := range store.limiters {
		if now.Sub(limiterEntry.lastAccess) > time.Hour {
			expiredKeys = append(expiredKeys, key)
		}
	}

	return expiredKeys
}

// removeExpiredKeys removes the specified keys from the limiters map.
func (store *RateLimiterStore) removeExpiredKeys(expiredKeys []string) {
	for _, key := range expiredKeys {
		delete(store.limiters, key)
	}
}

// logCleanupResult logs the cleanup operation if any keys were removed.
func (store *RateLimiterStore) logCleanupResult(expiredKeys []string) {
	if len(expiredKeys) > zeroValue {
		store.logger.Debug("Cleaned up expired rate limiters",
			"count", len(expiredKeys))
	}
}

// getClientID extracts client identifier from request (IP address).
func getClientID(c *gin.Context) string {
	// Try to get real IP from headers first
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if comma := strings.Index(ip, ","); comma != notFoundIndex {
			ip = strings.TrimSpace(ip[:comma])
		}
		return ip
	}

	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}

	// Fall back to remote address
	return c.ClientIP()
}

// getRateLimitType determines the rate limit type based on the request path.
func getRateLimitType(path string) RateLimitType {
	// Define auth endpoints that should have stricter rate limiting
	authPaths := []string{
		"/api/v1/auth/",
		"/auth/",
		"/login",
		"/register",
		"/reset-password",
		"/verify-email",
	}

	// Check if path matches any auth patterns
	for _, authPath := range authPaths {
		if strings.Contains(path, authPath) {
			return AuthRateLimit
		}
	}

	// Also apply auth rate limits to user creation/modification endpoints
	// These are sensitive operations that should be rate limited more strictly
	if path == "/api/v1/users" || path == "/users" {
		return AuthRateLimit
	}

	return GeneralRateLimit
}

// setRateLimitHeaders sets rate limit headers in the response.
func setRateLimitHeaders(c *gin.Context, limiter *rate.Limiter, limitType RateLimitType, allowed bool) {
	var limit int
	var window string

	switch limitType {
	case AuthRateLimit:
		limit = int(defaultAuthRequestsPerMinute)
		window = strconv.Itoa(retryAfterSeconds)
	default:
		limit = int(defaultGeneralRequestsPerMinute)
		window = strconv.Itoa(retryAfterSeconds)
	}

	// X-RateLimit-Limit: Maximum number of requests allowed in the time window
	c.Header("X-RateLimit-Limit", strconv.Itoa(limit))

	// X-RateLimit-Window: Time window in seconds
	c.Header("X-RateLimit-Window", window)

	// X-RateLimit-Remaining: Number of requests remaining in current window
	// This is an approximation based on current tokens
	tokens := limiter.Tokens()
	remaining := int(tokens)
	if remaining < zeroValue {
		remaining = zeroValue
	}
	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

	// X-RateLimit-Reset: Unix timestamp when the rate limit resets
	// For token bucket, this is an approximation
	resetTime := time.Now().Add(time.Minute).Unix()
	c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, base10))

	// X-RateLimit-Type: Type of rate limit applied
	c.Header("X-RateLimit-Type", string(limitType))
}

// RateLimiter returns a rate limiting middleware.
func RateLimiter(config ...RateLimitConfig) gin.HandlerFunc {
	cfg := DefaultRateLimitConfig()
	if len(config) > zeroValue {
		cfg = config[zeroValue]
	}

	store := NewRateLimiterStore(cfg)

	return func(c *gin.Context) {
		clientID := getClientID(c)
		limitType := getRateLimitType(c.Request.URL.Path)
		limiter := store.getLimiter(clientID, limitType)

		// Check if request is allowed
		if !limiter.Allow() {
			cfg.Logger.Warn("Rate limit exceeded",
				clientIDField, clientID,
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				limitTypeField, limitType,
				"user_agent", c.GetHeader("User-Agent"))

			setRateLimitHeaders(c, limiter, limitType, false)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"message":     fmt.Sprintf("Rate limit exceeded for %s endpoints", limitType),
				"retry_after": retryAfterSeconds,
			})
			c.Abort()
			return
		}

		// Set rate limit headers for successful requests
		setRateLimitHeaders(c, limiter, limitType, true)

		cfg.Logger.Debug("Request allowed",
			clientIDField, clientID,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			limitTypeField, limitType)

		c.Next()
	}
}

// WithRateLimit creates rate limit middleware with logger.
func WithRateLimit(logger *slog.Logger) gin.HandlerFunc {
	return RateLimiter(RateLimitConfig{
		Logger: logger,
	})
}
