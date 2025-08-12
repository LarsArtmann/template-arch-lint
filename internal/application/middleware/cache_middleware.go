// HTTP caching middleware with intelligent cache strategies
package middleware

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/LarsArtmann/template-arch-lint/internal/observability"
)

// CacheMiddleware provides HTTP response caching
type CacheMiddleware struct {
	cacheManager *observability.CacheManager
	logger       *slog.Logger
	config       *CacheMiddlewareConfig
}

// CacheMiddlewareConfig holds cache middleware configuration
type CacheMiddlewareConfig struct {
	// Cache behavior
	DefaultTTL         time.Duration `json:"default_ttl"`
	MaxCacheSize       int64         `json:"max_cache_size"`
	
	// Cache strategies by route pattern
	RouteStrategies    map[string]*CacheStrategy `json:"route_strategies"`
	
	// Cache headers
	EnableETags        bool     `json:"enable_etags"`
	EnableLastModified bool     `json:"enable_last_modified"`
	VaryHeaders        []string `json:"vary_headers"`
	
	// Cache control
	CacheableStatus    []int    `json:"cacheable_status_codes"`
	CacheableMethods   []string `json:"cacheable_methods"`
	SkipPatterns       []string `json:"skip_patterns"`
}

// CacheStrategy defines caching strategy for specific routes
type CacheStrategy struct {
	TTL                time.Duration `json:"ttl"`
	VaryBy             []string      `json:"vary_by"` // Headers to vary cache by
	InvalidateOn       []string      `json:"invalidate_on"` // Events that invalidate cache
	Conditional        bool          `json:"conditional"` // Enable conditional requests
	Compress           bool          `json:"compress"`
	StaleWhileRevalidate time.Duration `json:"stale_while_revalidate"`
}

// ResponseWriter wraps gin.ResponseWriter to capture response data
type CachedResponseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

// NewCacheMiddleware creates a new cache middleware
func NewCacheMiddleware(
	cacheManager *observability.CacheManager,
	logger *slog.Logger,
	config *CacheMiddlewareConfig,
) *CacheMiddleware {
	if config == nil {
		config = DefaultCacheMiddlewareConfig()
	}
	
	return &CacheMiddleware{
		cacheManager: cacheManager,
		logger:       logger,
		config:      config,
	}
}

// DefaultCacheMiddlewareConfig returns default cache middleware configuration
func DefaultCacheMiddlewareConfig() *CacheMiddlewareConfig {
	return &CacheMiddlewareConfig{
		DefaultTTL:         15 * time.Minute,
		MaxCacheSize:       100 * 1024 * 1024, // 100MB
		EnableETags:        true,
		EnableLastModified: true,
		VaryHeaders:        []string{"Accept", "Accept-Encoding"},
		CacheableStatus:    []int{200, 301, 302, 304, 404, 410},
		CacheableMethods:   []string{"GET", "HEAD"},
		RouteStrategies: map[string]*CacheStrategy{
			"/api/v1/users":   {TTL: 5 * time.Minute, VaryBy: []string{"Accept"}},
			"/users":          {TTL: 10 * time.Minute, Conditional: true},
			"/health":         {TTL: 30 * time.Second},
			"/metrics":        {TTL: 1 * time.Minute},
		},
		SkipPatterns: []string{"/debug/", "/admin/"},
	}
}

// Handler returns the cache middleware handler
func (cm *CacheMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if caching should be skipped
		if cm.shouldSkipCache(c) {
			c.Next()
			return
		}
		
		// Only cache GET and HEAD requests
		if !cm.isCacheableMethod(c.Request.Method) {
			c.Next()
			return
		}
		
		// Generate cache key
		cacheKey := cm.generateCacheKey(c)
		
		// Try to serve from cache
		if cm.tryServeFromCache(c, cacheKey) {
			return
		}
		
		// Wrap response writer to capture response
		cachedWriter := &CachedResponseWriter{
			ResponseWriter: c.Writer,
			body:          bytes.NewBuffer(nil),
			statusCode:    200,
		}
		c.Writer = cachedWriter
		
		// Process request
		c.Next()
		
		// Cache response if appropriate
		cm.cacheResponse(c, cacheKey, cachedWriter)
	}
}

