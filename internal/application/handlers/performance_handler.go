package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceHandler provides performance monitoring and profiling endpoints.
type PerformanceHandler struct {
	logger    *slog.Logger
	startTime time.Time
}

// NewPerformanceHandler creates a new performance handler.
func NewPerformanceHandler(logger *slog.Logger) *PerformanceHandler {
	return &PerformanceHandler{
		logger:    logger,
		startTime: time.Now(),
	}
}

// RuntimeStats provides runtime statistics and metrics.
func (h *PerformanceHandler) RuntimeStats(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get GC stats
	gcStats := debug.GCStats{}
	debug.ReadGCStats(&gcStats)

	// Get build info
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		buildInfo = &debug.BuildInfo{}
	}

	stats := gin.H{
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(h.startTime).String(),
		"memory": gin.H{
			"alloc_mb":         bToMb(m.Alloc),
			"total_alloc_mb":   bToMb(m.TotalAlloc),
			"sys_mb":           bToMb(m.Sys),
			"heap_alloc_mb":    bToMb(m.HeapAlloc),
			"heap_sys_mb":      bToMb(m.HeapSys),
			"heap_idle_mb":     bToMb(m.HeapIdle),
			"heap_inuse_mb":    bToMb(m.HeapInuse),
			"heap_released_mb": bToMb(m.HeapReleased),
			"heap_objects":     m.HeapObjects,
			"stack_inuse_mb":   bToMb(m.StackInuse),
			"stack_sys_mb":     bToMb(m.StackSys),
		},
		"gc": gin.H{
			"num_gc":          m.NumGC,
			"num_forced_gc":   m.NumForcedGC,
			"pause_total_ns":  m.PauseTotalNs,
			"last_gc":         time.Unix(0, int64(m.LastGC)).UTC(),
			"gc_cpu_fraction": m.GCCPUFraction,
			"next_gc_mb":      bToMb(m.NextGC),
		},
		"goroutines": gin.H{
			"total": runtime.NumGoroutine(),
		},
		"cpu": gin.H{
			"num_cpu":      runtime.NumCPU(),
			"gomaxprocs":   runtime.GOMAXPROCS(0),
			"num_cgo_call": runtime.NumCgoCall(),
		},
		"build": gin.H{
			"go_version": runtime.Version(),
			"path":       buildInfo.Path,
			"version":    buildInfo.Main.Version,
		},
	}

	h.logger.Debug("Runtime stats requested", "goroutines", runtime.NumGoroutine(), "alloc_mb", bToMb(m.Alloc))
	c.JSON(http.StatusOK, stats)
}

// ForceGC triggers garbage collection and returns stats.
func (h *PerformanceHandler) ForceGC(c *gin.Context) {
	var beforeGC, afterGC runtime.MemStats

	// Get memory stats before GC
	runtime.ReadMemStats(&beforeGC)
	beforeTime := time.Now()

	// Force garbage collection
	runtime.GC()

	// Get memory stats after GC
	afterTime := time.Now()
	runtime.ReadMemStats(&afterGC)

	result := gin.H{
		"timestamp":      time.Now().UTC(),
		"gc_duration_ms": afterTime.Sub(beforeTime).Milliseconds(),
		"before_gc": gin.H{
			"alloc_mb":     bToMb(beforeGC.Alloc),
			"heap_objects": beforeGC.HeapObjects,
			"num_gc":       beforeGC.NumGC,
		},
		"after_gc": gin.H{
			"alloc_mb":     bToMb(afterGC.Alloc),
			"heap_objects": afterGC.HeapObjects,
			"num_gc":       afterGC.NumGC,
		},
		"freed": gin.H{
			"memory_mb": bToMb(beforeGC.Alloc - afterGC.Alloc),
			"objects":   beforeGC.HeapObjects - afterGC.HeapObjects,
		},
	}

	h.logger.Info("Forced garbage collection",
		"duration_ms", afterTime.Sub(beforeTime).Milliseconds(),
		"freed_mb", bToMb(beforeGC.Alloc-afterGC.Alloc),
		"freed_objects", beforeGC.HeapObjects-afterGC.HeapObjects,
	)

	c.JSON(http.StatusOK, result)
}

