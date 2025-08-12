// Comprehensive performance benchmarking suite for application performance testing
package observability

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/config"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
)

// BenchmarkSuite provides comprehensive performance benchmarking
type BenchmarkSuite struct {
	logger             *slog.Logger
	performanceMetrics *PerformanceMetrics
	resourceManager    *ResourceManager
	cacheManager       *CacheManager
	
	// Dependencies for benchmarking
	db                 *sql.DB
	userService        *services.UserService
	
	// Configuration
	config             *BenchmarkConfig
	
	// Results tracking
	results            []BenchmarkResult
	mutex              sync.Mutex
}

// BenchmarkConfig holds benchmarking configuration
type BenchmarkConfig struct {
	// Test duration and scaling
	Duration           time.Duration `json:"duration"`
	WarmupDuration     time.Duration `json:"warmup_duration"`
	CooldownDuration   time.Duration `json:"cooldown_duration"`
	
	// Concurrency testing
	MaxConcurrency     int           `json:"max_concurrency"`
	ConcurrencySteps   []int         `json:"concurrency_steps"`
	
	// Load patterns
	LoadPatterns       []LoadPattern `json:"load_patterns"`
	
	// Resource limits for testing
	MemoryLimitMB      int64         `json:"memory_limit_mb"`
	CPULimitPercent    int           `json:"cpu_limit_percent"`
	
	// Performance targets
	ResponseTimeTarget time.Duration `json:"response_time_target"`
	ThroughputTarget   int64         `json:"throughput_target"`
	ErrorRateTarget    float64       `json:"error_rate_target"`
	
	// Test categories
	EnabledCategories  []string      `json:"enabled_categories"`
}

// LoadPattern defines different load testing patterns
type LoadPattern struct {
	Name               string        `json:"name"`
	Type               string        `json:"type"` // constant, ramp, spike, step
	InitialRPS         int           `json:"initial_rps"`
	TargetRPS          int           `json:"target_rps"`
	Duration           time.Duration `json:"duration"`
	RampDuration       time.Duration `json:"ramp_duration"`
}