// shouldSkipCache determines if caching should be skipped for this request
func (cm *CacheMiddleware) shouldSkipCache(c *gin.Context) bool {
	path := c.Request.URL.Path
	
	// Check skip patterns
	for _, pattern := range cm.config.SkipPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	
	// Skip if Cache-Control: no-cache header is present
	if cacheControl := c.GetHeader("Cache-Control"); strings.Contains(cacheControl, "no-cache") {
		return true
	}
	
	// Skip if Authorization header is present (unless specifically configured)
	if c.GetHeader("Authorization") != "" {
		return true
	}
	
	return false
}

// isCacheableMethod checks if the HTTP method is cacheable
func (cm *CacheMiddleware) isCacheableMethod(method string) bool {
	for _, cacheable := range cm.config.CacheableMethods {
		if method == cacheable {
			return true
		}
	}
	return false
}

// generateCacheKey creates a unique cache key for the request
func (cm *CacheMiddleware) generateCacheKey(c *gin.Context) string {
	// Base key components
	components := []string{
		c.Request.Method,
		c.Request.URL.Path,
		c.Request.URL.RawQuery,
	}
	
	// Add strategy-specific vary headers
	strategy := cm.getCacheStrategy(c.Request.URL.Path)
	if strategy != nil {
		for _, header := range strategy.VaryBy {
			if value := c.GetHeader(header); value != "" {
				components = append(components, header+":"+value)
			}
		}
	}
	
	// Add default vary headers
	for _, header := range cm.config.VaryHeaders {
		if value := c.GetHeader(header); value != "" {
			components = append(components, header+":"+value)
		}
	}
	
	// Create hash of components
	key := strings.Join(components, "|")
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("http_cache:%x", hash)
}

// getCacheStrategy returns the cache strategy for a given path
func (cm *CacheMiddleware) getCacheStrategy(path string) *CacheStrategy {
	// Find most specific matching strategy
	var bestMatch *CacheStrategy
	maxMatchLength := 0
	
	for pattern, strategy := range cm.config.RouteStrategies {
		if strings.HasPrefix(path, pattern) && len(pattern) > maxMatchLength {
			bestMatch = strategy
			maxMatchLength = len(pattern)
		}
	}
	
	return bestMatch
}

// tryServeFromCache attempts to serve the response from cache
func (cm *CacheMiddleware) tryServeFromCache(c *gin.Context, cacheKey string) bool {
	cached, found := cm.cacheManager.Get(c.Request.Context(), cacheKey)
	if !found {
		return false
	}
	
	cachedResponse, ok := cached.(*CachedResponse)
	if !ok {
		cm.logger.Warn("Invalid cached response type", "key", cacheKey)
		return false
	}
	
	// Check if cached response is still valid
	if time.Since(cachedResponse.Timestamp) > cachedResponse.TTL {
		// Check if we can serve stale response
		strategy := cm.getCacheStrategy(c.Request.URL.Path)
		if strategy != nil && strategy.StaleWhileRevalidate > 0 {
			if time.Since(cachedResponse.Timestamp) < cachedResponse.TTL+strategy.StaleWhileRevalidate {
				cm.serveStaleResponse(c, cachedResponse)
				return true
			}
		}
		return false
	}
	
	// Handle conditional requests
	if cm.config.EnableETags && cachedResponse.ETag != "" {
		if ifNoneMatch := c.GetHeader("If-None-Match"); ifNoneMatch == cachedResponse.ETag {
			c.Writer.WriteHeader(http.StatusNotModified)
			return true
		}
	}
	
	if cm.config.EnableLastModified && !cachedResponse.LastModified.IsZero() {
		if ifModifiedSince := c.GetHeader("If-Modified-Since"); ifModifiedSince != "" {
			if t, err := http.ParseTime(ifModifiedSince); err == nil {
				if !cachedResponse.LastModified.After(t) {
					c.Writer.WriteHeader(http.StatusNotModified)
					return true
				}
			}
		}
	}
	
	// Serve cached response
	cm.serveCachedResponse(c, cachedResponse)
	return true
}

