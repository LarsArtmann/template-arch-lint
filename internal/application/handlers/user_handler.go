package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"charm.land/log/v2"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/gin-gonic/gin"
)

// ID generation constants.
const userIDByteLength = 8

// UserHandler handles HTTP requests for user management.
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// generateUserID generates a new random user ID.
func generateUserID() string {
	bytes := make([]byte, userIDByteLength)
	_, _ = rand.Read(bytes)

	return hex.EncodeToString(bytes)
}

// errorResponse sends a standardized JSON error response.
func errorResponse(c *gin.Context, status int, errCode, message string) {
	c.JSON(status, gin.H{
		"error":   errCode,
		"message": message,
	})
}

// bindRequest binds and validates JSON request body.
func bindRequest[T any](c *gin.Context, req *T) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		log.Error("Invalid request format", "error", err)
		errorResponse(c, http.StatusBadRequest, "invalid_request_format", err.Error())
		return false
	}
	return true
}

// userToJSON converts a user entity to a JSON response map.
func userToJSON(user *entities.User) gin.H {
	return gin.H{
		"id":        user.ID.String(),
		"email":     user.GetEmail().String(),
		"name":      user.GetUserName().String(),
		"createdAt": user.GetCreatedAt(),
		"updatedAt": user.GetUpdatedAt(),
	}
}

// CreateUser handles user creation requests.
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req struct {
		Email string `binding:"required,email"         json:"email"`
		Name  string `binding:"required,min=2,max=100" json:"name"`
	}
	if !bindRequest(c, &req) {
		return
	}

	userID, err := values.NewUserID(generateUserID())
	if err != nil {
		log.Error("Failed to generate user ID", "error", err)
		errorResponse(c, http.StatusInternalServerError, "user_id_generation_failed", "Failed to generate user ID")
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), userID, req.Email, req.Name)
	if err != nil {
		log.Error("Failed to create user", "error", err)
		errorResponse(c, http.StatusInternalServerError, "user_creation_failed", "Failed to create user")
		return
	}

	c.JSON(http.StatusCreated, userToJSON(user))
}

// parseUserID extracts and validates user ID from URL parameter.
func parseUserID(c *gin.Context) (values.UserID, bool) {
	idStr := c.Param("id")
	userID, err := values.NewUserID(idStr)
	if err != nil {
		log.Error("Invalid user ID format", "error", err)
		errorResponse(c, http.StatusBadRequest, "invalid_user_id", "Invalid user ID format")
		return values.UserID{}, false
	}
	return userID, true
}

// GetUser handles user retrieval requests.
func (h *UserHandler) GetUser(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		log.Error("Failed to get user", "error", err)
		errorResponse(c, http.StatusNotFound, "user_not_found", "User not found")
		return
	}

	c.JSON(http.StatusOK, userToJSON(user))
}

// UpdateUser handles user update requests.
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	var req struct {
		Email string `binding:"omitempty,email"         json:"email"`
		Name  string `binding:"omitempty,min=2,max=100" json:"name"`
	}
	if !bindRequest(c, &req) {
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userID, req.Email, req.Name)
	if err != nil {
		log.Error("Failed to update user", "error", err)
		errorResponse(c, http.StatusInternalServerError, "user_update_failed", "Failed to update user")
		return
	}

	c.JSON(http.StatusOK, userToJSON(user))
}

// DeleteUser handles user deletion requests.
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	err := h.userService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		log.Error("Failed to delete user", "error", err)
		errorResponse(c, http.StatusInternalServerError, "user_deletion_failed", "Failed to delete user")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}
