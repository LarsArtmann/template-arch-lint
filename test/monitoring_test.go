package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"github.com/LarsArtmann/template-arch-lint/internal/container"
)

// Health check response structures
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version,omitempty"`
	Checks    map[string]HealthCheck `json:"checks,omitempty"`
}

type HealthCheck struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp"`
	Duration  string                 `json:"duration,omitempty"`
}

// TestMonitoringEndpoints tests all monitoring and observability endpoints
func TestMonitoringEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	t.Run("Health Checks", func(t *testing.T) {
		testHealthEndpoints(t, router)
	})

	t.Run("Metrics", func(t *testing.T) {
		testMetricsEndpoints(t, router)
	})

	t.Run("Profiling", func(t *testing.T) {
		testProfilingEndpoints(t, router)
	})

	t.Run("Readiness and Liveness", func(t *testing.T) {
		testReadinessLiveness(t, router)
	})

	t.Run("Application Info", func(t *testing.T) {
		testApplicationInfo(t, router)
	})
}

// testHealthEndpoints tests various health check endpoints
func testHealthEndpoints(t *testing.T, router *gin.Engine) {
	t.Helper()

	t.Run("Basic Health Check", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var health HealthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &health); err != nil {
			t.Errorf("Failed to parse health response: %v", err)
			return
		}

		if health.Status != "ok" && health.Status != "healthy" {
			t.Errorf("Expected healthy status, got %s", health.Status)
		}

		if health.Timestamp == "" {
			t.Error("Health check should include timestamp")
		}

		// Validate timestamp format
		if _, err := time.Parse(time.RFC3339, health.Timestamp); err != nil {
			t.Errorf("Invalid timestamp format: %s", health.Timestamp)
		}

		t.Logf("✅ Health check passed: %+v", health)
	})

	t.Run("Health Check with Accept Header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("Expected JSON content type, got %s", contentType)
		}
	})

	t.Run("Detailed Health Check", func(t *testing.T) {
		// Try detailed health endpoint if it exists
		req, _ := http.NewRequest("GET", "/health/detailed", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// It's OK if this endpoint doesn't exist (404)
		if w.Code == http.StatusOK {
			var health HealthResponse
			if err := json.Unmarshal(w.Body.Bytes(), &health); err == nil {
				// Validate detailed response structure
				if health.Checks != nil {
					for name, check := range health.Checks {
						if check.Status == "" {
							t.Errorf("Health check '%s' missing status", name)
						}
						if check.Timestamp == "" {
							t.Errorf("Health check '%s' missing timestamp", name)
						}
					}
				}
				t.Logf("✅ Detailed health check passed: %+v", health)
			}
		}
	})
}

// testMetricsEndpoints tests Prometheus metrics and other metric endpoints
func testMetricsEndpoints(t *testing.T, router *gin.Engine) {
	t.Helper()

	t.Run("Prometheus Metrics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Metrics endpoint might not be available in test mode
		if w.Code == http.StatusOK {
			metricsContent := w.Body.String()

			// Validate Prometheus format
			expectedMetrics := []string{
				"http_requests_total",
				"http_request_duration_seconds",
				"go_memstats",
				"process_",
			}

			foundMetrics := 0
			for _, metric := range expectedMetrics {
				if strings.Contains(metricsContent, metric) {
					foundMetrics++
				}
			}

			if foundMetrics == 0 {
				t.Error("No expected metrics found in response")
			}

			// Check content type
			contentType := w.Header().Get("Content-Type")
			expectedContentType := "text/plain"
			if !strings.Contains(contentType, expectedContentType) && contentType != "" {
				t.Errorf("Expected content type %s, got %s", expectedContentType, contentType)
			}

			t.Logf("✅ Metrics endpoint working, found %d metric types", foundMetrics)
		} else if w.Code == http.StatusNotFound {
			t.Log("⚠️ Metrics endpoint not available (404)")
		} else {
			t.Errorf("Unexpected status for metrics endpoint: %d", w.Code)
		}
	})

	t.Run("Application Metrics", func(t *testing.T) {
		// Test custom application metrics endpoint if it exists
		req, _ := http.NewRequest("GET", "/api/v1/metrics", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			var metrics map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &metrics); err == nil {
				// Validate application metrics structure
				expectedFields := []string{"uptime", "requests", "users"}
				for _, field := range expectedFields {
					if _, exists := metrics[field]; exists {
						t.Logf("Found metric: %s = %v", field, metrics[field])
					}
				}
				t.Logf("✅ Application metrics available: %+v", metrics)
			}
		}
	})
}