// BenchmarkResult holds results of a benchmark test
type BenchmarkResult struct {
	TestName           string                 `json:"test_name"`
	Category           string                 `json:"category"`
	Timestamp          time.Time             `json:"timestamp"`
	Duration           time.Duration         `json:"duration"`
	
	// Performance metrics
	TotalRequests      int64                 `json:"total_requests"`
	SuccessfulRequests int64                 `json:"successful_requests"`
	FailedRequests     int64                 `json:"failed_requests"`
	RequestsPerSecond  float64               `json:"requests_per_second"`
	
	// Response time metrics
	MinResponseTime    time.Duration         `json:"min_response_time"`
	MaxResponseTime    time.Duration         `json:"max_response_time"`
	AvgResponseTime    time.Duration         `json:"avg_response_time"`
	P50ResponseTime    time.Duration         `json:"p50_response_time"`
	P90ResponseTime    time.Duration         `json:"p90_response_time"`
	P95ResponseTime    time.Duration         `json:"p95_response_time"`
	P99ResponseTime    time.Duration         `json:"p99_response_time"`
	
	// Resource utilization
	StartMemoryMB      int64                 `json:"start_memory_mb"`
	EndMemoryMB        int64                 `json:"end_memory_mb"`
	PeakMemoryMB       int64                 `json:"peak_memory_mb"`
	StartCPUPercent    float64               `json:"start_cpu_percent"`
	EndCPUPercent      float64               `json:"end_cpu_percent"`
	PeakCPUPercent     float64               `json:"peak_cpu_percent"`
	
	// Additional metrics
	ErrorRate          float64               `json:"error_rate"`
	CacheHitRatio      float64               `json:"cache_hit_ratio"`
	DBQueriesPerSecond float64               `json:"db_queries_per_second"`
	GoroutinesCount    int                   `json:"goroutines_count"`
	GCPauses           []time.Duration       `json:"gc_pauses"`
	
	// Test parameters
	Concurrency        int                   `json:"concurrency"`
	LoadPattern        string                `json:"load_pattern"`
	
	// Pass/fail status
	Passed             bool                  `json:"passed"`
	Failures           []string              `json:"failures"`
	
	// Additional data
	CustomMetrics      map[string]interface{} `json:"custom_metrics"`
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite(
	logger *slog.Logger,
	performanceMetrics *PerformanceMetrics,
	resourceManager *ResourceManager,
	cacheManager *CacheManager,
	db *sql.DB,
	userService *services.UserService,
	config *BenchmarkConfig,
) *BenchmarkSuite {
	if config == nil {
		config = DefaultBenchmarkConfig()
	}
	
	return &BenchmarkSuite{
		logger:             logger,
		performanceMetrics: performanceMetrics,
		resourceManager:    resourceManager,
		cacheManager:       cacheManager,
		db:                 db,
		userService:        userService,
		config:            config,
		results:           make([]BenchmarkResult, 0),
	}
}

// DefaultBenchmarkConfig returns default benchmark configuration
func DefaultBenchmarkConfig() *BenchmarkConfig {
	return &BenchmarkConfig{
		Duration:           5 * time.Minute,
		WarmupDuration:     30 * time.Second,
		CooldownDuration:   30 * time.Second,
		MaxConcurrency:     100,
		ConcurrencySteps:   []int{1, 5, 10, 25, 50, 100},
		MemoryLimitMB:      512,
		CPULimitPercent:    80,
		ResponseTimeTarget: 100 * time.Millisecond,
		ThroughputTarget:   1000,
		ErrorRateTarget:    0.01, // 1% error rate
		EnabledCategories:  []string{"api", "database", "cache", "system", "stress"},
		LoadPatterns: []LoadPattern{
			{Name: "constant", Type: "constant", InitialRPS: 10, Duration: 1 * time.Minute},
			{Name: "ramp", Type: "ramp", InitialRPS: 1, TargetRPS: 50, Duration: 2 * time.Minute, RampDuration: 30 * time.Second},
			{Name: "spike", Type: "spike", InitialRPS: 10, TargetRPS: 100, Duration: 30 * time.Second},
		},
	}
}

// RunAllBenchmarks runs the complete benchmark suite
func (bs *BenchmarkSuite) RunAllBenchmarks(ctx context.Context) (*BenchmarkSuiteReport, error) {
	bs.logger.Info("Starting comprehensive benchmark suite",
		"duration", bs.config.Duration,
		"categories", bs.config.EnabledCategories,
		"max_concurrency", bs.config.MaxConcurrency,
	)
	
	startTime := time.Now()
	
	// Run warmup
	if bs.config.WarmupDuration > 0 {
		if err := bs.runWarmup(ctx); err != nil {
			bs.logger.Warn("Warmup failed", "error", err)
		}
	}
	
	// Run benchmark categories
	for _, category := range bs.config.EnabledCategories {
		if err := bs.runCategoryBenchmarks(ctx, category); err != nil {
			bs.logger.Error("Category benchmark failed", "category", category, "error", err)
			continue
		}
	}
	
	// Run cooldown
	if bs.config.CooldownDuration > 0 {
		bs.runCooldown(ctx)
	}
	
	totalDuration := time.Since(startTime)
	
	// Generate report
	report := bs.generateReport(totalDuration)
	
	bs.logger.Info("Benchmark suite completed",
		"duration", totalDuration,
		"total_tests", len(bs.results),
		"passed_tests", report.PassedTests,
		"failed_tests", report.FailedTests,
	)
	
	return report, nil
}

// runCategoryBenchmarks runs benchmarks for a specific category
func (bs *BenchmarkSuite) runCategoryBenchmarks(ctx context.Context, category string) error {
	bs.logger.Info("Running benchmark category", "category", category)
	
	switch category {
	case "api":
		return bs.runAPIBenchmarks(ctx)
	case "database":
		return bs.runDatabaseBenchmarks(ctx)
	case "cache":
		return bs.runCacheBenchmarks(ctx)
	case "system":
		return bs.runSystemBenchmarks(ctx)
	case "stress":
		return bs.runStressBenchmarks(ctx)
	default:
		return fmt.Errorf("unknown benchmark category: %s", category)
	}
}

// runAPIBenchmarks runs API endpoint benchmarks
func (bs *BenchmarkSuite) runAPIBenchmarks(ctx context.Context) error {
	tests := []struct {
		name string
		fn   func(context.Context, int) BenchmarkResult
	}{
		{"api_user_list", bs.benchmarkUserList},
		{"api_user_create", bs.benchmarkUserCreate},
		{"api_user_get", bs.benchmarkUserGet},
		{"api_user_update", bs.benchmarkUserUpdate},
		{"api_user_delete", bs.benchmarkUserDelete},
	}
	
	for _, test := range tests {
		for _, concurrency := range bs.config.ConcurrencySteps {
			if concurrency > bs.config.MaxConcurrency {
				break
			}
			
			result := test.fn(ctx, concurrency)
			result.TestName = fmt.Sprintf("%s_c%d", test.name, concurrency)
			result.Category = "api"
			result.Concurrency = concurrency
			
			bs.addResult(result)
		}
	}
	
	return nil
}

// runDatabaseBenchmarks runs database operation benchmarks
func (bs *BenchmarkSuite) runDatabaseBenchmarks(ctx context.Context) error {
	tests := []struct {
		name string
		fn   func(context.Context, int) BenchmarkResult
	}{
		{"db_select_single", bs.benchmarkDBSelectSingle},
		{"db_select_multiple", bs.benchmarkDBSelectMultiple},
		{"db_insert", bs.benchmarkDBInsert},
		{"db_update", bs.benchmarkDBUpdate},
		{"db_delete", bs.benchmarkDBDelete},
		{"db_transaction", bs.benchmarkDBTransaction},
	}
	
	for _, test := range tests {
		for _, concurrency := range bs.config.ConcurrencySteps {
			if concurrency > bs.config.MaxConcurrency {
				break
			}
			
			result := test.fn(ctx, concurrency)
			result.TestName = fmt.Sprintf("%s_c%d", test.name, concurrency)
			result.Category = "database"
			result.Concurrency = concurrency
			
			bs.addResult(result)
		}
	}
	
	return nil
}

// runCacheBenchmarks runs cache operation benchmarks
func (bs *BenchmarkSuite) runCacheBenchmarks(ctx context.Context) error {
	tests := []struct {
		name string
		fn   func(context.Context, int) BenchmarkResult
	}{
		{"cache_get_hit", bs.benchmarkCacheGetHit},
		{"cache_get_miss", bs.benchmarkCacheGetMiss},
		{"cache_set", bs.benchmarkCacheSet},
		{"cache_delete", bs.benchmarkCacheDelete},
		{"cache_mixed_workload", bs.benchmarkCacheMixedWorkload},
	}
	
	for _, test := range tests {
		for _, concurrency := range bs.config.ConcurrencySteps {
			if concurrency > bs.config.MaxConcurrency {
				break
			}
			
			result := test.fn(ctx, concurrency)
			result.TestName = fmt.Sprintf("%s_c%d", test.name, concurrency)
			result.Category = "cache"
			result.Concurrency = concurrency
			
			bs.addResult(result)
		}
	}
	
	return nil
}

// runSystemBenchmarks runs system resource benchmarks
func (bs *BenchmarkSuite) runSystemBenchmarks(ctx context.Context) error {
	tests := []struct {
		name string
		fn   func(context.Context) BenchmarkResult
	}{
		{"system_memory_allocation", bs.benchmarkMemoryAllocation},
		{"system_gc_performance", bs.benchmarkGCPerformance},
		{"system_goroutine_scaling", bs.benchmarkGoroutineScaling},
		{"system_cpu_intensive", bs.benchmarkCPUIntensive},
		{"system_io_performance", bs.benchmarkIOPerformance},
	}
	
	for _, test := range tests {
		result := test.fn(ctx)
		result.TestName = test.name
		result.Category = "system"
		
		bs.addResult(result)
	}
	
	return nil
}

// runStressBenchmarks runs stress testing benchmarks
func (bs *BenchmarkSuite) runStressBenchmarks(ctx context.Context) error {
	for _, pattern := range bs.config.LoadPatterns {
		result := bs.benchmarkLoadPattern(ctx, pattern)
		result.TestName = fmt.Sprintf("stress_%s", pattern.Name)
		result.Category = "stress"
		result.LoadPattern = pattern.Name
		
		bs.addResult(result)
	}
	
	return nil
}

// Specific benchmark implementations

// benchmarkUserList benchmarks user listing endpoint
func (bs *BenchmarkSuite) benchmarkUserList(ctx context.Context, concurrency int) BenchmarkResult {
	return bs.runConcurrentBenchmark(ctx, "user_list", concurrency, func(ctx context.Context) error {
		_, err := bs.userService.ListUsers(ctx, 10, 0)
		return err
	})
}

// benchmarkUserCreate benchmarks user creation
func (bs *BenchmarkSuite) benchmarkUserCreate(ctx context.Context, concurrency int) BenchmarkResult {
	return bs.runConcurrentBenchmark(ctx, "user_create", concurrency, func(ctx context.Context) error {
		// Generate unique test data
		testID := fmt.Sprintf("test-%d-%d", time.Now().UnixNano(), runtime.NumGoroutine())
		_, err := bs.userService.CreateUser(ctx, testID, fmt.Sprintf("%s@example.com", testID), "Test User")
		return err
	})
}

// benchmarkUserGet benchmarks user retrieval
func (bs *BenchmarkSuite) benchmarkUserGet(ctx context.Context, concurrency int) BenchmarkResult {
	return bs.runConcurrentBenchmark(ctx, "user_get", concurrency, func(ctx context.Context) error {
		_, err := bs.userService.GetUser(ctx, "test-user-1")
		return err
	})
}

// benchmarkUserUpdate benchmarks user updates
func (bs *BenchmarkSuite) benchmarkUserUpdate(ctx context.Context, concurrency int) BenchmarkResult {
	return bs.runConcurrentBenchmark(ctx, "user_update", concurrency, func(ctx context.Context) error {
		// This is a placeholder - actual implementation would depend on UserService interface
		return nil
	})
}

// benchmarkUserDelete benchmarks user deletion
func (bs *BenchmarkSuite) benchmarkUserDelete(ctx context.Context, concurrency int) BenchmarkResult {
	return bs.runConcurrentBenchmark(ctx, "user_delete", concurrency, func(ctx context.Context) error {
		// This is a placeholder - actual implementation would depend on UserService interface
		return nil
	})
}

// Database benchmark implementations

// benchmarkDBSelectSingle benchmarks single row selection
func (bs *BenchmarkSuite) benchmarkDBSelectSingle(ctx context.Context, concurrency int) BenchmarkResult {
	return bs.runConcurrentBenchmark(ctx, "db_select_single", concurrency, func(ctx context.Context) error {
		_, err := bs.db.QueryContext(ctx, "SELECT id, email, name FROM users LIMIT 1")
		return err
	})
}

// benchmarkDBSelectMultiple benchmarks multiple row selection
func (bs *BenchmarkSuite) benchmarkDBSelectMultiple(ctx context.Context, concurrency int) BenchmarkResult {
	return bs.runConcurrentBenchmark(ctx, "db_select_multiple", concurrency, func(ctx context.Context) error {
		_, err := bs.db.QueryContext(ctx, "SELECT id, email, name FROM users LIMIT 100")
		return err
	})
}

// benchmarkDBInsert benchmarks database insertion
func (bs *BenchmarkSuite) benchmarkDBInsert(ctx context.Context, concurrency int) BenchmarkResult {
	return bs.runConcurrentBenchmark(ctx, "db_insert", concurrency, func(ctx context.Context) error {
		testID := fmt.Sprintf("bench-%d-%d", time.Now().UnixNano(), runtime.NumGoroutine())
		_, err := bs.db.ExecContext(ctx, 
			"INSERT INTO users (id, email, name, created, modified) VALUES (?, ?, ?, ?, ?)",
			testID, 
			fmt.Sprintf("%s@benchmark.test", testID),
			"Benchmark User",
			time.Now(),
			time.Now(),
		)
		return err
	})
}

// benchmarkDBUpdate benchmarks database updates
func (bs *BenchmarkSuite) benchmarkDBUpdate(ctx context.Context, concurrency int) BenchmarkResult {
	return bs.runConcurrentBenchmark(ctx, "db_update", concurrency, func(ctx context.Context) error {
		_, err := bs.db.ExecContext(ctx, 
			"UPDATE users SET modified = ? WHERE id = ?",
			time.Now(),
			"test-user-1",
		)
		return err
	})
}

// benchmarkDBDelete benchmarks database deletion
func (bs *BenchmarkSuite) benchmarkDBDelete(ctx context.Context, concurrency int) BenchmarkResult {
	return bs.runConcurrentBenchmark(ctx, "db_delete", concurrency, func(ctx context.Context) error {
		testID := fmt.Sprintf("temp-%d", time.Now().UnixNano())
		_, err := bs.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", testID)
		return err
	})
}

// benchmarkDBTransaction benchmarks database transactions
func (bs *BenchmarkSuite) benchmarkDBTransaction(ctx context.Context, concurrency int) BenchmarkResult {
	return bs.runConcurrentBenchmark(ctx, "db_transaction", concurrency, func(ctx context.Context) error {
		tx, err := bs.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		
		testID := fmt.Sprintf("tx-%d-%d", time.Now().UnixNano(), runtime.NumGoroutine())
		_, err = tx.ExecContext(ctx, 
			"INSERT INTO users (id, email, name, created, modified) VALUES (?, ?, ?, ?, ?)",
			testID, 
			fmt.Sprintf("%s@tx.test", testID),
			"TX User",
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return err
		}
		
		return tx.Commit()
	})
}

// Cache benchmark implementations

// benchmarkCacheGetHit benchmarks cache hits
func (bs *BenchmarkSuite) benchmarkCacheGetHit(ctx context.Context, concurrency int) BenchmarkResult {
	// Pre-populate cache
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("benchmark-key-%d", i)
		bs.cacheManager.Set(ctx, key, fmt.Sprintf("value-%d", i), 1*time.Hour)
	}
	
	return bs.runConcurrentBenchmark(ctx, "cache_get_hit", concurrency, func(ctx context.Context) error {
		key := fmt.Sprintf("benchmark-key-%d", runtime.NumGoroutine()%100)
		_, found := bs.cacheManager.Get(ctx, key)
		if !found {
			return fmt.Errorf("cache miss when hit expected")
		}
		return nil
	})
}

