package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"github.com/LarsArtmann/template-arch-lint/internal/container"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	inmemory "github.com/LarsArtmann/template-arch-lint/internal/infrastructure/repositories"
)

// BenchmarkResults holds performance metrics
type BenchmarkResults struct {
	Operation     string
	TotalTime     time.Duration
	AverageTime   time.Duration
	MinTime       time.Duration
	MaxTime       time.Duration
	Operations    int
	OpsPerSecond  float64
	MemoryUsage   int64
	Concurrency   int
}

// Performance test constants
const (
	BenchmarkIterations = 1000
	ConcurrentUsers     = 50
	BenchmarkTimeout    = 30 * time.Second
)

// BenchmarkUserOperations benchmarks core user operations
func BenchmarkUserOperations(b *testing.B) {
	userRepo := inmemory.NewInMemoryUserRepository()
	userService := services.NewUserService(userRepo)
	ctx := context.Background()

	b.Run("CreateUser", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			id, _ := values.NewUserID(fmt.Sprintf("user-%d", i))
			email := fmt.Sprintf("user-%d@example.com", i)
			name := fmt.Sprintf("User %d", i)
			
			_, err := userService.CreateUser(ctx, id, email, name)
			if err != nil {
				b.Fatalf("CreateUser failed: %v", err)
			}
		}
	})

	// Setup users for read benchmarks
	setupUsers := func(count int) {
		for i := 0; i < count; i++ {
			id, _ := values.NewUserID(fmt.Sprintf("bench-user-%d", i))
			email := fmt.Sprintf("bench-user-%d@example.com", i)
			name := fmt.Sprintf("Bench User %d", i)
			userService.CreateUser(ctx, id, email, name)
		}
	}

	b.Run("GetUser", func(b *testing.B) {
		setupUsers(1000)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			id, _ := values.NewUserID(fmt.Sprintf("bench-user-%d", i%1000))
			_, err := userService.GetUser(ctx, id)
			if err != nil {
				b.Fatalf("GetUser failed: %v", err)
			}
		}
	})

	b.Run("ListUsers", func(b *testing.B) {
		setupUsers(100)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			_, err := userService.ListUsers(ctx)
			if err != nil {
				b.Fatalf("ListUsers failed: %v", err)
			}
		}
	})

	b.Run("UpdateUser", func(b *testing.B) {
		setupUsers(1000)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			id, _ := values.NewUserID(fmt.Sprintf("bench-user-%d", i%1000))
			email := fmt.Sprintf("updated-user-%d@example.com", i)
			name := fmt.Sprintf("Updated User %d", i)
			
			_, err := userService.UpdateUser(ctx, id, email, name)
			if err != nil {
				b.Fatalf("UpdateUser failed: %v", err)
			}
		}
	})

	b.Run("FunctionalOperations", func(b *testing.B) {
		setupUsers(100)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			switch i % 4 {
			case 0:
				userService.FilterActiveUsers(ctx)
			case 1:
				userService.GetUserEmailsWithResult(ctx)
			case 2:
				userService.GetUserStats(ctx)
			case 3:
				filters := map[string]interface{}{
					"domain": "example.com",
					"active": true,
				}
				userService.GetUsersWithFilters(ctx, filters)
			}
		}
	})
}

// BenchmarkValueObjects benchmarks value object operations
func BenchmarkValueObjects(b *testing.B) {
	b.Run("EmailCreation", func(b *testing.B) {
		emails := generateTestEmails(b.N)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			_, err := values.NewEmail(emails[i%len(emails)])
			if err != nil {
				b.Fatalf("NewEmail failed: %v", err)
			}
		}
	})

	b.Run("UserIDCreation", func(b *testing.B) {
		userIDs := generateTestUserIDs(b.N)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			_, err := values.NewUserID(userIDs[i%len(userIDs)])
			if err != nil {
				b.Fatalf("NewUserID failed: %v", err)
			}
		}
	})

	b.Run("UserIDGeneration", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			_, err := values.GenerateUserID()
			if err != nil {
				b.Fatalf("GenerateUserID failed: %v", err)
			}
		}
	})

	b.Run("UserNameCreation", func(b *testing.B) {
		names := generateTestUserNames(b.N)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			_, err := values.NewUserName(names[i%len(names)])
			if err != nil {
				b.Fatalf("NewUserName failed: %v", err)
			}
		}
	})
}

