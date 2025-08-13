// Package handlers provides general HTTP testing utilities.
// These helpers provide reusable patterns for all HTTP handler testing scenarios.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/testhelpers/base"
)

// HTTPTestEnvironment provides a general HTTP testing environment.
type HTTPTestEnvironment struct {
	*base.GinkgoSuite
	router *gin.Engine
	logger *slog.Logger
	ctx    context.Context
}

// NewHTTPTestEnvironment creates a new HTTP test environment.
func NewHTTPTestEnvironment() *HTTPTestEnvironment {
	gin.SetMode(gin.TestMode)

	return &HTTPTestEnvironment{
		GinkgoSuite: base.NewGinkgoSuite(),
		router:      gin.New(),
		logger:      slog.New(slog.NewTextHandler(GinkgoWriter, nil)),
		ctx:         context.Background(),
	}
}

// Setup initializes the HTTP test environment.
func (env *HTTPTestEnvironment) Setup() {
	env.GinkgoSuite.Setup()
	env.ctx = env.GetContext()
	env.setupMiddleware()
}

// setupMiddleware configures common middleware for testing.
func (env *HTTPTestEnvironment) setupMiddleware() {
	// Add recovery middleware to catch panics in tests
	env.router.Use(gin.Recovery())

	// Add test-specific middleware if needed
	env.router.Use(func(c *gin.Context) {
		c.Header("X-Test-Environment", "true")
		c.Next()
	})
}

// GetRouter returns the Gin router for handler registration.
func (env *HTTPTestEnvironment) GetRouter() *gin.Engine {
	return env.router
}

// GetLogger returns the test logger.
func (env *HTTPTestEnvironment) GetLogger() *slog.Logger {
	return env.logger
}

// ExecuteRequest executes an HTTP request and returns the response recorder.
func (env *HTTPTestEnvironment) ExecuteRequest(req *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	env.router.ServeHTTP(recorder, req)
	return recorder
}

// GenericRequestBuilder provides a flexible HTTP request builder.
type GenericRequestBuilder struct {
	method      string
	path        string
	body        io.Reader
	headers     map[string]string
	queryParams map[string][]string
	cookies     []*http.Cookie
}

// NewGenericRequestBuilder creates a new generic request builder.
func NewGenericRequestBuilder() *GenericRequestBuilder {
	return &GenericRequestBuilder{
		headers:     make(map[string]string),
		queryParams: make(map[string][]string),
		cookies:     make([]*http.Cookie, 0),
	}
}

// Method sets the HTTP method.
func (b *GenericRequestBuilder) Method(method string) *GenericRequestBuilder {
	b.method = method
	return b
}

// Path sets the request path.
func (b *GenericRequestBuilder) Path(path string) *GenericRequestBuilder {
	b.path = path
	return b
}

// Body sets the request body from a reader.
func (b *GenericRequestBuilder) Body(body io.Reader) *GenericRequestBuilder {
	b.body = body
	return b
}

// JSONBody sets the request body as JSON.
func (b *GenericRequestBuilder) JSONBody(obj any) *GenericRequestBuilder {
	jsonBytes, err := json.Marshal(obj)
	Expect(err).ToNot(HaveOccurred())
	b.body = bytes.NewReader(jsonBytes)
	b.headers["Content-Type"] = "application/json"
	return b
}

// StringBody sets the request body from a string.
func (b *GenericRequestBuilder) StringBody(body string) *GenericRequestBuilder {
	b.body = strings.NewReader(body)
	return b
}

// Header sets a request header.
func (b *GenericRequestBuilder) Header(key, value string) *GenericRequestBuilder {
	b.headers[key] = value
	return b
}

// Headers sets multiple request headers.
func (b *GenericRequestBuilder) Headers(headers map[string]string) *GenericRequestBuilder {
	for key, value := range headers {
		b.headers[key] = value
	}
	return b
}

// QueryParam adds a query parameter.
func (b *GenericRequestBuilder) QueryParam(key, value string) *GenericRequestBuilder {
	if b.queryParams[key] == nil {
		b.queryParams[key] = make([]string, 0)
	}
	b.queryParams[key] = append(b.queryParams[key], value)
	return b
}

// QueryParams sets multiple query parameters.
func (b *GenericRequestBuilder) QueryParams(params map[string]string) *GenericRequestBuilder {
	for key, value := range params {
		b.QueryParam(key, value)
	}
	return b
}

// Cookie adds a cookie to the request.
func (b *GenericRequestBuilder) Cookie(cookie *http.Cookie) *GenericRequestBuilder {
	b.cookies = append(b.cookies, cookie)
	return b
}

