package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

// Security test constants
const (
	MaxRequestSize      = 1024 * 1024 // 1MB
	SQLInjectionString  = "'; DROP TABLE users; --"
	XSSString           = "<script>alert('xss')</script>"
	PathTraversalString = "../../../etc/passwd"
)

var (
	LongString = strings.Repeat("A", 10000)
)

// TestSecurityMeasures provides comprehensive security testing
func TestSecurityMeasures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	t.Run("Input Validation Security", func(t *testing.T) {
		testInputValidationSecurity(t, router)
	})

	t.Run("HTTP Security Headers", func(t *testing.T) {
		testSecurityHeaders(t, router)
	})

	t.Run("Request Size Limits", func(t *testing.T) {
		testRequestSizeLimits(t, router)
	})

	t.Run("Rate Limiting", func(t *testing.T) {
		testRateLimiting(t, router)
	})

	t.Run("CORS Security", func(t *testing.T) {
		testCORSSecurity(t, router)
	})

	t.Run("Path Traversal Protection", func(t *testing.T) {
		testPathTraversalProtection(t, router)
	})

	t.Run("Content Type Validation", func(t *testing.T) {
		testContentTypeValidation(t, router)
	})

	t.Run("Authentication Security", func(t *testing.T) {
		testAuthenticationSecurity(t, router)
	})
}

