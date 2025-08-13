// Package handlers provides test helpers for HTTP handler testing.
// These helpers eliminate repetitive HTTP testing setup and assertion patterns.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/application/handlers"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/testhelpers/base"
	repositoryHelpers "github.com/LarsArtmann/template-arch-lint/internal/testhelpers/infrastructure/repositories"
)

// UserHandlerTestEnvironment provides a complete testing environment for UserHandler.
type UserHandlerTestEnvironment struct {
	*base.GinkgoSuite
	handler     *handlers.UserHandler
	mockRepo    *repositoryHelpers.MockUserRepository
	userService *services.UserService
	router      *gin.Engine
	logger      *slog.Logger
	ctx         context.Context
}

// NewUserHandlerTestEnvironment creates a new handler test environment.
func NewUserHandlerTestEnvironment() *UserHandlerTestEnvironment {
	gin.SetMode(gin.TestMode)

	mockRepo := repositoryHelpers.SetupMockRepository()
	logger := slog.New(slog.NewTextHandler(GinkgoWriter, nil))
	userService := services.NewUserService(mockRepo)
	handler := handlers.NewUserHandler(userService, logger)

	return &UserHandlerTestEnvironment{
		GinkgoSuite: base.NewGinkgoSuite(),
		handler:     handler,
		mockRepo:    mockRepo,
		userService: userService,
		logger:      logger,
		ctx:         context.Background(),
	}
}

// Setup initializes the test environment.
func (env *UserHandlerTestEnvironment) Setup() {
	env.GinkgoSuite.Setup()
	env.ctx = env.GetContext()
	env.setupRouter()
	env.resetMocks()
}

// setupRouter configures the Gin router with all handler routes.
func (env *UserHandlerTestEnvironment) setupRouter() {
	env.router = gin.New()
	env.router.POST("/users", env.handler.CreateUser)
	env.router.GET("/users/:id", env.handler.GetUser)
	env.router.PUT("/users/:id", env.handler.UpdateUser)
	env.router.DELETE("/users/:id", env.handler.DeleteUser)
	env.router.GET("/users", env.handler.ListUsers)
	env.router.GET("/users-stats", env.handler.GetUserStats)
	env.router.GET("/active-users", env.handler.GetActiveUsers)
	env.router.GET("/user-emails", env.handler.GetUserEmails)
}

// resetMocks clears all mock state for fresh testing.
func (env *UserHandlerTestEnvironment) resetMocks() {
	env.mockRepo.Clear()
	env.mockRepo.ResetCallLog()
	env.mockRepo.SetShouldError(false, "")
}

// Reset clears the environment state between tests.
func (env *UserHandlerTestEnvironment) Reset() {
	env.GinkgoSuite.Reset()
	env.resetMocks()
}

// PreloadTestUser adds a test user to the mock repository.
func (env *UserHandlerTestEnvironment) PreloadTestUser(user *entities.User) {
	env.mockRepo.PreloadUsers(user)
}

// PreloadMultipleUsers adds multiple test users to the mock repository.
func (env *UserHandlerTestEnvironment) PreloadMultipleUsers(users ...*entities.User) {
	env.mockRepo.PreloadUsers(users...)
}

// SetRepositoryError configures the mock repository to return errors.
func (env *UserHandlerTestEnvironment) SetRepositoryError(shouldError bool, errorType string) {
	env.mockRepo.SetShouldError(shouldError, errorType)
}

// HTTPRequestBuilder provides a fluent API for building HTTP requests.
type HTTPRequestBuilder struct {
	method      string
	path        string
	body        any
	headers     map[string]string
	queryParams map[string]string
}

// NewHTTPRequestBuilder creates a new request builder.
func NewHTTPRequestBuilder() *HTTPRequestBuilder {
	return &HTTPRequestBuilder{
		headers:     make(map[string]string),
		queryParams: make(map[string]string),
	}
}

// POST sets the request method to POST.
func (b *HTTPRequestBuilder) POST(path string) *HTTPRequestBuilder {
	b.method = "POST"
	b.path = path
	return b
}

// GET sets the request method to GET.
func (b *HTTPRequestBuilder) GET(path string) *HTTPRequestBuilder {
	b.method = "GET"
	b.path = path
	return b
}

// PUT sets the request method to PUT.
func (b *HTTPRequestBuilder) PUT(path string) *HTTPRequestBuilder {
	b.method = "PUT"
	b.path = path
	return b
}

// DELETE sets the request method to DELETE.
func (b *HTTPRequestBuilder) DELETE(path string) *HTTPRequestBuilder {
	b.method = "DELETE"
	b.path = path
	return b
}

// WithJSONBody sets the request body as JSON.
func (b *HTTPRequestBuilder) WithJSONBody(body any) *HTTPRequestBuilder {
	b.body = body
	b.headers["Content-Type"] = "application/json"
	return b
}

// WithHeader adds a header to the request.
func (b *HTTPRequestBuilder) WithHeader(key, value string) *HTTPRequestBuilder {
	b.headers[key] = value
	return b
}

// WithQueryParam adds a query parameter to the request.
func (b *HTTPRequestBuilder) WithQueryParam(key, value string) *HTTPRequestBuilder {
	b.queryParams[key] = value
	return b
}