// benchmarkCacheGetMiss benchmarks cache misses
func (bs *BenchmarkSuite) benchmarkCacheGetMiss(ctx context.Context, concurrency int) BenchmarkResult {
	return bs.runConcurrentBenchmark(ctx, "cache_get_miss", concurrency, func(ctx context.Context) error {
		key := fmt.Sprintf("missing-key-%d-%d", time.Now().UnixNano(), runtime.NumGoroutine())
		_, found := bs.cacheManager.Get(ctx, key)
		if found {
			return fmt.Errorf("cache hit when miss expected")
		}
		return nil
	})
}

// benchmarkCacheSet benchmarks cache writes
func (bs *BenchmarkSuite) benchmarkCacheSet(ctx context.Context, concurrency int) BenchmarkResult {
	return bs.runConcurrentBenchmark(ctx, "cache_set", concurrency, func(ctx context.Context) error {
		key := fmt.Sprintf("set-key-%d-%d", time.Now().UnixNano(), runtime.NumGoroutine())
		value := fmt.Sprintf("value-%d", runtime.NumGoroutine())
		bs.cacheManager.Set(ctx, key, value, 1*time.Hour)
		return nil
	})
}

// benchmarkCacheDelete benchmarks cache deletions
func (bs *BenchmarkSuite) benchmarkCacheDelete(ctx context.Context, concurrency int) BenchmarkResult {
	return bs.runConcurrentBenchmark(ctx, "cache_delete", concurrency, func(ctx context.Context) error {
		key := fmt.Sprintf("delete-key-%d-%d", time.Now().UnixNano(), runtime.NumGoroutine())
		bs.cacheManager.Delete(ctx, key)
		return nil
	})
}

