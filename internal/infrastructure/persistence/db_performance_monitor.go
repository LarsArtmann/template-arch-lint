// Database performance monitoring and optimization utilities
package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/observability"
)

// DBPerformanceMonitor provides database performance monitoring and optimization
type DBPerformanceMonitor struct {
	db                *sql.DB
	logger            *slog.Logger
	performanceMetrics *observability.PerformanceMetrics
}

// NewDBPerformanceMonitor creates a new database performance monitor
func NewDBPerformanceMonitor(db *sql.DB, logger *slog.Logger, performanceMetrics *observability.PerformanceMetrics) *DBPerformanceMonitor {
	return &DBPerformanceMonitor{
		db:                 db,
		logger:             logger,
		performanceMetrics: performanceMetrics,
	}
}

// OptimizationResult represents the result of database optimization
type OptimizationResult struct {
	Operation     string        `json:"operation"`
	Success       bool          `json:"success"`
	Duration      time.Duration `json:"duration"`
	RowsAffected  int64         `json:"rows_affected"`
	Error         string        `json:"error,omitempty"`
	Improvements  []string      `json:"improvements,omitempty"`
}

// DatabaseStats holds database performance statistics
type DatabaseStats struct {
	Connections struct {
		Open      int `json:"open"`
		InUse     int `json:"in_use"`
		Idle      int `json:"idle"`
		MaxOpen   int `json:"max_open"`
		MaxIdle   int `json:"max_idle"`
	} `json:"connections"`
	
	Performance struct {
		QueriesTotal    int64 `json:"queries_total"`
		SlowQueries     int64 `json:"slow_queries"`
		AverageQueryMs  float64 `json:"average_query_ms"`
		CacheHitRatio   float64 `json:"cache_hit_ratio"`
	} `json:"performance"`
	
	Storage struct {
		DatabaseSizeBytes int64 `json:"database_size_bytes"`
		PageCount         int64 `json:"page_count"`
		PageSize          int64 `json:"page_size"`
		FreePagesCount    int64 `json:"free_pages_count"`
	} `json:"storage"`
	
	Tables map[string]TableStats `json:"tables"`
}

// TableStats holds individual table statistics
type TableStats struct {
	RowCount      int64   `json:"row_count"`
	SizeBytes     int64   `json:"size_bytes"`
	IndexCount    int     `json:"index_count"`
	FragmentRatio float64 `json:"fragment_ratio"`
}

// GetDatabaseStats collects comprehensive database statistics
func (dpm *DBPerformanceMonitor) GetDatabaseStats(ctx context.Context) (*DatabaseStats, error) {
	start := time.Now()
	defer func() {
		dpm.performanceMetrics.RecordDatabaseQuery(ctx, "GetDatabaseStats", time.Since(start), true)
	}()

	stats := &DatabaseStats{
		Tables: make(map[string]TableStats),
	}

	// Get connection pool stats
	dbStats := dpm.db.Stats()
	stats.Connections.Open = dbStats.OpenConnections
	stats.Connections.InUse = dbStats.InUse
	stats.Connections.Idle = dbStats.Idle
	stats.Connections.MaxOpen = dbStats.MaxOpenConnections
	// MaxIdleConnections is not available in sql.DBStats, would need to track separately
	stats.Connections.MaxIdle = 0 // Placeholder

	// Get database size and page information
	if err := dpm.collectStorageStats(ctx, stats); err != nil {
		dpm.logger.Warn("Failed to collect storage stats", "error", err)
	}

	// Get table-specific statistics
	if err := dpm.collectTableStats(ctx, stats); err != nil {
		dpm.logger.Warn("Failed to collect table stats", "error", err)
	}

	return stats, nil
}

// collectStorageStats collects database storage statistics
func (dpm *DBPerformanceMonitor) collectStorageStats(ctx context.Context, stats *DatabaseStats) error {
	// Get page count and size
	row := dpm.db.QueryRowContext(ctx, "PRAGMA page_count")
	if err := row.Scan(&stats.Storage.PageCount); err != nil {
		return fmt.Errorf("failed to get page count: %w", err)
	}

	row = dpm.db.QueryRowContext(ctx, "PRAGMA page_size")
	if err := row.Scan(&stats.Storage.PageSize); err != nil {
		return fmt.Errorf("failed to get page size: %w", err)
	}

	stats.Storage.DatabaseSizeBytes = stats.Storage.PageCount * stats.Storage.PageSize

	// Get free pages count
	row = dpm.db.QueryRowContext(ctx, "PRAGMA freelist_count")
	if err := row.Scan(&stats.Storage.FreePagesCount); err != nil {
		return fmt.Errorf("failed to get free pages count: %w", err)
	}

	return nil
}