// BenchmarkHTTPEndpoints benchmarks HTTP endpoints
func BenchmarkHTTPEndpoints(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := setupBenchmarkRouter(b)

	b.Run("HealthCheck", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Fatalf("Health check failed with status %d", w.Code)
			}
		}
	})

	b.Run("CreateUserEndpoint", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			userPayload := map[string]string{
				"id":    fmt.Sprintf("bench-user-%d", i),
				"email": fmt.Sprintf("bench-user-%d@example.com", i),
				"name":  fmt.Sprintf("Bench User %d", i),
			}
			
			jsonPayload, _ := json.Marshal(userPayload)
			req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusCreated {
				b.Fatalf("Create user failed with status %d", w.Code)
			}
		}
	})

	b.Run("GetUserEndpoint", func(b *testing.B) {
		// Pre-create users for benchmarking
		for i := 0; i < 1000; i++ {
			userPayload := map[string]string{
				"id":    fmt.Sprintf("get-bench-user-%d", i),
				"email": fmt.Sprintf("get-bench-user-%d@example.com", i),
				"name":  fmt.Sprintf("Get Bench User %d", i),
			}
			
			jsonPayload, _ := json.Marshal(userPayload)
			req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
		
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			userID := fmt.Sprintf("get-bench-user-%d", i%1000)
			req, _ := http.NewRequest("GET", "/api/v1/users/"+userID, nil)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Fatalf("Get user failed with status %d", w.Code)
			}
		}
	})

	b.Run("ListUsersEndpoint", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("GET", "/api/v1/users", nil)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Fatalf("List users failed with status %d", w.Code)
			}
		}
	})
}

// BenchmarkConcurrentOperations benchmarks concurrent scenarios
func BenchmarkConcurrentOperations(b *testing.B) {
	userRepo := inmemory.NewInMemoryUserRepository()
	userService := services.NewUserService(userRepo)
	ctx := context.Background()

	b.Run("ConcurrentCreateUsers", func(b *testing.B) {
		b.SetParallelism(10)
		b.ResetTimer()
		
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				id, _ := values.NewUserID(fmt.Sprintf("concurrent-user-%d-%d", b.N, i))
				email := fmt.Sprintf("concurrent-user-%d-%d@example.com", b.N, i)
				name := fmt.Sprintf("Concurrent User %d-%d", b.N, i)
				
				_, err := userService.CreateUser(ctx, id, email, name)
				if err != nil {
					b.Fatalf("Concurrent CreateUser failed: %v", err)
				}
				i++
			}
		})
	})

	b.Run("ConcurrentReadUsers", func(b *testing.B) {
		// Pre-create users
		for i := 0; i < 1000; i++ {
			id, _ := values.NewUserID(fmt.Sprintf("read-user-%d", i))
			email := fmt.Sprintf("read-user-%d@example.com", i)
			name := fmt.Sprintf("Read User %d", i)
			userService.CreateUser(ctx, id, email, name)
		}
		
		b.SetParallelism(10)
		b.ResetTimer()
		
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				id, _ := values.NewUserID(fmt.Sprintf("read-user-%d", i%1000))
				_, err := userService.GetUser(ctx, id)
				if err != nil {
					b.Fatalf("Concurrent GetUser failed: %v", err)
				}
				i++
			}
		})
	})

	b.Run("ConcurrentMixedOperations", func(b *testing.B) {
		b.SetParallelism(5)
		b.ResetTimer()
		
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				switch i % 4 {
				case 0: // Create
					id, _ := values.NewUserID(fmt.Sprintf("mixed-user-%d-%d", b.N, i))
					email := fmt.Sprintf("mixed-user-%d-%d@example.com", b.N, i)
					name := fmt.Sprintf("Mixed User %d-%d", b.N, i)
					userService.CreateUser(ctx, id, email, name)
				case 1: // Read
					userService.ListUsers(ctx)
				case 2: // Stats
					userService.GetUserStats(ctx)
				case 3: // Filter
					userService.FilterActiveUsers(ctx)
				}
				i++
			}
		})
	})
}