// Build creates the HTTP request.
func (b *GenericRequestBuilder) Build() *http.Request {
	req, err := http.NewRequestWithContext(context.Background(), b.method, b.path, b.body)
	Expect(err).ToNot(HaveOccurred())

	// Set headers
	for key, value := range b.headers {
		req.Header.Set(key, value)
	}

	// Set query parameters
	if len(b.queryParams) > 0 {
		q := req.URL.Query()
		for key, values := range b.queryParams {
			for _, value := range values {
				q.Add(key, value)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	// Add cookies
	for _, cookie := range b.cookies {
		req.AddCookie(cookie)
	}

	return req
}

// GenericResponseAsserter provides comprehensive response assertions.
type GenericResponseAsserter struct {
	recorder *httptest.ResponseRecorder
}

// NewGenericResponseAsserter creates a response asserter.
func NewGenericResponseAsserter(recorder *httptest.ResponseRecorder) *GenericResponseAsserter {
	return &GenericResponseAsserter{recorder: recorder}
}

// Status verifies the HTTP status code.
func (a *GenericResponseAsserter) Status(expectedStatus int) *GenericResponseAsserter {
	Expect(a.recorder.Code).To(Equal(expectedStatus),
		"Expected status %d, got %d. Response body: %s",
		expectedStatus, a.recorder.Code, a.recorder.Body.String())
	return a
}

// StatusOK verifies the status is 200 OK.
func (a *GenericResponseAsserter) StatusOK() *GenericResponseAsserter {
	return a.Status(http.StatusOK)
}

// StatusCreated verifies the status is 201 Created.
func (a *GenericResponseAsserter) StatusCreated() *GenericResponseAsserter {
	return a.Status(http.StatusCreated)
}

// StatusBadRequest verifies the status is 400 Bad Request.
func (a *GenericResponseAsserter) StatusBadRequest() *GenericResponseAsserter {
	return a.Status(http.StatusBadRequest)
}

// StatusNotFound verifies the status is 404 Not Found.
func (a *GenericResponseAsserter) StatusNotFound() *GenericResponseAsserter {
	return a.Status(http.StatusNotFound)
}

// StatusInternalServerError verifies the status is 500 Internal Server Error.
func (a *GenericResponseAsserter) StatusInternalServerError() *GenericResponseAsserter {
	return a.Status(http.StatusInternalServerError)
}

// Header verifies a response header value.
func (a *GenericResponseAsserter) Header(key, expectedValue string) *GenericResponseAsserter {
	actualValue := a.recorder.Header().Get(key)
	Expect(actualValue).To(Equal(expectedValue),
		"Expected header %s to be %s, got %s", key, expectedValue, actualValue)
	return a
}

// HeaderContains verifies a response header contains a substring.
func (a *GenericResponseAsserter) HeaderContains(key, expectedSubstring string) *GenericResponseAsserter {
	actualValue := a.recorder.Header().Get(key)
	Expect(actualValue).To(ContainSubstring(expectedSubstring),
		"Expected header %s to contain %s, got %s", key, expectedSubstring, actualValue)
	return a
}

// ContentType verifies the Content-Type header.
func (a *GenericResponseAsserter) ContentType(expectedType string) *GenericResponseAsserter {
	return a.Header("Content-Type", expectedType)
}

// JSON unmarshals the response body as JSON into the target.
func (a *GenericResponseAsserter) JSON(target any) *GenericResponseAsserter {
	err := json.Unmarshal(a.recorder.Body.Bytes(), target)
	Expect(err).ToNot(HaveOccurred(), "Failed to unmarshal JSON response: %s", a.recorder.Body.String())
	return a
}

// Body verifies the response body content.
func (a *GenericResponseAsserter) Body(expectedBody string) *GenericResponseAsserter {
	actualBody := a.recorder.Body.String()
	Expect(actualBody).To(Equal(expectedBody))
	return a
}

// BodyContains verifies the response body contains a substring.
func (a *GenericResponseAsserter) BodyContains(expectedSubstring string) *GenericResponseAsserter {
	actualBody := a.recorder.Body.String()
	Expect(actualBody).To(ContainSubstring(expectedSubstring),
		"Expected body to contain %s, got: %s", expectedSubstring, actualBody)
	return a
}

// BodyEmpty verifies the response body is empty.
func (a *GenericResponseAsserter) BodyEmpty() *GenericResponseAsserter {
	Expect(a.recorder.Body.Len()).To(BeZero())
	return a
}

// BodyNotEmpty verifies the response body is not empty.
func (a *GenericResponseAsserter) BodyNotEmpty() *GenericResponseAsserter {
	Expect(a.recorder.Body.Len()).To(BeNumerically(">", 0))
	return a
}

// JSONField verifies a specific field in a JSON response.
func (a *GenericResponseAsserter) JSONField(fieldPath string, expectedValue any) *GenericResponseAsserter {
	var response map[string]any
	a.JSON(&response)

	// Simple field path support (e.g., "user.email")
	fields := strings.Split(fieldPath, ".")
	current := response

	for i, field := range fields {
		if i == len(fields)-1 {
			// Last field, check value
			Expect(current[field]).To(Equal(expectedValue),
				"Expected field %s to be %v, got %v", fieldPath, expectedValue, current[field])
		} else {
			// Navigate deeper
			next, ok := current[field].(map[string]any)
			Expect(ok).To(BeTrue(), "Field %s is not an object", field)
			current = next
		}
	}

	return a
}

// MockResponseWriter provides a mock HTTP response writer for testing.
type MockResponseWriter struct {
	StatusCode int
	Headers    http.Header
	Body       *bytes.Buffer
}

// NewMockResponseWriter creates a new mock response writer.
func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		StatusCode: 200,
		Headers:    make(http.Header),
		Body:       &bytes.Buffer{},
	}
}

// Header returns the header map for the response.
func (w *MockResponseWriter) Header() http.Header {
	return w.Headers
}

// Write writes data to the response body.
func (w *MockResponseWriter) Write(data []byte) (int, error) {
	return w.Body.Write(data)
}

// WriteHeader sets the status code for the response.
func (w *MockResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

// HTTPTestScenario provides a reusable test scenario framework.
type HTTPTestScenario struct {
	Name        string
	Method      string
	Path        string
	Body        any
	Headers     map[string]string
	QueryParams map[string]string

	ExpectedStatus  int
	ExpectedBody    string
	ExpectedHeaders map[string]string

	Setup    func(*HTTPTestEnvironment)
	Teardown func(*HTTPTestEnvironment)
	Validate func(*HTTPTestEnvironment, *GenericResponseAsserter)
}

// Execute runs the test scenario.
func (scenario *HTTPTestScenario) Execute(env *HTTPTestEnvironment) {
	scenario.runSetup(env)
	defer scenario.runTeardown(env)

	req := scenario.buildRequest()
	recorder := env.ExecuteRequest(req)
	asserter := NewGenericResponseAsserter(recorder)

	scenario.performAssertions(asserter)
	scenario.runCustomValidation(env, asserter)
}

func (scenario *HTTPTestScenario) runSetup(env *HTTPTestEnvironment) {
	if scenario.Setup != nil {
		scenario.Setup(env)
	}
}

func (scenario *HTTPTestScenario) runTeardown(env *HTTPTestEnvironment) {
	if scenario.Teardown != nil {
		scenario.Teardown(env)
	}
}

func (scenario *HTTPTestScenario) buildRequest() *http.Request {
	builder := NewGenericRequestBuilder().
		Method(scenario.Method).
		Path(scenario.Path)

	if scenario.Body != nil {
		builder.JSONBody(scenario.Body)
	}

	if scenario.Headers != nil {
		builder.Headers(scenario.Headers)
	}

	if scenario.QueryParams != nil {
		builder.QueryParams(scenario.QueryParams)
	}

	return builder.Build()
}

func (scenario *HTTPTestScenario) performAssertions(asserter *GenericResponseAsserter) {
	if scenario.ExpectedStatus != 0 {
		asserter.Status(scenario.ExpectedStatus)
	}

	if scenario.ExpectedBody != "" {
		asserter.Body(scenario.ExpectedBody)
	}

	if scenario.ExpectedHeaders != nil {
		for key, value := range scenario.ExpectedHeaders {
			asserter.Header(key, value)
		}
	}
}

func (scenario *HTTPTestScenario) runCustomValidation(env *HTTPTestEnvironment, asserter *GenericResponseAsserter) {
	if scenario.Validate != nil {
		scenario.Validate(env, asserter)
	}
}

// HTTPTestSuite provides a test suite for running multiple HTTP scenarios.
type HTTPTestSuite struct {
	*base.GinkgoSuite
	env       *HTTPTestEnvironment
	scenarios []HTTPTestScenario
}

// NewHTTPTestSuite creates a new HTTP test suite.
func NewHTTPTestSuite(env *HTTPTestEnvironment) *HTTPTestSuite {
	return &HTTPTestSuite{
		GinkgoSuite: base.NewGinkgoSuite(),
		env:         env,
		scenarios:   make([]HTTPTestScenario, 0),
	}
}

// AddScenario adds a test scenario to the suite.
func (suite *HTTPTestSuite) AddScenario(scenario HTTPTestScenario) {
	suite.scenarios = append(suite.scenarios, scenario)
}

// RunAllScenarios executes all test scenarios.
func (suite *HTTPTestSuite) RunAllScenarios() {
	for _, scenario := range suite.scenarios {
		scenario.Execute(suite.env)
	}
}

// Convenience functions

// SetupHTTPTest creates a basic HTTP test environment.
func SetupHTTPTest() *HTTPTestEnvironment {
	env := NewHTTPTestEnvironment()
	env.Setup()
	return env
}

// CreateJSONRequest creates a JSON request with the specified method, path, and body.
func CreateJSONRequest(method, path string, body any) *http.Request {
	return NewGenericRequestBuilder().
		Method(method).
		Path(path).
		JSONBody(body).
		Build()
}

// CreateSimpleRequest creates a simple request with method and path.
func CreateSimpleRequest(method, path string) *http.Request {
	return NewGenericRequestBuilder().
		Method(method).
		Path(path).
		Build()
}

// AssertJSONError verifies a JSON error response.
func AssertJSONError(recorder *httptest.ResponseRecorder, expectedStatus int, expectedMessage string) {
	asserter := NewGenericResponseAsserter(recorder)
	asserter.Status(expectedStatus)

	var errorResponse struct {
		Error string `json:"error"`
	}
	asserter.JSON(&errorResponse)
	Expect(errorResponse.Error).To(Equal(expectedMessage))
}