// HealthMetrics provides application health and performance metrics.
func (h *PerformanceHandler) HealthMetrics(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Define health thresholds
	const (
		maxGoroutines     = 1000
		maxHeapMB         = 512
		maxGCPauseNs      = 100_000_000 // 100ms
		criticalGCPauseNs = 500_000_000 // 500ms
	)

	goroutines := runtime.NumGoroutine()
	heapMB := bToMb(m.HeapAlloc)
	avgGCPauseNs := uint64(0)
	if m.NumGC > 0 {
		avgGCPauseNs = m.PauseTotalNs / uint64(m.NumGC)
	}

	// Determine overall health status
	status := "healthy"
	issues := []string{}

	if goroutines > maxGoroutines {
		status = "warning"
		issues = append(issues, fmt.Sprintf("High goroutine count: %d", goroutines))
	}

	if heapMB > maxHeapMB {
		status = "warning"
		issues = append(issues, fmt.Sprintf("High heap usage: %.2f MB", heapMB))
	}

	if avgGCPauseNs > criticalGCPauseNs {
		status = "critical"
		issues = append(issues, fmt.Sprintf("Critical GC pause times: %.2f ms", float64(avgGCPauseNs)/1_000_000))
	} else if avgGCPauseNs > maxGCPauseNs {
		if status == "healthy" {
			status = "warning"
		}
		issues = append(issues, fmt.Sprintf("High GC pause times: %.2f ms", float64(avgGCPauseNs)/1_000_000))
	}

	result := gin.H{
		"timestamp": time.Now().UTC(),
		"status":    status,
		"uptime":    time.Since(h.startTime).String(),
		"issues":    issues,
		"metrics": gin.H{
			"goroutines":      goroutines,
			"heap_mb":         heapMB,
			"gc_pause_avg_ms": float64(avgGCPauseNs) / 1_000_000,
			"gc_count":        m.NumGC,
			"cpu_count":       runtime.NumCPU(),
		},
		"thresholds": gin.H{
			"max_goroutines":       maxGoroutines,
			"max_heap_mb":          maxHeapMB,
			"max_gc_pause_ms":      float64(maxGCPauseNs) / 1_000_000,
			"critical_gc_pause_ms": float64(criticalGCPauseNs) / 1_000_000,
		},
	}

	statusCode := http.StatusOK
	switch status {
	case "critical":
		statusCode = http.StatusServiceUnavailable
	case "warning":
		statusCode = http.StatusAccepted
	}

	h.logger.Debug("Health metrics requested",
		"status", status,
		"goroutines", goroutines,
		"heap_mb", heapMB,
		"issues_count", len(issues),
	)

	c.JSON(statusCode, result)
}

// DebugInfo provides debugging information.
func (h *PerformanceHandler) DebugInfo(c *gin.Context) {
	buildInfo, _ := debug.ReadBuildInfo()

	info := gin.H{
		"timestamp": time.Now().UTC(),
		"runtime": gin.H{
			"go_version":   runtime.Version(),
			"goos":         runtime.GOOS,
			"goarch":       runtime.GOARCH,
			"num_cpu":      runtime.NumCPU(),
			"gomaxprocs":   runtime.GOMAXPROCS(0),
			"num_cgo_call": runtime.NumCgoCall(),
		},
		"build": gin.H{
			"path":    buildInfo.Path,
			"version": buildInfo.Main.Version,
		},
		"application": gin.H{
			"uptime":     time.Since(h.startTime).String(),
			"start_time": h.startTime.UTC(),
		},
	}

	c.JSON(http.StatusOK, info)
}

// MemoryDump provides detailed memory allocation information.
func (h *PerformanceHandler) MemoryDump(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	dump := gin.H{
		"timestamp": time.Now().UTC(),
		"general": gin.H{
			"alloc_mb":       bToMb(m.Alloc),
			"total_alloc_mb": bToMb(m.TotalAlloc),
			"sys_mb":         bToMb(m.Sys),
			"lookups":        m.Lookups,
			"mallocs":        m.Mallocs,
			"frees":          m.Frees,
		},
		"heap": gin.H{
			"heap_alloc_mb":    bToMb(m.HeapAlloc),
			"heap_sys_mb":      bToMb(m.HeapSys),
			"heap_idle_mb":     bToMb(m.HeapIdle),
			"heap_inuse_mb":    bToMb(m.HeapInuse),
			"heap_released_mb": bToMb(m.HeapReleased),
			"heap_objects":     m.HeapObjects,
		},
		"stack": gin.H{
			"stack_inuse_mb": bToMb(m.StackInuse),
			"stack_sys_mb":   bToMb(m.StackSys),
		},
		"off_heap": gin.H{
			"mspan_inuse_mb":   bToMb(m.MSpanInuse),
			"mspan_sys_mb":     bToMb(m.MSpanSys),
			"mcache_inuse_mb":  bToMb(m.MCacheInuse),
			"mcache_sys_mb":    bToMb(m.MCacheSys),
			"buck_hash_sys_mb": bToMb(m.BuckHashSys),
		},
		"gc": gin.H{
			"next_gc_mb":      bToMb(m.NextGC),
			"last_gc":         time.Unix(0, int64(m.LastGC)).UTC(),
			"pause_total_ns":  m.PauseTotalNs,
			"num_gc":          m.NumGC,
			"num_forced_gc":   m.NumForcedGC,
			"gc_sys_mb":       bToMb(m.GCSys),
			"other_sys_mb":    bToMb(m.OtherSys),
			"gc_cpu_fraction": m.GCCPUFraction,
		},
	}

	c.JSON(http.StatusOK, dump)
}

// bToMb converts bytes to megabytes.
func bToMb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}