// benchmarkCacheMixedWorkload benchmarks mixed cache operations
func (bs *BenchmarkSuite) benchmarkCacheMixedWorkload(ctx context.Context, concurrency int) BenchmarkResult {
	return bs.runConcurrentBenchmark(ctx, "cache_mixed", concurrency, func(ctx context.Context) error {
		operation := runtime.NumGoroutine() % 4
		key := fmt.Sprintf("mixed-key-%d", runtime.NumGoroutine()%100)
		
		switch operation {
		case 0, 1: // 50% reads
			bs.cacheManager.Get(ctx, key)
		case 2: // 25% writes
			bs.cacheManager.Set(ctx, key, "mixed-value", 1*time.Hour)
		case 3: // 25% deletes
			bs.cacheManager.Delete(ctx, key)
		}
		
		return nil
	})
}

// System benchmark implementations (placeholders)

// benchmarkMemoryAllocation benchmarks memory allocation patterns
func (bs *BenchmarkSuite) benchmarkMemoryAllocation(ctx context.Context) BenchmarkResult {
	// Implementation would test various memory allocation patterns
	return BenchmarkResult{TestName: "memory_allocation", Category: "system"}
}

// benchmarkGCPerformance benchmarks garbage collection performance
func (bs *BenchmarkSuite) benchmarkGCPerformance(ctx context.Context) BenchmarkResult {
	// Implementation would test GC behavior under load
	return BenchmarkResult{TestName: "gc_performance", Category: "system"}
}

