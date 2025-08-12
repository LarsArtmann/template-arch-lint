// Multi-tier caching system with intelligent invalidation and performance monitoring
package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// CacheManager provides intelligent multi-tier caching with performance monitoring
type CacheManager struct {
	logger             *slog.Logger
	performanceMetrics *PerformanceMetrics
	
	// Cache tiers
	l1Cache            *InMemoryCache  // L1: In-memory cache
	l2Cache            *RedisCache     // L2: Redis cache (optional)
	
	// Configuration
	config             *CacheConfig
	
	// Metrics and monitoring
	metrics            *CacheMetrics
	mutex              sync.RWMutex
}

// CacheConfig holds caching configuration
type CacheConfig struct {
	// L1 Cache (In-Memory)
	L1Enabled           bool          `json:"l1_enabled"`
	L1MaxSize           int64         `json:"l1_max_size"`
	L1TTL               time.Duration `json:"l1_ttl"`
	L1EvictionPolicy    string        `json:"l1_eviction_policy"` // LRU, LFU, FIFO
	
	// L2 Cache (Redis)
	L2Enabled           bool          `json:"l2_enabled"`
	L2RedisURL          string        `json:"l2_redis_url"`
	L2TTL               time.Duration `json:"l2_ttl"`
	L2MaxMemory         string        `json:"l2_max_memory"`
	L2ClusterMode       bool          `json:"l2_cluster_mode"`
	
	// Performance and monitoring
	MonitoringInterval  time.Duration `json:"monitoring_interval"`
	HitRatioThreshold   float64       `json:"hit_ratio_threshold"`
	EvictionThreshold   float64       `json:"eviction_threshold"`
	WarmupEnabled       bool          `json:"warmup_enabled"`
	
	// Hot path optimization
	HotPathDetection    bool          `json:"hot_path_detection"`
	HotPathThreshold    int64         `json:"hot_path_threshold"`
	HotPathCacheTTL     time.Duration `json:"hot_path_cache_ttl"`
}

// CacheMetrics holds cache performance metrics
type CacheMetrics struct {
	Timestamp           time.Time `json:"timestamp"`
	
	// L1 Cache metrics
	L1Hits              int64     `json:"l1_hits"`
	L1Misses            int64     `json:"l1_misses"`
	L1Size              int64     `json:"l1_size"`
	L1Evictions         int64     `json:"l1_evictions"`
	L1HitRatio          float64   `json:"l1_hit_ratio"`
	
	// L2 Cache metrics
	L2Hits              int64     `json:"l2_hits"`
	L2Misses            int64     `json:"l2_misses"`
	L2Size              int64     `json:"l2_size"`
	L2Evictions         int64     `json:"l2_evictions"`
	L2HitRatio          float64   `json:"l2_hit_ratio"`
	
	// Overall metrics
	TotalHits           int64     `json:"total_hits"`
	TotalMisses         int64     `json:"total_misses"`
	OverallHitRatio     float64   `json:"overall_hit_ratio"`
	
	// Hot path metrics
	HotPaths            []HotPath `json:"hot_paths"`
}

// HotPath represents a detected hot path
type HotPath struct {
	Key         string    `json:"key"`
	AccessCount int64     `json:"access_count"`
	LastAccess  time.Time `json:"last_access"`
	CacheTier   string    `json:"cache_tier"`
}

// CacheItem represents an item in the cache
type CacheItem struct {
	Key        string      `json:"key"`
	Value      interface{} `json:"value"`
	Timestamp  time.Time   `json:"timestamp"`
	TTL        time.Duration `json:"ttl"`
	AccessCount int64      `json:"access_count"`
	LastAccess time.Time   `json:"last_access"`
	Size       int64       `json:"size"`
}

// InMemoryCache represents the L1 in-memory cache
type InMemoryCache struct {
	items          map[string]*CacheItem
	mutex          sync.RWMutex
	maxSize        int64
	currentSize    int64
	hits           int64
	misses         int64
	evictions      int64
	evictionPolicy string
}