// TestPerformanceBaselines establishes performance baselines
func TestPerformanceBaselines(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	t.Run("UserService Performance", func(t *testing.T) {
		results := measureUserServicePerformance(t)
		
		// Assert performance baselines
		if results.OpsPerSecond < 1000 {
			t.Errorf("User creation performance below baseline: %.2f ops/sec (expected >1000)", results.OpsPerSecond)
		}
		
		if results.AverageTime > 10*time.Millisecond {
			t.Errorf("Average response time too high: %v (expected <10ms)", results.AverageTime)
		}
		
		t.Logf("Performance Results: %+v", results)
	})

	t.Run("Memory Usage", func(t *testing.T) {
		initialMem := getMemUsage()
		
		// Perform operations
		userRepo := inmemory.NewInMemoryUserRepository()
		userService := services.NewUserService(userRepo)
		ctx := context.Background()
		
		for i := 0; i < 10000; i++ {
			id, _ := values.NewUserID(fmt.Sprintf("mem-user-%d", i))
			email := fmt.Sprintf("mem-user-%d@example.com", i)
			name := fmt.Sprintf("Mem User %d", i)
			userService.CreateUser(ctx, id, email, name)
		}
		
		finalMem := getMemUsage()
		memIncrease := finalMem - initialMem
		
		t.Logf("Memory usage increased by %d bytes for 10k users", memIncrease)
		
		// Assert reasonable memory usage (less than 100MB for 10k users)
		if memIncrease > 100*1024*1024 {
			t.Errorf("Memory usage too high: %d bytes (expected <100MB)", memIncrease)
		}
	})

	t.Run("Concurrent Performance", func(t *testing.T) {
		results := measureConcurrentPerformance(t, ConcurrentUsers)
		
		// Assert concurrent performance
		if results.OpsPerSecond < 500 {
			t.Errorf("Concurrent performance below baseline: %.2f ops/sec (expected >500)", results.OpsPerSecond)
		}
		
		t.Logf("Concurrent Performance Results: %+v", results)
	})

	t.Run("HTTP Endpoint Performance", func(t *testing.T) {
		results := measureHTTPEndpointPerformance(t)
		
		// Assert HTTP performance
		if results.OpsPerSecond < 100 {
			t.Errorf("HTTP performance below baseline: %.2f ops/sec (expected >100)", results.OpsPerSecond)
		}
		
		if results.AverageTime > 100*time.Millisecond {
			t.Errorf("HTTP average response time too high: %v (expected <100ms)", results.AverageTime)
		}
		
		t.Logf("HTTP Performance Results: %+v", results)
	})
}

// Helper functions

func setupBenchmarkRouter(b *testing.B) *gin.Engine {
	b.Helper()
	
	gin.SetMode(gin.TestMode)
	
	diContainer := container.New()
	b.Cleanup(func() {
		if err := diContainer.Shutdown(); err != nil {
			b.Errorf("Error shutting down container: %v", err)
		}
	})
	
	if err := diContainer.RegisterAll(); err != nil {
		b.Fatalf("Failed to register dependencies: %v", err)
	}
	
	injector := diContainer.GetInjector()
	return do.MustInvoke[*gin.Engine](injector)
}

func measureUserServicePerformance(t *testing.T) BenchmarkResults {
	t.Helper()
	
	userRepo := inmemory.NewInMemoryUserRepository()
	userService := services.NewUserService(userRepo)
	ctx := context.Background()
	
	iterations := BenchmarkIterations
	times := make([]time.Duration, iterations)
	
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		opStart := time.Now()
		
		id, _ := values.NewUserID(fmt.Sprintf("perf-user-%d", i))
		email := fmt.Sprintf("perf-user-%d@example.com", i)
		name := fmt.Sprintf("Perf User %d", i)
		
		_, err := userService.CreateUser(ctx, id, email, name)
		if err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}
		
		times[i] = time.Since(opStart)
	}
	
	totalTime := time.Since(start)
	
	return calculateBenchmarkResults("UserService.CreateUser", times, totalTime, iterations, 1)
}

