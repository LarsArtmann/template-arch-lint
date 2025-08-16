package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestRateLimiterMiddleware(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Rate Limiter Middleware Suite")
}

var _ = ginkgo.Describe("Rate Limiter Middleware", func() {
	var (
		router *gin.Engine
		logger *slog.Logger
	)

	ginkgo.BeforeEach(func() {
		gin.SetMode(gin.TestMode)

		// Create test logger that outputs to discard to avoid test noise
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelError, // Only show errors during tests
		}))

		router = gin.New()
		router.Use(gin.Recovery())
	})

	ginkgo.Describe("Rate Limiting Configuration", func() {
		ginkgo.It("should use default configuration when none provided", func() {
			config := DefaultRateLimitConfig()

			gomega.Expect(config.GeneralRPS).To(gomega.BeNumerically("~", 100.0/60.0, 0.01))
			gomega.Expect(config.GeneralBurst).To(gomega.Equal(10))
			gomega.Expect(config.AuthRPS).To(gomega.BeNumerically("~", 10.0/60.0, 0.01))
			gomega.Expect(config.AuthBurst).To(gomega.Equal(3))
			gomega.Expect(config.CleanupInterval).To(gomega.Equal(5 * time.Minute))
		})

		ginkgo.It("should create rate limiter store with correct configuration", func() {
			config := RateLimitConfig{
				GeneralRPS:      2.0,
				GeneralBurst:    5,
				AuthRPS:         0.5,
				AuthBurst:       2,
				Logger:          logger,
				CleanupInterval: time.Second,
			}

			store := NewRateLimiterStore(config)

			gomega.Expect(store.generalRPS).To(gomega.Equal(2.0))
			gomega.Expect(store.generalBurst).To(gomega.Equal(5))
			gomega.Expect(store.authRPS).To(gomega.Equal(0.5))
			gomega.Expect(store.authBurst).To(gomega.Equal(2))
		})
	})

	ginkgo.Describe("Rate Limit Type Detection", func() {
		ginkgo.It("should detect auth endpoints correctly", func() {
			authPaths := []string{
				"/api/v1/auth/login",
				"/auth/register",
				"/login",
				"/register",
				"/reset-password",
				"/verify-email",
				"/api/v1/users",
				"/users",
			}

			for _, path := range authPaths {
				limitType := getRateLimitType(path)
				gomega.Expect(limitType).To(gomega.Equal(AuthRateLimit),
					"Expected auth rate limit for path: %s", path)
			}
		})

		ginkgo.It("should detect general endpoints correctly", func() {
			generalPaths := []string{
				"/api/v1/products",
				"/health",
				"/metrics",
				"/api/v1/orders",
				"/dashboard",
			}

			for _, path := range generalPaths {
				limitType := getRateLimitType(path)
				gomega.Expect(limitType).To(gomega.Equal(GeneralRateLimit),
					"Expected general rate limit for path: %s", path)
			}
		})
	})

	ginkgo.Describe("Client ID Extraction", func() {
		ginkgo.It("should extract client ID from X-Forwarded-For header", func() {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			clientID := getClientID(c)
			gomega.Expect(clientID).To(gomega.Equal("203.0.113.1"))
		})

		ginkgo.It("should extract client ID from X-Real-IP header", func() {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Real-IP", "203.0.113.2")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			clientID := getClientID(c)
			gomega.Expect(clientID).To(gomega.Equal("203.0.113.2"))
		})

		ginkgo.It("should prioritize X-Forwarded-For over X-Real-IP", func() {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Forwarded-For", "203.0.113.1")
			req.Header.Set("X-Real-IP", "203.0.113.2")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			clientID := getClientID(c)
			gomega.Expect(clientID).To(gomega.Equal("203.0.113.1"))
		})
	})

	ginkgo.Describe("Rate Limiting Behavior", func() {
		ginkgo.BeforeEach(func() {
			// Use very permissive rate limits for testing
			config := RateLimitConfig{
				GeneralRPS:      10.0, // 10 RPS
				GeneralBurst:    5,    // Burst of 5
				AuthRPS:         2.0,  // 2 RPS
				AuthBurst:       2,    // Burst of 2
				Logger:          logger,
				CleanupInterval: time.Second, // Use shorter interval for tests
			}

			router.Use(RateLimiter(config))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})
			router.POST("/api/v1/users", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "user created"})
			})
		})

		ginkgo.It("should allow requests within general rate limit", func() {
			for i := 0; i < 5; i++ { // Within burst limit
				req, _ := http.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Forwarded-For", "203.0.113.1")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				gomega.Expect(w.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(w.Header().Get("X-RateLimit-Limit")).To(gomega.Equal("100"))
				gomega.Expect(w.Header().Get("X-RateLimit-Type")).To(gomega.Equal("general"))
			}
		})

		ginkgo.It("should allow requests within auth rate limit", func() {
			for i := 0; i < 2; i++ { // Within burst limit
				req, _ := http.NewRequest("POST", "/api/v1/users", nil)
				req.Header.Set("X-Forwarded-For", "203.0.113.2")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				gomega.Expect(w.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(w.Header().Get("X-RateLimit-Limit")).To(gomega.Equal("10"))
				gomega.Expect(w.Header().Get("X-RateLimit-Type")).To(gomega.Equal("auth"))
			}
		})

		ginkgo.It("should block requests exceeding auth rate limit", func() {
			// Exhaust the burst limit first
			for i := 0; i < 2; i++ {
				req, _ := http.NewRequest("POST", "/api/v1/users", nil)
				req.Header.Set("X-Forwarded-For", "203.0.113.3")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				gomega.Expect(w.Code).To(gomega.Equal(http.StatusOK))
			}

			// This request should be rate limited
			req, _ := http.NewRequest("POST", "/api/v1/users", nil)
			req.Header.Set("X-Forwarded-For", "203.0.113.3")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			gomega.Expect(w.Code).To(gomega.Equal(http.StatusTooManyRequests))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(response["error"]).To(gomega.Equal("Too many requests"))
			gomega.Expect(response["retry_after"]).To(gomega.Equal(float64(60)))
		})

		ginkgo.It("should include correct rate limit headers", func() {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Forwarded-For", "203.0.113.4")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			gomega.Expect(w.Code).To(gomega.Equal(http.StatusOK))

			// Check headers
			gomega.Expect(w.Header().Get("X-RateLimit-Limit")).To(gomega.Equal("100"))
			gomega.Expect(w.Header().Get("X-RateLimit-Window")).To(gomega.Equal("60"))
			gomega.Expect(w.Header().Get("X-RateLimit-Type")).To(gomega.Equal("general"))

			// Remaining should be a valid number
			remaining := w.Header().Get("X-RateLimit-Remaining")
			_, err := strconv.Atoi(remaining)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			// Reset should be a valid timestamp
			reset := w.Header().Get("X-RateLimit-Reset")
			_, err = strconv.ParseInt(reset, 10, 64)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		})

		ginkgo.It("should isolate rate limits per client", func() {
			// Client 1 exhausts their auth rate limit
			for i := 0; i < 3; i++ {
				req, _ := http.NewRequest("POST", "/api/v1/users", nil)
				req.Header.Set("X-Forwarded-For", "203.0.113.5")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if i < 2 {
					gomega.Expect(w.Code).To(gomega.Equal(http.StatusOK))
				} else {
					gomega.Expect(w.Code).To(gomega.Equal(http.StatusTooManyRequests))
				}
			}

			// Client 2 should still be able to make requests
			req, _ := http.NewRequest("POST", "/api/v1/users", nil)
			req.Header.Set("X-Forwarded-For", "203.0.113.6")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			gomega.Expect(w.Code).To(gomega.Equal(http.StatusOK))
		})
	})

	ginkgo.Describe("Convenience Functions", func() {
		ginkgo.It("should create middleware with provided logger", func() {
			middleware := WithRateLimit(logger)
			gomega.Expect(middleware).ToNot(gomega.BeNil())
		})
	})
})
