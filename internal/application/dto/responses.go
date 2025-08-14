// Package dto contains Data Transfer Objects for API layer.
// These DTOs separate domain entities from API representations,
// providing type safety and consistent JSON serialization.
package dto

import (
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
)

// APIResponse represents a standardized API response wrapper.
type APIResponse[T any] struct {
	Success       bool      `json:"success"`
	Data          *T        `json:"data,omitempty"`
	Error         *Error    `json:"error,omitempty"`
	Message       string    `json:"message,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// Error represents a standardized API error response.
type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
	Type    string            `json:"type"`
}

// PaginatedResponse represents a paginated API response.
type PaginatedResponse[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
}

// Pagination contains pagination metadata.
type Pagination struct {
	Page       int `json:"page"`
	Size       int `json:"size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	EmailDomain string    `json:"email_domain"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

// UserListResponse represents a list of users.
type UserListResponse = PaginatedResponse[UserResponse]

// CreateUserRequest represents a user creation request.
type CreateUserRequest struct {
	ID    string `json:"id" binding:"required" validate:"required,min=3,max=50"`
	Email string `json:"email" binding:"required,email" validate:"required,email"`
	Name  string `json:"name" binding:"required" validate:"required,min=2,max=100"`
}

// UpdateUserRequest represents a user update request.
type UpdateUserRequest struct {
	Email string `json:"email" binding:"required,email" validate:"required,email"`
	Name  string `json:"name" binding:"required" validate:"required,min=2,max=100"`
}

// SuccessResponse creates a successful API response.
func SuccessResponse[T any](data T, message string, correlationID string) APIResponse[T] {
	return APIResponse[T]{
		Success:       true,
		Data:          &data,
		Message:       message,
		CorrelationID: correlationID,
		Timestamp:     time.Now(),
	}
}

// ErrorResponse creates an error API response.
func ErrorResponse(code, message, errorType, correlationID string, details map[string]string) APIResponse[any] {
	return APIResponse[any]{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
			Details: details,
			Type:    errorType,
		},
		CorrelationID: correlationID,
		Timestamp:     time.Now(),
	}
}

// ValidationErrorResponse creates a validation error response.
func ValidationErrorResponse(details map[string]string, correlationID string) APIResponse[any] {
	return ErrorResponse(
		"VALIDATION_ERROR",
		"Request validation failed",
		"validation",
		correlationID,
		details,
	)
}

// NotFoundErrorResponse creates a not found error response.
func NotFoundErrorResponse(resource, id, correlationID string) APIResponse[any] {
	return ErrorResponse(
		"NOT_FOUND",
		resource+" not found",
		"not_found",
		correlationID,
		map[string]string{"resource": resource, "id": id},
	)
}

// InternalErrorResponse creates an internal server error response.
func InternalErrorResponse(correlationID string) APIResponse[any] {
	return ErrorResponse(
		"INTERNAL_ERROR",
		"An internal error occurred",
		"internal",
		correlationID,
		nil,
	)
}

// ToUserResponse converts a domain User entity to API UserResponse.
func ToUserResponse(user *entities.User) UserResponse {
	return UserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		Name:        user.Name,
		EmailDomain: user.EmailDomain(),
		Created:     user.Created,
		Modified:    user.Modified,
	}
}

// ToUserListResponse converts a slice of domain User entities to API UserListResponse.
func ToUserListResponse(users []*entities.User, page, size int) UserListResponse {
	items := make([]UserResponse, len(users))
	for i, user := range users {
		items[i] = ToUserResponse(user)
	}

	total := len(users)
	totalPages := (total + size - 1) / size // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	return UserListResponse{
		Items: items,
		Pagination: Pagination{
			Page:       page,
			Size:       size,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}