// testProfilingEndpoints tests pprof profiling endpoints
func testProfilingEndpoints(t *testing.T, router *gin.Engine) {
	t.Helper()

	profilingEndpoints := []struct {
		name     string
		endpoint string
	}{
		{"Profile Index", "/debug/pprof/"},
		{"CPU Profile", "/debug/pprof/profile"},
		{"Heap Profile", "/debug/pprof/heap"},
		{"Goroutines", "/debug/pprof/goroutine"},
		{"Memory Stats", "/debug/pprof/allocs"},
		{"Mutex Profile", "/debug/pprof/mutex"},
		{"Block Profile", "/debug/pprof/block"},
		{"Thread Creation", "/debug/pprof/threadcreate"},
	}

	for _, test := range profilingEndpoints {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", test.endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Profiling endpoints might be disabled in production/test
			if w.Code == http.StatusOK {
				t.Logf("✅ Profiling endpoint %s available", test.endpoint)
				
				// Basic validation of profile content
				if test.endpoint == "/debug/pprof/" {
					// Index page should contain links to other profiles
					content := w.Body.String()
					if !strings.Contains(content, "goroutine") {
						t.Error("Profile index should contain goroutine link")
					}
				}
			} else if w.Code == http.StatusNotFound {
				t.Logf("⚠️ Profiling endpoint %s not available (disabled)", test.endpoint)
			} else if w.Code == http.StatusForbidden {
				t.Logf("⚠️ Profiling endpoint %s forbidden (security disabled)", test.endpoint)
			} else {
				t.Errorf("Unexpected status for %s: %d", test.endpoint, w.Code)
			}
		})
	}
}

// testReadinessLiveness tests Kubernetes-style readiness and liveness probes
func testReadinessLiveness(t *testing.T, router *gin.Engine) {
	t.Helper()

	t.Run("Liveness Probe", func(t *testing.T) {
		endpoints := []string{"/health/live", "/healthz", "/ping"}
		
		found := false
		for _, endpoint := range endpoints {
			req, _ := http.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code == http.StatusOK {
				found = true
				t.Logf("✅ Liveness probe available at %s", endpoint)
				
				// Validate response
				response := w.Body.String()
				if response == "" {
					t.Error("Liveness probe should return content")
				}
				break
			}
		}
		
		if !found {
			t.Log("⚠️ No liveness probe endpoint found")
		}
	})

	t.Run("Readiness Probe", func(t *testing.T) {
		endpoints := []string{"/health/ready", "/ready"}
		
		found := false
		for _, endpoint := range endpoints {
			req, _ := http.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code == http.StatusOK {
				found = true
				t.Logf("✅ Readiness probe available at %s", endpoint)
				
				// Readiness should check dependencies
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
					if status, exists := response["status"]; exists {
						if status != "ready" && status != "ok" {
							t.Errorf("Unexpected readiness status: %v", status)
						}
					}
				}
				break
			}
		}
		
		if !found {
			t.Log("⚠️ No readiness probe endpoint found")
		}
	})

	t.Run("Startup Probe", func(t *testing.T) {
		endpoints := []string{"/health/startup", "/startup"}
		
		for _, endpoint := range endpoints {
			req, _ := http.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code == http.StatusOK {
				t.Logf("✅ Startup probe available at %s", endpoint)
				break
			}
		}
	})
}

// testApplicationInfo tests application information endpoints
func testApplicationInfo(t *testing.T, router *gin.Engine) {
	t.Helper()

	t.Run("Version Info", func(t *testing.T) {
		endpoints := []string{"/version", "/info", "/api/v1/info"}
		
		found := false
		for _, endpoint := range endpoints {
			req, _ := http.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code == http.StatusOK {
				found = true
				t.Logf("✅ Version info available at %s", endpoint)
				
				var info map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &info); err == nil {
					// Validate version info structure
					expectedFields := []string{"version", "build", "commit"}
					for _, field := range expectedFields {
						if value, exists := info[field]; exists {
							t.Logf("Version info %s: %v", field, value)
						}
					}
				}
				break
			}
		}
		
		if !found {
			t.Log("⚠️ No version info endpoint found")
		}
	})

	t.Run("Build Info", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/build", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code == http.StatusOK {
			var buildInfo map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &buildInfo); err == nil {
				expectedFields := []string{"time", "version", "go_version"}
				for _, field := range expectedFields {
					if value, exists := buildInfo[field]; exists {
						t.Logf("Build info %s: %v", field, value)
					}
				}
				t.Logf("✅ Build info available: %+v", buildInfo)
			}
		}
	})
}