// RedisCache represents the L2 Redis cache (placeholder for Redis implementation)
type RedisCache struct {
	enabled    bool
	hits       int64
	misses     int64
	evictions  int64
	// Redis client would go here
}

// NewCacheManager creates a new cache manager
func NewCacheManager(logger *slog.Logger, performanceMetrics *PerformanceMetrics, config *CacheConfig) *CacheManager {
	if config == nil {
		config = DefaultCacheConfig()
	}
	
	cm := &CacheManager{
		logger:             logger,
		performanceMetrics: performanceMetrics,
		config:            config,
		metrics:           &CacheMetrics{},
	}
	
	// Initialize L1 cache
	if config.L1Enabled {
		cm.l1Cache = &InMemoryCache{
			items:          make(map[string]*CacheItem),
			maxSize:        config.L1MaxSize,
			evictionPolicy: config.L1EvictionPolicy,
		}
	}
	
	// Initialize L2 cache
	if config.L2Enabled {
		cm.l2Cache = &RedisCache{
			enabled: true,
		}
		// Initialize Redis client here in production
	}
	
	return cm
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		L1Enabled:           true,
		L1MaxSize:           100 * 1024 * 1024, // 100MB
		L1TTL:               30 * time.Minute,
		L1EvictionPolicy:    "LRU",
		L2Enabled:           false,
		L2TTL:               2 * time.Hour,
		MonitoringInterval:  60 * time.Second,
		HitRatioThreshold:   0.8,
		EvictionThreshold:   0.9,
		WarmupEnabled:       true,
		HotPathDetection:    true,
		HotPathThreshold:    100,
		HotPathCacheTTL:     1 * time.Hour,
	}
}

// StartMonitoring starts cache monitoring and optimization
func (cm *CacheManager) StartMonitoring(ctx context.Context) {
	ticker := time.NewTicker(cm.config.MonitoringInterval)
	defer ticker.Stop()
	
	cm.logger.Info("Cache monitoring started",
		"interval", cm.config.MonitoringInterval,
		"l1_enabled", cm.config.L1Enabled,
		"l2_enabled", cm.config.L2Enabled,
	)
	
	for {
		select {
		case <-ctx.Done():
			cm.logger.Info("Cache monitoring stopped")
			return
		case <-ticker.C:
			cm.collectMetrics()
			cm.optimizeIfNeeded()
		}
	}
}

// Get retrieves a value from the cache with tier fallback
func (cm *CacheManager) Get(ctx context.Context, key string) (interface{}, bool) {
	// Try L1 cache first
	if cm.l1Cache != nil {
		if value, found := cm.l1Cache.Get(key); found {
			cm.recordCacheHit(ctx, "l1", key)
			return value, true
		}
		cm.recordCacheMiss(ctx, "l1", key)
	}
	
	// Try L2 cache if L1 miss
	if cm.l2Cache != nil && cm.l2Cache.enabled {
		if value, found := cm.l2Cache.Get(key); found {
			cm.recordCacheHit(ctx, "l2", key)
			
			// Promote to L1 cache
			if cm.l1Cache != nil {
				cm.l1Cache.Set(key, value, cm.config.L1TTL)
			}
			
			return value, true
		}
		cm.recordCacheMiss(ctx, "l2", key)
	}
	
	return nil, false
}

// Set stores a value in appropriate cache tiers
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	// Store in L1 cache
	if cm.l1Cache != nil {
		cm.l1Cache.Set(key, value, ttl)
	}
	
	// Store in L2 cache
	if cm.l2Cache != nil && cm.l2Cache.enabled {
		cm.l2Cache.Set(key, value, ttl)
	}
	
	// Detect hot paths
	if cm.config.HotPathDetection {
		cm.detectHotPath(key)
	}
	
	cm.logger.Debug("Cache set", "key", key, "ttl", ttl)
}