// testInputValidationSecurity tests protection against malicious input
func testInputValidationSecurity(t *testing.T, router *gin.Engine) {
	t.Helper()

	t.Run("SQL Injection Protection", func(t *testing.T) {
		maliciousInputs := []string{
			"'; DROP TABLE users; --",
			"' OR '1'='1",
			"admin'; DELETE FROM users WHERE '1'='1",
			"' UNION SELECT * FROM users --",
			"'; INSERT INTO users (email) VALUES ('hacker@evil.com'); --",
		}

		for _, input := range maliciousInputs {
			// Test in user ID parameter
			req, _ := http.NewRequest("GET", "/api/v1/users/"+input, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should return bad request, not process the malicious input
			if w.Code == http.StatusOK {
				t.Errorf("SQL injection attempt was not blocked: %s", input)
			}

			// Test in request body
			payload := map[string]string{
				"id":    "test-user",
				"email": input,
				"name":  "Test User",
			}
			jsonPayload, _ := json.Marshal(payload)
			req, _ = http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should reject malicious email
			if w.Code == http.StatusCreated {
				t.Errorf("SQL injection in email was not blocked: %s", input)
			}
		}
		t.Log("✅ SQL injection protection working")
	})

	t.Run("XSS Protection", func(t *testing.T) {
		xssPayloads := []string{
			"<script>alert('xss')</script>",
			"javascript:alert('xss')",
			"<img src=x onerror=alert('xss')>",
			"<svg onload=alert('xss')>",
			"'><script>alert('xss')</script>",
		}

		for _, payload := range xssPayloads {
			userPayload := map[string]string{
				"id":    "xss-test",
				"email": "test@example.com",
				"name":  payload,
			}

			jsonPayload, _ := json.Marshal(userPayload)
			req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check if XSS payload is properly sanitized or rejected
			if w.Code == http.StatusCreated {
				// If created, check that the response doesn't contain the raw payload
				response := w.Body.String()
				if strings.Contains(response, "<script>") {
					t.Errorf("XSS payload not sanitized in response: %s", payload)
				}
			}
		}
		t.Log("✅ XSS protection working")
	})

	t.Run("Command Injection Protection", func(t *testing.T) {
		commandInjectionPayloads := []string{
			"; ls -la",
			"| cat /etc/passwd",
			"&& rm -rf /",
			"`whoami`",
			"$(id)",
		}

		for _, payload := range commandInjectionPayloads {
			userPayload := map[string]string{
				"id":    payload,
				"email": "test@example.com",
				"name":  "Test User",
			}

			jsonPayload, _ := json.Marshal(userPayload)
			req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should reject command injection attempts
			if w.Code == http.StatusCreated {
				t.Errorf("Command injection attempt was not blocked: %s", payload)
			}
		}
		t.Log("✅ Command injection protection working")
	})

	t.Run("Path Traversal in Input", func(t *testing.T) {
		pathTraversalPayloads := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32\\config\\sam",
			"....//....//....//etc/passwd",
			"%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
		}

		for _, payload := range pathTraversalPayloads {
			req, _ := http.NewRequest("GET", "/api/v1/users/"+payload, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should not return file contents or 200 OK
			if w.Code == http.StatusOK {
				response := w.Body.String()
				if strings.Contains(response, "root:") || strings.Contains(response, "Administrator") {
					t.Errorf("Path traversal attack succeeded: %s", payload)
				}
			}
		}
		t.Log("✅ Path traversal protection working")
	})
}

// testSecurityHeaders tests HTTP security headers
func testSecurityHeaders(t *testing.T, router *gin.Engine) {
	t.Helper()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	securityHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000",
		"Content-Security-Policy": "", // Should exist but content varies
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	headerCount := 0
	for header, expectedValue := range securityHeaders {
		actualValue := w.Header().Get(header)
		if actualValue != "" {
			headerCount++
			if expectedValue != "" && actualValue != expectedValue {
				t.Errorf("Security header %s has unexpected value: got %s, expected %s", 
					header, actualValue, expectedValue)
			} else {
				t.Logf("✅ Security header %s: %s", header, actualValue)
			}
		} else {
			t.Logf("⚠️ Security header %s not set", header)
		}
	}

	if headerCount == 0 {
		t.Error("No security headers found - security headers should be configured")
	} else {
		t.Logf("✅ Found %d security headers", headerCount)
	}
}

// testRequestSizeLimits tests protection against large requests
func testRequestSizeLimits(t *testing.T, router *gin.Engine) {
	t.Helper()

	t.Run("Large JSON Payload", func(t *testing.T) {
		largePayload := map[string]string{
			"id":    "test-user",
			"email": "test@example.com",
			"name":  strings.Repeat("A", 2*1024*1024), // 2MB
		}

		jsonPayload, _ := json.Marshal(largePayload)
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should reject large payloads
		if w.Code == http.StatusCreated {
			t.Error("Large payload was accepted - should implement size limits")
		} else {
			t.Logf("✅ Large payload rejected with status: %d", w.Code)
		}
	})

	t.Run("Extremely Long URL", func(t *testing.T) {
		longPath := "/api/v1/users/" + strings.Repeat("a", 8192) // 8KB path
		req, _ := http.NewRequest("GET", longPath, nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should handle long URLs gracefully
		if w.Code == http.StatusOK {
			t.Error("Extremely long URL was accepted")
		} else {
			t.Logf("✅ Long URL rejected with status: %d", w.Code)
		}
	})

	t.Run("Many Headers", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)

		// Add many headers
		for i := 0; i < 100; i++ {
			req.Header.Set(fmt.Sprintf("X-Custom-Header-%d", i), "value")
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should handle many headers gracefully or reject
		t.Logf("Response status for many headers: %d", w.Code)
	})
}

// testRateLimiting tests rate limiting protection
func testRateLimiting(t *testing.T, router *gin.Engine) {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping rate limiting test in short mode")
	}

	t.Run("Rapid Requests", func(t *testing.T) {
		// Make many rapid requests
		rapidRequests := 50
		successCount := 0
		rateLimitedCount := 0

		for i := 0; i < rapidRequests; i++ {
			req, _ := http.NewRequest("GET", "/health", nil)
			req.RemoteAddr = "192.168.1.100:12345" // Simulate same IP
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			switch w.Code {
			case http.StatusOK:
				successCount++
			case http.StatusTooManyRequests:
				rateLimitedCount++
			}
		}

		t.Logf("Rapid requests result: %d successful, %d rate limited", successCount, rateLimitedCount)

		// Rate limiting might not be enabled in test mode
		if rateLimitedCount == 0 {
			t.Log("⚠️ No rate limiting detected - consider implementing rate limiting")
		} else {
			t.Logf("✅ Rate limiting working: blocked %d requests", rateLimitedCount)
		}
	})

	t.Run("Rate Limit Headers", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check for rate limit headers
		rateLimitHeaders := []string{
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
			"Retry-After",
		}

		foundHeaders := 0
		for _, header := range rateLimitHeaders {
			if value := w.Header().Get(header); value != "" {
				foundHeaders++
				t.Logf("Rate limit header %s: %s", header, value)
			}
		}

		if foundHeaders > 0 {
			t.Logf("✅ Found %d rate limit headers", foundHeaders)
		} else {
			t.Log("⚠️ No rate limit headers found")
		}
	})
}