// cacheResponse stores the response in cache if appropriate
func (cm *CacheMiddleware) cacheResponse(c *gin.Context, cacheKey string, writer *CachedResponseWriter) {
	// Check if status code is cacheable
	if !cm.isCacheableStatus(writer.statusCode) {
		return
	}
	
	// Don't cache empty responses
	if writer.body.Len() == 0 {
		return
	}
	
	// Get cache strategy and TTL
	strategy := cm.getCacheStrategy(c.Request.URL.Path)
	ttl := cm.config.DefaultTTL
	if strategy != nil && strategy.TTL > 0 {
		ttl = strategy.TTL
	}
	
	// Create cached response
	cachedResponse := &CachedResponse{
		StatusCode:   writer.statusCode,
		Headers:      make(map[string]string),
		Body:         writer.body.Bytes(),
		Timestamp:    time.Now(),
		TTL:          ttl,
	}
	
	// Copy headers
	for key, values := range c.Writer.Header() {
		if len(values) > 0 {
			cachedResponse.Headers[key] = values[0]
		}
	}
	
	// Generate ETag if enabled
	if cm.config.EnableETags {
		etag := cm.generateETag(writer.body.Bytes())
		cachedResponse.ETag = etag
		c.Header("ETag", etag)
	}
	
	// Set Last-Modified if enabled
	if cm.config.EnableLastModified {
		lastModified := time.Now().UTC()
		cachedResponse.LastModified = lastModified
		c.Header("Last-Modified", lastModified.Format(http.TimeFormat))
	}
	
	// Set Cache-Control headers
	cm.setCacheControlHeaders(c, ttl, strategy)
	
	// Store in cache
	cm.cacheManager.Set(c.Request.Context(), cacheKey, cachedResponse, ttl)
	
	cm.logger.Debug("Response cached",
		"key", cacheKey,
		"ttl", ttl,
		"status", writer.statusCode,
		"size", writer.body.Len(),
	)
}

// isCacheableStatus checks if the status code is cacheable
func (cm *CacheMiddleware) isCacheableStatus(statusCode int) bool {
	for _, cacheable := range cm.config.CacheableStatus {
		if statusCode == cacheable {
			return true
		}
	}
	return false
}

// generateETag generates an ETag for the response body
func (cm *CacheMiddleware) generateETag(body []byte) string {
	hash := md5.Sum(body)
	return fmt.Sprintf(`"%x"`, hash)
}

// setCacheControlHeaders sets appropriate Cache-Control headers
func (cm *CacheMiddleware) setCacheControlHeaders(c *gin.Context, ttl time.Duration, strategy *CacheStrategy) {
	maxAge := int(ttl.Seconds())
	cacheControl := fmt.Sprintf("public, max-age=%d", maxAge)
	
	if strategy != nil && strategy.StaleWhileRevalidate > 0 {
		swr := int(strategy.StaleWhileRevalidate.Seconds())
		cacheControl += fmt.Sprintf(", stale-while-revalidate=%d", swr)
	}
	
	c.Header("Cache-Control", cacheControl)
}

// serveCachedResponse serves a cached response
func (cm *CacheMiddleware) serveCachedResponse(c *gin.Context, cached *CachedResponse) {
	// Set headers
	for key, value := range cached.Headers {
		c.Header(key, value)
	}
	
	// Set cache hit header
	c.Header("X-Cache", "HIT")
	
	// Write status and body
	c.Data(cached.StatusCode, "application/json", cached.Body)
}

// serveStaleResponse serves a stale response while revalidating
func (cm *CacheMiddleware) serveStaleResponse(c *gin.Context, cached *CachedResponse) {
	// Set headers
	for key, value := range cached.Headers {
		c.Header(key, value)
	}
	
	// Set stale cache header
	c.Header("X-Cache", "STALE")
	
	// Write status and body
	c.Data(cached.StatusCode, "application/json", cached.Body)
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode   int               `json:"status_code"`
	Headers      map[string]string `json:"headers"`
	Body         []byte            `json:"body"`
	Timestamp    time.Time         `json:"timestamp"`
	TTL          time.Duration     `json:"ttl"`
	ETag         string            `json:"etag"`
	LastModified time.Time         `json:"last_modified"`
}

// CachedResponseWriter methods

// Write captures the response body
func (w *CachedResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// WriteHeader captures the status code
func (w *CachedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// WriteString captures the response string
func (w *CachedResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// InvalidateCachePattern invalidates cache entries matching a pattern
func (cm *CacheMiddleware) InvalidateCachePattern(pattern string) {
	cm.cacheManager.InvalidatePattern(nil, pattern)
	cm.logger.Info("Cache pattern invalidated", "pattern", pattern)
}

// InvalidateRouteCache invalidates cache for specific route
func (cm *CacheMiddleware) InvalidateRouteCache(method, path string) {
	// Generate base cache key pattern
	pattern := fmt.Sprintf("http_cache:%s:%s", method, path)
	cm.InvalidateCachePattern(pattern)
}

// GetCacheMetrics returns cache performance metrics
func (cm *CacheMiddleware) GetCacheMetrics() *observability.CacheMetrics {
	return cm.cacheManager.GetMetrics()
}