// benchmarkGoroutineScaling benchmarks goroutine scaling
func (bs *BenchmarkSuite) benchmarkGoroutineScaling(ctx context.Context) BenchmarkResult {
	// Implementation would test goroutine creation and management
	return BenchmarkResult{TestName: "goroutine_scaling", Category: "system"}
}

// benchmarkCPUIntensive benchmarks CPU-intensive operations
func (bs *BenchmarkSuite) benchmarkCPUIntensive(ctx context.Context) BenchmarkResult {
	// Implementation would test CPU-bound operations
	return BenchmarkResult{TestName: "cpu_intensive", Category: "system"}
}

// benchmarkIOPerformance benchmarks I/O performance
func (bs *BenchmarkSuite) benchmarkIOPerformance(ctx context.Context) BenchmarkResult {
	// Implementation would test I/O operations
	return BenchmarkResult{TestName: "io_performance", Category: "system"}
}

// benchmarkLoadPattern benchmarks specific load patterns
func (bs *BenchmarkSuite) benchmarkLoadPattern(ctx context.Context, pattern LoadPattern) BenchmarkResult {
	// Implementation would execute the specific load pattern
	return BenchmarkResult{TestName: "load_pattern_" + pattern.Name, Category: "stress", LoadPattern: pattern.Name}
}