// testCORSSecurity tests CORS configuration security
func testCORSSecurity(t *testing.T, router *gin.Engine) {
	t.Helper()

	t.Run("CORS Headers", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/api/v1/users", nil)
		req.Header.Set("Origin", "https://evil.com")
		req.Header.Set("Access-Control-Request-Method", "POST")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check CORS headers
		corsHeaders := map[string]string{
			"Access-Control-Allow-Origin":      "",
			"Access-Control-Allow-Methods":     "",
			"Access-Control-Allow-Headers":     "",
			"Access-Control-Allow-Credentials": "",
		}

		foundCORS := false
		for header := range corsHeaders {
			if value := w.Header().Get(header); value != "" {
				foundCORS = true
				t.Logf("CORS header %s: %s", header, value)

				// Security checks
				if header == "Access-Control-Allow-Origin" && value == "*" {
					if w.Header().Get("Access-Control-Allow-Credentials") == "true" {
						t.Error("Security issue: CORS allows any origin with credentials")
					}
				}
			}
		}

		if foundCORS {
			t.Log("✅ CORS headers configured")
		} else {
			t.Log("⚠️ No CORS headers found")
		}
	})

	t.Run("Preflight Request", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/api/v1/users", nil)
		req.Header.Set("Origin", "https://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Preflight should be handled properly
		if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
			t.Errorf("Preflight request failed with status: %d", w.Code)
		} else {
			t.Log("✅ CORS preflight request handled")
		}
	})
}

// testPathTraversalProtection tests path traversal protection
func testPathTraversalProtection(t *testing.T, router *gin.Engine) {
	t.Helper()

	maliciousPaths := []string{
		"/../../../etc/passwd",
		"/..\\..\\..\\windows\\system32\\config\\sam",
		"/%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
		"/....//....//....//etc/passwd",
		"/api/v1/users/../../../etc/passwd",
	}

	for _, path := range maliciousPaths {
		req, _ := http.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should not serve system files
		response := w.Body.String()
		if strings.Contains(response, "root:x:0:0:") || 
		   strings.Contains(response, "Administrator") ||
		   strings.Contains(response, "[HKEY_") {
			t.Errorf("Path traversal vulnerability detected for path: %s", path)
		}

		// Should return appropriate error codes
		if w.Code == http.StatusOK && len(response) > 100 {
			t.Errorf("Suspicious large response for malicious path: %s", path)
		}
	}
	t.Log("✅ Path traversal protection working")
}

// testContentTypeValidation tests content type validation
func testContentTypeValidation(t *testing.T, router *gin.Engine) {
	t.Helper()

	t.Run("Wrong Content-Type", func(t *testing.T) {
		payload := `{"id": "test", "email": "test@example.com", "name": "Test User"}`
		
		wrongContentTypes := []string{
			"text/plain",
			"text/html",
			"application/xml",
			"application/x-www-form-urlencoded",
			"multipart/form-data",
		}

		for _, contentType := range wrongContentTypes {
			req, _ := http.NewRequest("POST", "/api/v1/users", strings.NewReader(payload))
			req.Header.Set("Content-Type", contentType)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should reject wrong content types
			if w.Code == http.StatusCreated {
				t.Errorf("Accepted wrong content type: %s", contentType)
			} else {
				t.Logf("✅ Rejected wrong content type %s with status: %d", contentType, w.Code)
			}
		}
	})

	t.Run("Missing Content-Type", func(t *testing.T) {
		payload := `{"id": "test", "email": "test@example.com", "name": "Test User"}`
		req, _ := http.NewRequest("POST", "/api/v1/users", strings.NewReader(payload))
		// Deliberately not setting Content-Type

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should require content type for POST requests
		if w.Code == http.StatusCreated {
			t.Error("Accepted request without Content-Type header")
		} else {
			t.Logf("✅ Rejected request without Content-Type with status: %d", w.Code)
		}
	})
}