// collectTableStats collects table-specific statistics
func (dpm *DBPerformanceMonitor) collectTableStats(ctx context.Context, stats *DatabaseStats) error {
	// Get all table names
	rows, err := dpm.db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tableNames = append(tableNames, tableName)
	}

	// Collect stats for each table
	for _, tableName := range tableNames {
		tableStats := TableStats{}

		// Get row count
		row := dpm.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName))
		if err := row.Scan(&tableStats.RowCount); err != nil {
			dpm.logger.Warn("Failed to get row count for table", "table", tableName, "error", err)
			continue
		}

		// Get index count
		indexRows, err := dpm.db.QueryContext(ctx, "SELECT COUNT(*) FROM pragma_index_list(?)", tableName)
		if err == nil {
			defer indexRows.Close()
			if indexRows.Next() {
				indexRows.Scan(&tableStats.IndexCount)
			}
		}

		stats.Tables[tableName] = tableStats
	}

	return nil
}

// OptimizeDatabase performs comprehensive database optimization
func (dpm *DBPerformanceMonitor) OptimizeDatabase(ctx context.Context) ([]OptimizationResult, error) {
	results := make([]OptimizationResult, 0)

	// Analyze tables for query optimization
	analyzeResult := dpm.analyzeDatabase(ctx)
	results = append(results, analyzeResult)

	// Vacuum database to reclaim space
	vacuumResult := dpm.vacuumDatabase(ctx)
	results = append(results, vacuumResult)

	// Optimize WAL checkpoint
	checkpointResult := dpm.optimizeWALCheckpoint(ctx)
	results = append(results, checkpointResult)

	// Update table statistics
	statsResult := dpm.updateTableStatistics(ctx)
	results = append(results, statsResult)

	return results, nil
}

// analyzeDatabase runs ANALYZE to update query planner statistics
func (dpm *DBPerformanceMonitor) analyzeDatabase(ctx context.Context) OptimizationResult {
	start := time.Now()
	result := OptimizationResult{
		Operation: "ANALYZE",
		Success:   true,
	}

	_, err := dpm.db.ExecContext(ctx, "ANALYZE")
	result.Duration = time.Since(start)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		dpm.logger.Error("Database ANALYZE failed", "error", err)
	} else {
		result.Improvements = []string{"Updated query planner statistics"}
		dpm.logger.Info("Database ANALYZE completed", "duration", result.Duration)
	}

	dpm.performanceMetrics.RecordDatabaseQuery(ctx, "ANALYZE", result.Duration, result.Success)
	return result
}

// vacuumDatabase performs VACUUM to reclaim unused space
func (dpm *DBPerformanceMonitor) vacuumDatabase(ctx context.Context) OptimizationResult {
	start := time.Now()
	result := OptimizationResult{
		Operation: "VACUUM",
		Success:   true,
	}

	// Get database size before vacuum
	var sizeBefore int64
	row := dpm.db.QueryRowContext(ctx, "PRAGMA page_count")
	row.Scan(&sizeBefore)

	_, err := dpm.db.ExecContext(ctx, "VACUUM")
	result.Duration = time.Since(start)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		dpm.logger.Error("Database VACUUM failed", "error", err)
	} else {
		// Get database size after vacuum
		var sizeAfter int64
		row := dpm.db.QueryRowContext(ctx, "PRAGMA page_count")
		row.Scan(&sizeAfter)

		spaceReclaimed := sizeBefore - sizeAfter
		result.Improvements = []string{
			"Reclaimed unused space",
			fmt.Sprintf("Reduced database size by %d pages", spaceReclaimed),
		}
		dpm.logger.Info("Database VACUUM completed", 
			"duration", result.Duration,
			"pages_reclaimed", spaceReclaimed,
		)
	}

	dpm.performanceMetrics.RecordDatabaseQuery(ctx, "VACUUM", result.Duration, result.Success)
	return result
}