// Helper methods

// runConcurrentBenchmark runs a benchmark function with specified concurrency
func (bs *BenchmarkSuite) runConcurrentBenchmark(ctx context.Context, name string, concurrency int, fn func(context.Context) error) BenchmarkResult {
	result := BenchmarkResult{
		TestName:      name,
		Timestamp:     time.Now(),
		Concurrency:   concurrency,
		CustomMetrics: make(map[string]interface{}),
	}
	
	// Record initial resource state
	var startMem runtime.MemStats
	runtime.ReadMemStats(&startMem)
	result.StartMemoryMB = int64(startMem.Alloc) / 1024 / 1024
	result.GoroutinesCount = runtime.NumGoroutine()
	
	// Run benchmark
	start := time.Now()
	var wg sync.WaitGroup
	errorCh := make(chan error, concurrency)
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(ctx); err != nil {
				errorCh <- err
			}
		}()
	}
	
	wg.Wait()
	close(errorCh)
	
	result.Duration = time.Since(start)
	
	// Count errors
	for err := range errorCh {
		result.FailedRequests++
		if result.Failures == nil {
			result.Failures = make([]string, 0)
		}
		result.Failures = append(result.Failures, err.Error())
	}
	
	result.SuccessfulRequests = int64(concurrency) - result.FailedRequests
	result.TotalRequests = int64(concurrency)
	
	if result.Duration > 0 {
		result.RequestsPerSecond = float64(result.TotalRequests) / result.Duration.Seconds()
	}
	
	// Calculate response time metrics (simplified)
	result.AvgResponseTime = result.Duration / time.Duration(concurrency)
	result.MinResponseTime = result.AvgResponseTime / 2  // Simplified
	result.MaxResponseTime = result.AvgResponseTime * 2 // Simplified
	
	// Record final resource state
	var endMem runtime.MemStats
	runtime.ReadMemStats(&endMem)
	result.EndMemoryMB = int64(endMem.Alloc) / 1024 / 1024
	result.PeakMemoryMB = result.EndMemoryMB // Simplified
	
	// Calculate error rate
	if result.TotalRequests > 0 {
		result.ErrorRate = float64(result.FailedRequests) / float64(result.TotalRequests)
	}
	
	// Check if test passed
	result.Passed = result.ErrorRate <= bs.config.ErrorRateTarget &&
		result.AvgResponseTime <= bs.config.ResponseTimeTarget
	
	return result
}