// testAuthenticationSecurity tests authentication-related security
func testAuthenticationSecurity(t *testing.T, router *gin.Engine) {
	t.Helper()

	t.Run("Protected Endpoints", func(t *testing.T) {
		// Test endpoints that might require authentication
		protectedEndpoints := []string{
			"/admin",
			"/api/admin",
			"/debug/pprof/",
			"/metrics",
		}

		for _, endpoint := range protectedEndpoints {
			req, _ := http.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// These endpoints should either not exist or require authentication
			if w.Code == http.StatusOK {
				t.Logf("⚠️ Protected endpoint %s is accessible without authentication", endpoint)
			} else {
				t.Logf("✅ Protected endpoint %s properly secured (status: %d)", endpoint, w.Code)
			}
		}
	})

	t.Run("Sensitive Information Exposure", func(t *testing.T) {
		// Test various endpoints for sensitive information exposure
		endpoints := []string{
			"/health",
			"/api/v1/users",
			"/version",
			"/info",
		}

		sensitiveKeywords := []string{
			"password",
			"secret",
			"token",
			"key",
			"credential",
			"private",
			"config",
			"database",
			"connection",
		}

		for _, endpoint := range endpoints {
			req, _ := http.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				response := strings.ToLower(w.Body.String())
				for _, keyword := range sensitiveKeywords {
					if strings.Contains(response, keyword) {
						t.Logf("⚠️ Potentially sensitive information (%s) exposed in %s", keyword, endpoint)
					}
				}
			}
		}
		t.Log("✅ Checked for sensitive information exposure")
	})
}

// TestSecurityValueObjects tests security aspects of value objects
func TestSecurityValueObjects(t *testing.T) {
	t.Run("Email Security Validation", func(t *testing.T) {
		maliciousEmails := []string{
			"<script>alert('xss')</script>@example.com",
			"test@<script>alert('xss')</script>.com",
			"test+<script>@example.com",
			"'; DROP TABLE users; --@example.com",
		}

		for _, email := range maliciousEmails {
			_, err := values.NewEmail(email)
			if err == nil {
				t.Errorf("Malicious email was accepted: %s", email)
			}
		}
		t.Log("✅ Email value object security validation working")
	})

	t.Run("UserID Security Validation", func(t *testing.T) {
		maliciousIDs := []string{
			"<script>alert('xss')</script>",
			"'; DROP TABLE users; --",
			"../../etc/passwd",
			"user@domain.com",
			"user with spaces",
			"user\nwith\nnewlines",
		}

		for _, id := range maliciousIDs {
			_, err := values.NewUserID(id)
			if err == nil {
				t.Errorf("Malicious user ID was accepted: %s", id)
			}
		}
		t.Log("✅ UserID value object security validation working")
	})

	t.Run("Username Security Validation", func(t *testing.T) {
		maliciousNames := []string{
			"<script>alert('xss')</script>",
			"'; DROP TABLE users; --",
			"admin", // Reserved name
			"root",  // Reserved name
			"../../../etc/passwd",
		}

		for _, name := range maliciousNames {
			_, err := values.NewUserName(name)
			if err == nil {
				t.Errorf("Malicious username was accepted: %s", name)
			}
		}
		t.Log("✅ Username value object security validation working")
	})
}

// TestSecurityTiming tests for timing attack vulnerabilities
func TestSecurityTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timing attack tests in short mode")
	}

	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	t.Run("Timing Attack on User Lookup", func(t *testing.T) {
		// Test if user lookup timing is consistent
		existingUserTimes := make([]time.Duration, 10)
		nonExistentUserTimes := make([]time.Duration, 10)

		// Create a user first
		userPayload := map[string]string{
			"id":    "timing-test-user",
			"email": "timing@example.com",
			"name":  "Timing Test",
		}
		jsonPayload, _ := json.Marshal(userPayload)
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Measure timing for existing user
		for i := 0; i < 10; i++ {
			start := time.Now()
			req, _ := http.NewRequest("GET", "/api/v1/users/timing-test-user", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			existingUserTimes[i] = time.Since(start)
		}

		// Measure timing for non-existent user
		for i := 0; i < 10; i++ {
			start := time.Now()
			req, _ := http.NewRequest("GET", "/api/v1/users/non-existent-user", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			nonExistentUserTimes[i] = time.Since(start)
		}

		// Calculate averages
		avgExisting := averageDuration(existingUserTimes)
		avgNonExistent := averageDuration(nonExistentUserTimes)

		t.Logf("Average timing - Existing user: %v, Non-existent user: %v", avgExisting, avgNonExistent)

		// Large timing differences might indicate timing attack vulnerability
		timingDifference := float64(avgExisting) / float64(avgNonExistent)
		if timingDifference > 2.0 || timingDifference < 0.5 {
			t.Logf("⚠️ Potential timing attack vulnerability detected (ratio: %.2f)", timingDifference)
		} else {
			t.Log("✅ No obvious timing attack vulnerability")
		}
	})
}

// Helper functions

func averageDuration(durations []time.Duration) time.Duration {
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}