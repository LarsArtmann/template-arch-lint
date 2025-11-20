package dto

// CreateUserRequest represents a user creation request.
type CreateUserRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required,min=2,max=50"`
}

// UpdateUserRequest represents a user update request.
type UpdateUserRequest struct {
	Email string `json:"email" binding:"omitempty,email"`
	Name  string `json:"name" binding:"omitempty,min=2,max=50"`
}

// UserResponse represents a user response.
type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse represents a success response.
type SuccessResponse struct {
	Message string `json:"message"`
}

// UsersResponse represents a list of users response.
type UsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
}