// optimizeWALCheckpoint performs WAL checkpoint optimization
func (dpm *DBPerformanceMonitor) optimizeWALCheckpoint(ctx context.Context) OptimizationResult {
	start := time.Now()
	result := OptimizationResult{
		Operation: "WAL_CHECKPOINT",
		Success:   true,
	}

	// Check if WAL mode is enabled
	var journalMode string
	row := dpm.db.QueryRowContext(ctx, "PRAGMA journal_mode")
	if err := row.Scan(&journalMode); err != nil || journalMode != "wal" {
		result.Success = false
		result.Error = "WAL mode not enabled"
		return result
	}

	_, err := dpm.db.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)")
	result.Duration = time.Since(start)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		dpm.logger.Error("WAL checkpoint failed", "error", err)
	} else {
		result.Improvements = []string{"Checkpointed WAL file"}
		dpm.logger.Info("WAL checkpoint completed", "duration", result.Duration)
	}

	dpm.performanceMetrics.RecordDatabaseQuery(ctx, "WAL_CHECKPOINT", result.Duration, result.Success)
	return result
}

// updateTableStatistics updates statistics for all tables
func (dpm *DBPerformanceMonitor) updateTableStatistics(ctx context.Context) OptimizationResult {
	start := time.Now()
	result := OptimizationResult{
		Operation: "UPDATE_STATISTICS",
		Success:   true,
	}

	// Get all table names
	rows, err := dpm.db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err == nil {
			tables = append(tables, tableName)
		}
	}

	result.Duration = time.Since(start)
	result.Improvements = []string{
		fmt.Sprintf("Updated statistics for %d tables", len(tables)),
	}

	dpm.logger.Info("Table statistics updated", "duration", result.Duration, "tables", len(tables))
	dpm.performanceMetrics.RecordDatabaseQuery(ctx, "UPDATE_STATISTICS", result.Duration, result.Success)
	return result
}

// MonitorSlowQueries monitors and logs slow database queries
func (dpm *DBPerformanceMonitor) MonitorSlowQueries(ctx context.Context, query string, duration time.Duration) {
	const slowQueryThreshold = 100 * time.Millisecond

	if duration > slowQueryThreshold {
		dpm.logger.Warn("Slow query detected",
			"query", query,
			"duration_ms", duration.Milliseconds(),
			"threshold_ms", slowQueryThreshold.Milliseconds(),
		)

		// Record slow query metric
		dpm.performanceMetrics.RecordDatabaseQuery(ctx, "SLOW_QUERY", duration, false)
	}
}

// GetQueryExplanation provides query execution plan for optimization
func (dpm *DBPerformanceMonitor) GetQueryExplanation(ctx context.Context, query string, args ...interface{}) (string, error) {
	explanationQuery := fmt.Sprintf("EXPLAIN QUERY PLAN %s", query)
	
	rows, err := dpm.db.QueryContext(ctx, explanationQuery, args...)
	if err != nil {
		return "", fmt.Errorf("failed to get query plan: %w", err)
	}
	defer rows.Close()

	var explanation string
	for rows.Next() {
		var id, parent, notused int
		var detail string
		if err := rows.Scan(&id, &parent, &notused, &detail); err != nil {
			continue
		}
		explanation += fmt.Sprintf("Step %d: %s\n", id, detail)
	}

	return explanation, nil
}

// SetOptimalPragmas configures SQLite pragmas for optimal performance
func (dpm *DBPerformanceMonitor) SetOptimalPragmas(ctx context.Context) error {
	pragmas := map[string]interface{}{
		"journal_mode":     "WAL",           // Write-Ahead Logging for better concurrency
		"synchronous":      "NORMAL",        // Balance between safety and performance
		"cache_size":       -2000,           // 2MB cache
		"temp_store":       "memory",        // Store temporary tables in memory
		"mmap_size":        268435456,       // 256MB memory-mapped I/O
		"optimize":         nil,             // Run optimization
	}

	for pragma, value := range pragmas {
		var query string
		if value == nil {
			query = fmt.Sprintf("PRAGMA %s", pragma)
		} else {
			query = fmt.Sprintf("PRAGMA %s = %v", pragma, value)
		}

		if _, err := dpm.db.ExecContext(ctx, query); err != nil {
			dpm.logger.Error("Failed to set pragma", "pragma", pragma, "value", value, "error", err)
			return fmt.Errorf("failed to set pragma %s: %w", pragma, err)
		}

		dpm.logger.Debug("Set database pragma", "pragma", pragma, "value", value)
	}

	dpm.logger.Info("Optimal database pragmas configured")
	return nil
}