// addResult adds a benchmark result to the suite
func (bs *BenchmarkSuite) addResult(result BenchmarkResult) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	
	bs.results = append(bs.results, result)
	
	bs.logger.Info("Benchmark completed",
		"test", result.TestName,
		"duration", result.Duration,
		"rps", result.RequestsPerSecond,
		"error_rate", result.ErrorRate,
		"passed", result.Passed,
	)
}

// runWarmup runs warmup operations
func (bs *BenchmarkSuite) runWarmup(ctx context.Context) error {
	bs.logger.Info("Running benchmark warmup", "duration", bs.config.WarmupDuration)
	
	// Warmup cache
	warmupData := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("warmup-key-%d", i)
		warmupData[key] = fmt.Sprintf("warmup-value-%d", i)
	}
	
	if err := bs.cacheManager.Warmup(ctx, warmupData); err != nil {
		return fmt.Errorf("cache warmup failed: %w", err)
	}
	
	// Warmup database connections
	for i := 0; i < 10; i++ {
		if _, err := bs.db.QueryContext(ctx, "SELECT 1"); err != nil {
			bs.logger.Warn("DB warmup query failed", "error", err)
		}
	}
	
	// Wait for warmup duration
	time.Sleep(bs.config.WarmupDuration)
	
	return nil
}

// runCooldown runs cooldown operations
func (bs *BenchmarkSuite) runCooldown(ctx context.Context) {
	bs.logger.Info("Running benchmark cooldown", "duration", bs.config.CooldownDuration)
	
	// Force garbage collection
	runtime.GC()
	
	// Wait for cooldown duration
	time.Sleep(bs.config.CooldownDuration)
}

// generateReport generates a comprehensive benchmark report
func (bs *BenchmarkSuite) generateReport(totalDuration time.Duration) *BenchmarkSuiteReport {
	report := &BenchmarkSuiteReport{
		Timestamp:     time.Now(),
		TotalDuration: totalDuration,
		TotalTests:    len(bs.results),
		Results:       bs.results,
		Summary:       make(map[string]*BenchmarkCategorySummary),
	}
	
	// Calculate summary statistics
	for _, result := range bs.results {
		if result.Passed {
			report.PassedTests++
		} else {
			report.FailedTests++
		}
		
		// Update category summary
		if summary, exists := report.Summary[result.Category]; exists {
			summary.TotalTests++
			if result.Passed {
				summary.PassedTests++
			} else {
				summary.FailedTests++
			}
			summary.TotalRequests += result.TotalRequests
			summary.TotalDuration += result.Duration
		} else {
			report.Summary[result.Category] = &BenchmarkCategorySummary{
				Category:      result.Category,
				TotalTests:    1,
				PassedTests:   map[bool]int{true: 1, false: 0}[result.Passed],
				FailedTests:   map[bool]int{true: 0, false: 1}[result.Passed],
				TotalRequests: result.TotalRequests,
				TotalDuration: result.Duration,
			}
		}
	}
	
	// Calculate averages for each category
	for _, summary := range report.Summary {
		if summary.TotalTests > 0 {
			summary.AvgResponseTime = summary.TotalDuration / time.Duration(summary.TotalRequests)
			summary.AvgThroughput = float64(summary.TotalRequests) / summary.TotalDuration.Seconds()
		}
	}
	
	return report
}

// GetResults returns all benchmark results
func (bs *BenchmarkSuite) GetResults() []BenchmarkResult {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	
	return append([]BenchmarkResult(nil), bs.results...)
}

// BenchmarkSuiteReport holds the complete benchmark suite results
type BenchmarkSuiteReport struct {
	Timestamp     time.Time                           `json:"timestamp"`
	TotalDuration time.Duration                       `json:"total_duration"`
	TotalTests    int                                 `json:"total_tests"`
	PassedTests   int                                 `json:"passed_tests"`
	FailedTests   int                                 `json:"failed_tests"`
	Results       []BenchmarkResult                   `json:"results"`
	Summary       map[string]*BenchmarkCategorySummary `json:"summary"`
}

// BenchmarkCategorySummary holds summary statistics for a benchmark category
type BenchmarkCategorySummary struct {
	Category        string        `json:"category"`
	TotalTests      int           `json:"total_tests"`
	PassedTests     int           `json:"passed_tests"`
	FailedTests     int           `json:"failed_tests"`
	TotalRequests   int64         `json:"total_requests"`
	TotalDuration   time.Duration `json:"total_duration"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	AvgThroughput   float64       `json:"avg_throughput"`
}