func measureConcurrentPerformance(t *testing.T, concurrency int) BenchmarkResults {
	t.Helper()
	
	userRepo := inmemory.NewInMemoryUserRepository()
	userService := services.NewUserService(userRepo)
	ctx := context.Background()
	
	iterations := BenchmarkIterations
	times := make([]time.Duration, iterations)
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			opStart := time.Now()
			
			id, _ := values.NewUserID(fmt.Sprintf("concurrent-perf-user-%d", index))
			email := fmt.Sprintf("concurrent-perf-user-%d@example.com", index)
			name := fmt.Sprintf("Concurrent Perf User %d", index)
			
			_, err := userService.CreateUser(ctx, id, email, name)
			if err != nil {
				t.Errorf("Concurrent CreateUser failed: %v", err)
				return
			}
			
			duration := time.Since(opStart)
			
			mu.Lock()
			times[index] = duration
			mu.Unlock()
		}(i)
	}
	
	wg.Wait()
	totalTime := time.Since(start)
	
	return calculateBenchmarkResults("ConcurrentUserService.CreateUser", times, totalTime, iterations, concurrency)
}

func measureHTTPEndpointPerformance(t *testing.T) BenchmarkResults {
	t.Helper()
	
	gin.SetMode(gin.TestMode)
	router := setupPerfTestRouter(t)
	
	iterations := 100 // Fewer iterations for HTTP due to overhead
	times := make([]time.Duration, iterations)
	
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		opStart := time.Now()
		
		userPayload := map[string]string{
			"id":    fmt.Sprintf("http-perf-user-%d", i),
			"email": fmt.Sprintf("http-perf-user-%d@example.com", i),
			"name":  fmt.Sprintf("HTTP Perf User %d", i),
		}
		
		jsonPayload, _ := json.Marshal(userPayload)
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusCreated {
			t.Fatalf("HTTP CreateUser failed with status %d", w.Code)
		}
		
		times[i] = time.Since(opStart)
	}
	
	totalTime := time.Since(start)
	
	return calculateBenchmarkResults("HTTP.CreateUser", times, totalTime, iterations, 1)
}

func setupPerfTestRouter(t *testing.T) *gin.Engine {
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

func calculateBenchmarkResults(operation string, times []time.Duration, totalTime time.Duration, iterations, concurrency int) BenchmarkResults {
	if len(times) == 0 {
		return BenchmarkResults{}
	}
	
	// Calculate statistics
	var total time.Duration
	minTime := times[0]
	maxTime := times[0]
	
	for _, t := range times {
		total += t
		if t < minTime {
			minTime = t
		}
		if t > maxTime {
			maxTime = t
		}
	}
	
	avgTime := total / time.Duration(len(times))
	opsPerSec := float64(iterations) / totalTime.Seconds()
	
	return BenchmarkResults{
		Operation:    operation,
		TotalTime:    totalTime,
		AverageTime:  avgTime,
		MinTime:      minTime,
		MaxTime:      maxTime,
		Operations:   iterations,
		OpsPerSecond: opsPerSec,
		MemoryUsage:  getMemUsage(),
		Concurrency:  concurrency,
	}
}

func getMemUsage() int64 {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	return int64(m.Alloc)
}

func generateTestEmails(count int) []string {
	emails := make([]string, min(count, 1000))
	for i := 0; i < len(emails); i++ {
		emails[i] = fmt.Sprintf("test%d@example.com", i)
	}
	return emails
}

func generateTestUserIDs(count int) []string {
	ids := make([]string, min(count, 1000))
	for i := 0; i < len(ids); i++ {
		ids[i] = fmt.Sprintf("user-%d", i)
	}
	return ids
}

func generateTestUserNames(count int) []string {
	names := make([]string, min(count, 1000))
	for i := 0; i < len(names); i++ {
		names[i] = fmt.Sprintf("Test User %d", i)
	}
	return names
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}