// Delete removes a value from all cache tiers
func (cm *CacheManager) Delete(ctx context.Context, key string) {
	if cm.l1Cache != nil {
		cm.l1Cache.Delete(key)
	}
	
	if cm.l2Cache != nil && cm.l2Cache.enabled {
		cm.l2Cache.Delete(key)
	}
	
	cm.logger.Debug("Cache delete", "key", key)
}

// InvalidatePattern invalidates cache entries matching a pattern
func (cm *CacheManager) InvalidatePattern(ctx context.Context, pattern string) {
	// Implement pattern-based invalidation
	if cm.l1Cache != nil {
		cm.l1Cache.InvalidatePattern(pattern)
	}
	
	if cm.l2Cache != nil && cm.l2Cache.enabled {
		cm.l2Cache.InvalidatePattern(pattern)
	}
	
	cm.logger.Info("Cache pattern invalidation", "pattern", pattern)
}

// Warmup pre-loads frequently accessed data into cache
func (cm *CacheManager) Warmup(ctx context.Context, warmupData map[string]interface{}) error {
	if !cm.config.WarmupEnabled {
		return nil
	}
	
	cm.logger.Info("Cache warmup started", "items", len(warmupData))
	
	for key, value := range warmupData {
		cm.Set(ctx, key, value, cm.config.HotPathCacheTTL)
	}
	
	cm.logger.Info("Cache warmup completed", "items", len(warmupData))
	return nil
}

// collectMetrics collects cache performance metrics
func (cm *CacheManager) collectMetrics() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	cm.metrics.Timestamp = time.Now()
	
	// Collect L1 metrics
	if cm.l1Cache != nil {
		cm.metrics.L1Hits = atomic.LoadInt64(&cm.l1Cache.hits)
		cm.metrics.L1Misses = atomic.LoadInt64(&cm.l1Cache.misses)
		cm.metrics.L1Size = atomic.LoadInt64(&cm.l1Cache.currentSize)
		cm.metrics.L1Evictions = atomic.LoadInt64(&cm.l1Cache.evictions)
		
		total := cm.metrics.L1Hits + cm.metrics.L1Misses
		if total > 0 {
			cm.metrics.L1HitRatio = float64(cm.metrics.L1Hits) / float64(total)
		}
	}
	
	// Collect L2 metrics
	if cm.l2Cache != nil && cm.l2Cache.enabled {
		cm.metrics.L2Hits = atomic.LoadInt64(&cm.l2Cache.hits)
		cm.metrics.L2Misses = atomic.LoadInt64(&cm.l2Cache.misses)
		cm.metrics.L2Evictions = atomic.LoadInt64(&cm.l2Cache.evictions)
		
		total := cm.metrics.L2Hits + cm.metrics.L2Misses
		if total > 0 {
			cm.metrics.L2HitRatio = float64(cm.metrics.L2Hits) / float64(total)
		}
	}
	
	// Calculate overall metrics
	cm.metrics.TotalHits = cm.metrics.L1Hits + cm.metrics.L2Hits
	cm.metrics.TotalMisses = cm.metrics.L1Misses + cm.metrics.L2Misses
	
	totalOperations := cm.metrics.TotalHits + cm.metrics.TotalMisses
	if totalOperations > 0 {
		cm.metrics.OverallHitRatio = float64(cm.metrics.TotalHits) / float64(totalOperations)
	}
	
	// Record metrics for observability
	cm.recordMetrics()
}

// recordMetrics records cache metrics for observability
func (cm *CacheManager) recordMetrics() {
	ctx := context.Background()
	
	// Record cache hits and misses
	if cm.metrics.L1Hits > 0 {
		cm.performanceMetrics.RecordCacheOperation(ctx, "l1_get", true)
	}
	
	if cm.metrics.L1Misses > 0 {
		cm.performanceMetrics.RecordCacheOperation(ctx, "l1_get", false)
	}
	
	cm.logger.Debug("Cache metrics collected",
		"l1_hit_ratio", cm.metrics.L1HitRatio,
		"l2_hit_ratio", cm.metrics.L2HitRatio,
		"overall_hit_ratio", cm.metrics.OverallHitRatio,
		"l1_size_mb", cm.metrics.L1Size/1024/1024,
		"l1_evictions", cm.metrics.L1Evictions,
	)
}

