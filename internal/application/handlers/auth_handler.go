// Package handlers provides HTTP request handlers.
package handlers

import (
	"net/http"

	httputil "github.com/LarsArtmann/template-arch-lint/internal/application/http"
	"github.com/LarsArtmann/template-arch-lint/internal/application/middleware"
	"github.com/LarsArtmann/template-arch-lint/internal/application/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	userRepo   repositories.UserRepository
	jwtService *services.JWTService
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(userRepo repositories.UserRepository, jwtService *services.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// LoginRequest represents the login request payload.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username,omitempty"`
}

// LoginResponse represents the login response.
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    int64        `json:"expires_at"`
	User         UserResponse `json:"user"`
}

// UserResponse represents user data in responses.
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// RefreshRequest represents the refresh token request payload.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshResponse represents the refresh token response.
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

// Login authenticates a user and returns JWT tokens.
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", "validation", nil)
		return
	}

	// For this demo, we'll find user by email or username
	// In a real app, you'd validate credentials (password, etc.)
	var user *entities.User
	var err error

	if req.Email != "" {
		_, emailErr := values.NewEmail(req.Email)
		if emailErr != nil {
			httputil.RespondError(c, http.StatusBadRequest, "INVALID_EMAIL", "Invalid email format", "validation", nil)
			return
		}
		user, err = h.userRepo.FindByEmail(c.Request.Context(), req.Email)
	} else if req.Username != "" {
		_, usernameErr := values.NewUserName(req.Username)
		if usernameErr != nil {
			httputil.RespondError(c, http.StatusBadRequest, "INVALID_USERNAME", "Invalid username format", "validation", nil)
			return
		}
		user, err = h.userRepo.FindByUsername(c.Request.Context(), req.Username)
	} else {
		httputil.RespondError(c, http.StatusBadRequest, "MISSING_CREDENTIALS", "Email or username required", "validation", nil)
		return
	}

	if err != nil {
		httputil.RespondError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid credentials", "authentication", nil)
		return
	}

	// Generate JWT token pair
	tokenPair, err := h.jwtService.GenerateTokenPair(user)
	if err != nil {
		httputil.RespondError(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate tokens", "internal", nil)
		return
	}

	response := LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		User: UserResponse{
			ID:       user.ID.String(),
			Username: user.Name,
			Email:    user.Email,
		},
	}

	httputil.RespondOK(c, response, "Login successful")
}

// RefreshToken generates a new access token using a refresh token.
// POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", "validation", nil)
		return
	}

	// Validate refresh token and extract claims
	claims, err := h.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		httputil.RespondError(c, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "Invalid refresh token", "authentication", nil)
		return
	}

	if claims.Type != "refresh" {
		httputil.RespondError(c, http.StatusUnauthorized, "INVALID_TOKEN_TYPE", "Token is not a refresh token", "authentication", nil)
		return
	}

	// Get user from database
	userID, err := values.NewUserID(claims.UserID)
	if err != nil {
		httputil.RespondError(c, http.StatusUnauthorized, "INVALID_USER_ID", "Invalid user ID in token", "authentication", nil)
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		httputil.RespondError(c, http.StatusUnauthorized, "USER_NOT_FOUND", "User not found", "authentication", nil)
		return
	}

	// Generate new access token
	newAccessToken, err := h.jwtService.RefreshAccessToken(req.RefreshToken, user)
	if err != nil {
		httputil.RespondError(c, http.StatusInternalServerError, "TOKEN_REFRESH_FAILED", "Failed to refresh token", "internal", nil)
		return
	}

	response := RefreshResponse{
		AccessToken: newAccessToken,
		ExpiresAt:   claims.ExpiresAt.Unix(),
	}

	httputil.RespondOK(c, response, "Token refreshed successfully")
}

// Me returns the current authenticated user's information.
// GET /api/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		httputil.RespondError(c, http.StatusUnauthorized, "NOT_AUTHENTICATED", "User not authenticated", "authentication", nil)
		return
	}

	id, err := values.NewUserID(userID)
	if err != nil {
		httputil.RespondError(c, http.StatusUnauthorized, "INVALID_USER_ID", "Invalid user ID", "authentication", nil)
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		httputil.RespondNotFound(c, "user", userID)
		return
	}

	response := UserResponse{
		ID:       user.ID.String(),
		Username: user.Name,
		Email:    user.Email,
	}

	httputil.RespondOK(c, response, "User information retrieved")
}