// Build creates the HTTP request.
func (b *HTTPRequestBuilder) Build() *http.Request {
	var bodyReader *bytes.Reader

	if b.body != nil {
		jsonBody, err := json.Marshal(b.body)
		Expect(err).ToNot(HaveOccurred())
		bodyReader = bytes.NewReader(jsonBody)
	}

	var req *http.Request
	var err error

	if bodyReader != nil {
		req, err = http.NewRequestWithContext(context.Background(), b.method, b.path, bodyReader)
	} else {
		req, err = http.NewRequestWithContext(context.Background(), b.method, b.path, nil)
	}

	Expect(err).ToNot(HaveOccurred())

	// Add headers
	for key, value := range b.headers {
		req.Header.Set(key, value)
	}

	// Add query parameters
	if len(b.queryParams) > 0 {
		q := req.URL.Query()
		for key, value := range b.queryParams {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	return req
}

// HTTPResponseAsserter provides utilities for asserting HTTP responses.
type HTTPResponseAsserter struct {
	recorder *httptest.ResponseRecorder
}

// NewHTTPResponseAsserter creates a response asserter.
func NewHTTPResponseAsserter(recorder *httptest.ResponseRecorder) *HTTPResponseAsserter {
	return &HTTPResponseAsserter{recorder: recorder}
}

// AssertStatusCode verifies the HTTP status code.
func (a *HTTPResponseAsserter) AssertStatusCode(expectedStatus int) *HTTPResponseAsserter {
	Expect(a.recorder.Code).To(Equal(expectedStatus))
	return a
}

// AssertStatusOK verifies the status is 200 OK.
func (a *HTTPResponseAsserter) AssertStatusOK() *HTTPResponseAsserter {
	return a.AssertStatusCode(http.StatusOK)
}

// AssertStatusCreated verifies the status is 201 Created.
func (a *HTTPResponseAsserter) AssertStatusCreated() *HTTPResponseAsserter {
	return a.AssertStatusCode(http.StatusCreated)
}

// AssertStatusBadRequest verifies the status is 400 Bad Request.
func (a *HTTPResponseAsserter) AssertStatusBadRequest() *HTTPResponseAsserter {
	return a.AssertStatusCode(http.StatusBadRequest)
}

// AssertStatusNotFound verifies the status is 404 Not Found.
func (a *HTTPResponseAsserter) AssertStatusNotFound() *HTTPResponseAsserter {
	return a.AssertStatusCode(http.StatusNotFound)
}

// AssertJSONResponse unmarshals the response body into the provided struct.
func (a *HTTPResponseAsserter) AssertJSONResponse(target any) *HTTPResponseAsserter {
	err := json.Unmarshal(a.recorder.Body.Bytes(), target)
	Expect(err).ToNot(HaveOccurred())
	return a
}

// AssertErrorResponse verifies the response contains an error message.
func (a *HTTPResponseAsserter) AssertErrorResponse(expectedMessage string) *HTTPResponseAsserter {
	var errorResponse ErrorResponse
	a.AssertJSONResponse(&errorResponse)
	Expect(errorResponse.Error).To(Equal(expectedMessage))
	return a
}

// AssertUserResponse verifies the response contains a user.
func (a *HTTPResponseAsserter) AssertUserResponse(expectedUser *entities.User) *HTTPResponseAsserter {
	var user entities.User
	a.AssertJSONResponse(&user)
	Expect(user.ID).To(Equal(expectedUser.ID))
	Expect(user.Email).To(Equal(expectedUser.Email))
	Expect(user.Name).To(Equal(expectedUser.Name))
	return a
}

// AssertUserListResponse verifies the response contains a user list.
func (a *HTTPResponseAsserter) AssertUserListResponse(expectedCount int) *HTTPResponseAsserter {
	var response ListUsersResponse
	a.AssertJSONResponse(&response)
	Expect(response.Total).To(Equal(expectedCount))
	Expect(response.Users).To(HaveLen(expectedCount))
	return a
}

// ErrorResponse represents an error response structure for testing HTTP handlers.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// UserResponse represents a user data structure in HTTP responses for testing.
type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// SuccessResponse represents a success message structure in HTTP responses for testing.
type SuccessResponse struct {
	Message string `json:"message"`
}

// ListUsersResponse represents a user list response structure for testing HTTP handlers.
type ListUsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
}

// CreateUserRequest represents a user creation request structure for testing HTTP handlers.
type CreateUserRequest struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UpdateUserRequest represents a user update request structure for testing HTTP handlers.
type UpdateUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Convenience methods for the test environment

// ExecuteRequest executes an HTTP request and returns the response recorder.
func (env *UserHandlerTestEnvironment) ExecuteRequest(req *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	env.router.ServeHTTP(recorder, req)
	return recorder
}

// POST executes a POST request with JSON body.
func (env *UserHandlerTestEnvironment) POST(path string, body any) *HTTPResponseAsserter {
	req := NewHTTPRequestBuilder().POST(path).WithJSONBody(body).Build()
	recorder := env.ExecuteRequest(req)
	return NewHTTPResponseAsserter(recorder)
}

// GET executes a GET request.
func (env *UserHandlerTestEnvironment) GET(path string) *HTTPResponseAsserter {
	req := NewHTTPRequestBuilder().GET(path).Build()
	recorder := env.ExecuteRequest(req)
	return NewHTTPResponseAsserter(recorder)
}

// PUT executes a PUT request with JSON body.
func (env *UserHandlerTestEnvironment) PUT(path string, body any) *HTTPResponseAsserter {
	req := NewHTTPRequestBuilder().PUT(path).WithJSONBody(body).Build()
	recorder := env.ExecuteRequest(req)
	return NewHTTPResponseAsserter(recorder)
}

// DELETE executes a DELETE request.
func (env *UserHandlerTestEnvironment) DELETE(path string) *HTTPResponseAsserter {
	req := NewHTTPRequestBuilder().DELETE(path).Build()
	recorder := env.ExecuteRequest(req)
	return NewHTTPResponseAsserter(recorder)
}

// User-specific convenience methods

// CreateUserRequest creates a request for creating a user.
func (env *UserHandlerTestEnvironment) CreateUserRequest(id, email, name string) *HTTPResponseAsserter {
	request := CreateUserRequest{
		ID:    id,
		Email: email,
		Name:  name,
	}
	return env.POST("/users", request)
}

// CreateValidUserRequest creates a valid user creation request.
func (env *UserHandlerTestEnvironment) CreateValidUserRequest() *HTTPResponseAsserter {
	return env.CreateUserRequest("test-user-123", "test@example.com", "Test User")
}

// GetUserRequest creates a request for getting a user by ID.
func (env *UserHandlerTestEnvironment) GetUserRequest(userID string) *HTTPResponseAsserter {
	return env.GET(fmt.Sprintf("/users/%s", userID))
}

// UpdateUserRequest creates a request for updating a user.
func (env *UserHandlerTestEnvironment) UpdateUserRequest(userID, email, name string) *HTTPResponseAsserter {
	request := UpdateUserRequest{
		Email: email,
		Name:  name,
	}
	return env.PUT(fmt.Sprintf("/users/%s", userID), request)
}

// DeleteUserRequest creates a request for deleting a user.
func (env *UserHandlerTestEnvironment) DeleteUserRequest(userID string) *HTTPResponseAsserter {
	return env.DELETE(fmt.Sprintf("/users/%s", userID))
}

// ListUsersRequest creates a request for listing all users.
func (env *UserHandlerTestEnvironment) ListUsersRequest() *HTTPResponseAsserter {
	return env.GET("/users")
}

// UserStatsRequest creates a request for getting user statistics.
func (env *UserHandlerTestEnvironment) UserStatsRequest() *HTTPResponseAsserter {
	return env.GET("/users-stats")
}

// ActiveUsersRequest creates a request for getting active users.
func (env *UserHandlerTestEnvironment) ActiveUsersRequest() *HTTPResponseAsserter {
	return env.GET("/active-users")
}

// UserEmailsRequest creates a request for getting user emails.
func (env *UserHandlerTestEnvironment) UserEmailsRequest() *HTTPResponseAsserter {
	return env.GET("/user-emails")
}

// Test scenario helpers

// TestCreateUserSuccess tests successful user creation scenario.
func (env *UserHandlerTestEnvironment) TestCreateUserSuccess() {
	response := env.CreateValidUserRequest()
	response.AssertStatusCreated()

	var user entities.User
	response.AssertJSONResponse(&user)
	Expect(user.ID.String()).To(Equal("test-user-123"))
	Expect(user.Email).To(Equal("test@example.com"))
	Expect(user.Name).To(Equal("Test User"))
}

// TestCreateUserValidationError tests user creation with validation errors.
func (env *UserHandlerTestEnvironment) TestCreateUserValidationError(invalidField, invalidValue string) {
	var request CreateUserRequest

	switch invalidField {
	case "id":
		request = CreateUserRequest{ID: invalidValue, Email: "test@example.com", Name: "Test User"}
	case "email":
		request = CreateUserRequest{ID: "test-user-123", Email: invalidValue, Name: "Test User"}
	case "name":
		request = CreateUserRequest{ID: "test-user-123", Email: "test@example.com", Name: invalidValue}
	}

	response := env.POST("/users", request)
	response.AssertStatusBadRequest()
}

// TestGetUserSuccess tests successful user retrieval.
func (env *UserHandlerTestEnvironment) TestGetUserSuccess(user *entities.User) {
	env.PreloadTestUser(user)

	response := env.GetUserRequest(user.ID.String())
	response.AssertStatusOK().AssertUserResponse(user)
}

// TestGetUserNotFound tests user not found scenario.
func (env *UserHandlerTestEnvironment) TestGetUserNotFound() {
	response := env.GetUserRequest("nonexistent-user")
	response.AssertStatusNotFound()
}

// SetupUserHandlerTest creates and configures a new user handler test environment.
func SetupUserHandlerTest() *UserHandlerTestEnvironment {
	env := NewUserHandlerTestEnvironment()
	env.Setup()
	return env
}