// optimizeIfNeeded performs cache optimization if needed
func (cm *CacheManager) optimizeIfNeeded() {
	// Check hit ratio and optimize
	if cm.metrics.OverallHitRatio < cm.config.HitRatioThreshold {
		cm.optimizeHitRatio()
	}
	
	// Check eviction rate and optimize
	if cm.l1Cache != nil {
		evictionRate := float64(cm.metrics.L1Evictions) / float64(cm.metrics.L1Size+1)
		if evictionRate > cm.config.EvictionThreshold {
			cm.optimizeEvictionRate()
		}
	}
}

// optimizeHitRatio optimizes cache configuration to improve hit ratio
func (cm *CacheManager) optimizeHitRatio() {
	cm.logger.Info("Optimizing cache for better hit ratio",
		"current_hit_ratio", cm.metrics.OverallHitRatio,
		"threshold", cm.config.HitRatioThreshold,
	)
	
	// Increase L1 cache TTL
	if cm.config.L1TTL < 2*time.Hour {
		cm.config.L1TTL = cm.config.L1TTL * 2
		cm.logger.Info("Increased L1 cache TTL", "new_ttl", cm.config.L1TTL)
	}
	
	// Increase L2 cache TTL
	if cm.config.L2TTL < 8*time.Hour {
		cm.config.L2TTL = cm.config.L2TTL * 2
		cm.logger.Info("Increased L2 cache TTL", "new_ttl", cm.config.L2TTL)
	}
}

// optimizeEvictionRate optimizes cache configuration to reduce evictions
func (cm *CacheManager) optimizeEvictionRate() {
	cm.logger.Info("Optimizing cache to reduce eviction rate",
		"evictions", cm.metrics.L1Evictions,
		"cache_size", cm.metrics.L1Size,
	)
	
	// Increase L1 cache size if possible
	if cm.config.L1MaxSize < 500*1024*1024 { // Max 500MB
		cm.config.L1MaxSize = cm.config.L1MaxSize * 2
		if cm.l1Cache != nil {
			cm.l1Cache.maxSize = cm.config.L1MaxSize
		}
		cm.logger.Info("Increased L1 cache size", "new_size_mb", cm.config.L1MaxSize/1024/1024)
	}
}

// detectHotPath detects frequently accessed cache keys
func (cm *CacheManager) detectHotPath(key string) {
	// Implementation for hot path detection would go here
	// This is a simplified placeholder
}

// recordCacheHit records a cache hit for metrics
func (cm *CacheManager) recordCacheHit(ctx context.Context, tier, key string) {
	cm.performanceMetrics.RecordCacheOperation(ctx, tier+"_get", true)
}

// recordCacheMiss records a cache miss for metrics
func (cm *CacheManager) recordCacheMiss(ctx context.Context, tier, key string) {
	cm.performanceMetrics.RecordCacheOperation(ctx, tier+"_get", false)
}

// GetMetrics returns current cache metrics
func (cm *CacheManager) GetMetrics() *CacheMetrics {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	metrics := *cm.metrics
	return &metrics
}

// GetHealthStatus returns cache health status
func (cm *CacheManager) GetHealthStatus() map[string]interface{} {
	metrics := cm.GetMetrics()
	
	healthStatus := map[string]interface{}{
		"healthy": true,
		"metrics": metrics,
		"issues":  []string{},
	}
	
	var issues []string
	
	// Check hit ratio
	if metrics.OverallHitRatio < cm.config.HitRatioThreshold {
		issues = append(issues, fmt.Sprintf("Low cache hit ratio: %.2f%%", metrics.OverallHitRatio*100))
	}
	
	// Check eviction rate
	if cm.l1Cache != nil && metrics.L1Size > 0 {
		evictionRate := float64(metrics.L1Evictions) / float64(metrics.L1Size) * 100
		if evictionRate > cm.config.EvictionThreshold*100 {
			issues = append(issues, fmt.Sprintf("High eviction rate: %.2f%%", evictionRate))
		}
	}
	
	if len(issues) > 0 {
		healthStatus["healthy"] = false
	}
	
	healthStatus["issues"] = issues
	return healthStatus
}

