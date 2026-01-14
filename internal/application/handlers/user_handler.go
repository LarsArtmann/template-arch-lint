package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

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
	var req struct {
		Email string `binding:"required,email"         json:"email"`
		Name  string `binding:"required,min=2,max=100" json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("Invalid request format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request_format",
			"message": err.Error(),
		})

		return
	}

	// Generate new user ID
	userID, err := values.NewUserID(generateUserID())
	if err != nil {
		log.Error("Failed to generate user ID", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "user_id_generation_failed",
			"message": "Failed to generate user ID",
		})

		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), userID, req.Email, req.Name)
	if err != nil {
		log.Error("Failed to create user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "user_creation_failed",
			"message": "Failed to create user",
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        user.ID.String(),
		"email":     user.GetEmail().String(),
		"name":      user.GetUserName().String(),
		"createdAt": user.GetCreatedAt(),
		"updatedAt": user.GetUpdatedAt(),
	})
}

// GetUser handles user retrieval requests.
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := values.NewUserID(idStr)
	if err != nil {
		log.Error("Invalid user ID format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_user_id",
			"message": "Invalid user ID format",
		})

		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		log.Error("Failed to get user", "error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "user_not_found",
			"message": "User not found",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        user.ID.String(),
		"email":     user.GetEmail().String(),
		"name":      user.GetUserName().String(),
		"createdAt": user.GetCreatedAt(),
		"updatedAt": user.GetUpdatedAt(),
	})
}

// UpdateUser handles user update requests.
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := values.NewUserID(idStr)
	if err != nil {
		log.Error("Invalid user ID format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_user_id",
			"message": "Invalid user ID format",
		})

		return
	}

	var req struct {
		Email string `binding:"omitempty,email"         json:"email"`
		Name  string `binding:"omitempty,min=2,max=100" json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("Invalid request format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request_format",
			"message": err.Error(),
		})

		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userID, req.Email, req.Name)
	if err != nil {
		log.Error("Failed to update user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "user_update_failed",
			"message": "Failed to update user",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        user.ID.String(),
		"email":     user.GetEmail().String(),
		"name":      user.GetUserName().String(),
		"createdAt": user.GetCreatedAt(),
		"updatedAt": user.GetUpdatedAt(),
	})
}

// DeleteUser handles user deletion requests.
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := values.NewUserID(idStr)
	if err != nil {
		log.Error("Invalid user ID format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_user_id",
			"message": "Invalid user ID format",
		})

		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		log.Error("Failed to delete user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "user_deletion_failed",
			"message": "Failed to delete user",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}