// TestHealthCheckComponents tests individual health check components
func TestHealthCheckComponents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	t.Run("Database Health", func(t *testing.T) {
		// Test database connectivity through health endpoint
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			var health HealthResponse
			if err := json.Unmarshal(w.Body.Bytes(), &health); err == nil {
				if health.Checks != nil {
					if dbCheck, exists := health.Checks["database"]; exists {
						if dbCheck.Status != "ok" && dbCheck.Status != "healthy" {
							t.Errorf("Database health check failed: %s", dbCheck.Status)
						} else {
							t.Logf("✅ Database health check passed: %s", dbCheck.Status)
						}
					}
				}
			}
		}
	})

	t.Run("Memory Health", func(t *testing.T) {
		// Test memory usage monitoring
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			var health HealthResponse
			if err := json.Unmarshal(w.Body.Bytes(), &health); err == nil {
				if health.Checks != nil {
					if memCheck, exists := health.Checks["memory"]; exists {
						if memCheck.Details != nil {
							if usage, exists := memCheck.Details["usage_percent"]; exists {
								if usageFloat, ok := usage.(float64); ok && usageFloat > 90 {
									t.Errorf("Memory usage too high: %.2f%%", usageFloat)
								}
							}
						}
						t.Logf("✅ Memory health check available")
					}
				}
			}
		}
	})

	t.Run("Disk Health", func(t *testing.T) {
		// Test disk space monitoring
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			var health HealthResponse
			if err := json.Unmarshal(w.Body.Bytes(), &health); err == nil {
				if health.Checks != nil {
					if diskCheck, exists := health.Checks["disk"]; exists {
						t.Logf("✅ Disk health check available: %s", diskCheck.Status)
					}
				}
			}
		}
	})
}

// TestMonitoringPerformance tests the performance of monitoring endpoints
func TestMonitoringPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	t.Run("Health Check Performance", func(t *testing.T) {
		iterations := 100
		start := time.Now()
		
		for i := 0; i < iterations; i++ {
			req, _ := http.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				t.Errorf("Health check failed on iteration %d: status %d", i, w.Code)
			}
		}
		
		duration := time.Since(start)
		avgTime := duration / time.Duration(iterations)
		
		t.Logf("Health check performance: %d requests in %v (avg: %v)", iterations, duration, avgTime)
		
		// Health checks should be fast (< 10ms average)
		if avgTime > 10*time.Millisecond {
			t.Errorf("Health check too slow: %v (expected < 10ms)", avgTime)
		}
	})

	t.Run("Metrics Endpoint Performance", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()
		
		start := time.Now()
		router.ServeHTTP(w, req)
		duration := time.Since(start)
		
		if w.Code == http.StatusOK {
			t.Logf("Metrics endpoint response time: %v", duration)
			
			// Metrics collection should be reasonable (< 100ms)
			if duration > 100*time.Millisecond {
				t.Errorf("Metrics endpoint too slow: %v (expected < 100ms)", duration)
			}
		}
	})
}

// TestMonitoringConcurrency tests monitoring endpoints under concurrent load
func TestMonitoringConcurrency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	t.Run("Concurrent Health Checks", func(t *testing.T) {
		concurrency := 10
		iterations := 100
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		done := make(chan error, concurrency)
		
		for i := 0; i < concurrency; i++ {
			go func(worker int) {
				defer func() { done <- nil }()
				
				for j := 0; j < iterations/concurrency; j++ {
					select {
					case <-ctx.Done():
						return
					default:
						req, _ := http.NewRequest("GET", "/health", nil)
						w := httptest.NewRecorder()
						router.ServeHTTP(w, req)
						
						if w.Code != http.StatusOK {
							done <- fmt.Errorf("worker %d: health check failed with status %d", worker, w.Code)
							return
						}
					}
				}
			}(i)
		}
		
		// Wait for all workers
		for i := 0; i < concurrency; i++ {
			if err := <-done; err != nil {
				t.Errorf("Concurrent health check failed: %v", err)
			}
		}
		
		t.Logf("✅ Concurrent health checks completed successfully")
	})
}

// TestMonitoringConfiguration tests monitoring configuration
func TestMonitoringConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Monitoring Features Enabled", func(t *testing.T) {
		// This would typically test if monitoring features are properly configured
		// For now, we'll test that the router is properly set up
		router := setupTestRouter(t)
		
		// Test that basic endpoints are available
		endpoints := []string{"/health"}
		
		for _, endpoint := range endpoints {
			req, _ := http.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code == http.StatusOK {
				t.Logf("✅ Monitoring endpoint %s configured", endpoint)
			} else {
				t.Errorf("Monitoring endpoint %s not configured properly: status %d", endpoint, w.Code)
			}
		}
	})
}

// Helper function to setup test router (reused from integration tests)
func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	
	gin.SetMode(gin.TestMode)
	
	diContainer := container.New()
	t.Cleanup(func() {
		if err := diContainer.Shutdown(); err != nil {
			t.Errorf("Error shutting down container: %v", err)
		}
	})
	
	if err := diContainer.RegisterAll(); err != nil {
		t.Fatalf("Failed to register dependencies: %v", err)
	}
	
	injector := diContainer.GetInjector()
	return do.MustInvoke[*gin.Engine](injector)
}