// InMemoryCache methods

// Get retrieves a value from L1 cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	item, exists := c.items[key]
	if !exists {
		atomic.AddInt64(&c.misses, 1)
		return nil, false
	}
	
	// Check TTL
	if time.Since(item.Timestamp) > item.TTL {
		atomic.AddInt64(&c.misses, 1)
		// Remove expired item
		go func() {
			c.mutex.Lock()
			delete(c.items, key)
			c.currentSize -= item.Size
			c.mutex.Unlock()
		}()
		return nil, false
	}
	
	// Update access metrics
	atomic.AddInt64(&item.AccessCount, 1)
	item.LastAccess = time.Now()
	atomic.AddInt64(&c.hits, 1)
	
	return item.Value, true
}

// Set stores a value in L1 cache
func (c *InMemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Calculate item size (simplified)
	valueBytes, _ := json.Marshal(value)
	itemSize := int64(len(key) + len(valueBytes) + 64) // Approximate overhead
	
	// Check if we need to evict items
	for c.currentSize+itemSize > c.maxSize && len(c.items) > 0 {
		c.evictOne()
	}
	
	item := &CacheItem{
		Key:        key,
		Value:      value,
		Timestamp:  time.Now(),
		TTL:        ttl,
		AccessCount: 0,
		LastAccess: time.Now(),
		Size:       itemSize,
	}
	
	// Remove existing item if present
	if existing, exists := c.items[key]; exists {
		c.currentSize -= existing.Size
	}
	
	c.items[key] = item
	c.currentSize += itemSize
}

// Delete removes an item from L1 cache
func (c *InMemoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if item, exists := c.items[key]; exists {
		delete(c.items, key)
		c.currentSize -= item.Size
	}
}

// InvalidatePattern removes items matching a pattern
func (c *InMemoryCache) InvalidatePattern(pattern string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Simple pattern matching (would use regex in production)
	for key, item := range c.items {
		if key == pattern { // Simplified matching
			delete(c.items, key)
			c.currentSize -= item.Size
		}
	}
}

// evictOne evicts one item based on eviction policy
func (c *InMemoryCache) evictOne() {
	if len(c.items) == 0 {
		return
	}
	
	var keyToEvict string
	var oldestTime time.Time = time.Now()
	
	// Simple LRU eviction
	for key, item := range c.items {
		if item.LastAccess.Before(oldestTime) {
			oldestTime = item.LastAccess
			keyToEvict = key
		}
	}
	
	if keyToEvict != "" {
		item := c.items[keyToEvict]
		delete(c.items, keyToEvict)
		c.currentSize -= item.Size
		atomic.AddInt64(&c.evictions, 1)
	}
}

// RedisCache methods (placeholders for Redis implementation)

// Get retrieves a value from L2 cache
func (c *RedisCache) Get(key string) (interface{}, bool) {
	if !c.enabled {
		return nil, false
	}
	// Redis implementation would go here
	atomic.AddInt64(&c.misses, 1)
	return nil, false
}

// Set stores a value in L2 cache
func (c *RedisCache) Set(key string, value interface{}, ttl time.Duration) {
	if !c.enabled {
		return
	}
	// Redis implementation would go here
}

// Delete removes a value from L2 cache
func (c *RedisCache) Delete(key string) {
	if !c.enabled {
		return
	}
	// Redis implementation would go here
}

// InvalidatePattern removes items matching a pattern from L2 cache
func (c *RedisCache) InvalidatePattern(pattern string) {
	if !c.enabled {
		return
	}
	// Redis implementation would go here
}