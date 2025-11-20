package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/LarsArtmann/template-arch-lint/internal/application/dto"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

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
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateUser handles user creation requests.
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("Invalid request format", "error", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request_format",
			Message: err.Error(),
		})
		return
	}

	// Generate new user ID
	userID, err := values.NewUserID(generateUserID())
	if err != nil {
		log.Error("Failed to generate user ID", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "user_id_generation_failed",
			Message: "Failed to generate user ID",
		})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), userID, req.Email, req.Name)
	if err != nil {
		log.Error("Failed to create user", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "user_creation_failed",
			Message: "Failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.UserResponse{
		ID:    user.ID.String(),
		Email: user.GetEmail().String(),
		Name:  user.GetUserName().String(),
	})
}

// GetUser handles user retrieval requests.
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := values.NewUserID(idStr)
	if err != nil {
		log.Error("Invalid user ID format", "error", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		log.Error("Failed to get user", "error", err)
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "user_not_found",
			Message: "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		ID:    user.ID.String(),
		Email: user.GetEmail().String(),
		Name:  user.GetUserName().String(),
	})
}

// UpdateUser handles user update requests.
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := values.NewUserID(idStr)
	if err != nil {
		log.Error("Invalid user ID format", "error", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("Invalid request format", "error", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request_format",
			Message: err.Error(),
		})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userID, req.Email, req.Name)
	if err != nil {
		log.Error("Failed to update user", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "user_update_failed",
			Message: "Failed to update user",
		})
		return
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		ID:    user.ID.String(),
		Email: user.GetEmail().String(),
		Name:  user.GetUserName().String(),
	})
}

// DeleteUser handles user deletion requests.
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := values.NewUserID(idStr)
	if err != nil {
		log.Error("Invalid user ID format", "error", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		log.Error("Failed to delete user", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "user_deletion_failed",
			Message: "Failed to delete user",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "User deleted successfully",
	})
}
