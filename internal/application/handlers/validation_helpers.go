// Package handlers provides HTTP request handlers for the application layer
package handlers

import (
	"net/http"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/gin-gonic/gin"
)

// ValidationHelper provides common validation patterns for handlers.
type ValidationHelper struct{}

// NewValidationHelper creates a new validation helper.
func NewValidationHelper() *ValidationHelper {
	return &ValidationHelper{}
}

// ValidateUserID validates a user ID and returns a standardized error response if invalid.
func (v *ValidationHelper) ValidateUserID(c *gin.Context, userIDStr string) (values.UserID, bool) {
	userID, err := values.NewUserID(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID format",
			"details": err.Error(),
		})
		return userID, false
	}
	return userID, true
}

// ValidateEmail validates an email and returns a standardized error response if invalid.
func (v *ValidationHelper) ValidateEmail(c *gin.Context, emailStr string) (values.Email, bool) {
	email, err := values.NewEmail(emailStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid email format",
			"details": err.Error(),
		})
		return email, false
	}
	return email, true
}

// ValidateUserName validates a username and returns a standardized error response if invalid.
func (v *ValidationHelper) ValidateUserName(c *gin.Context, nameStr string) (values.UserName, bool) {
	name, err := values.NewUserName(nameStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid username format",
			"details": err.Error(),
		})
		return name, false
	}
	return name, true
}

// SendValidationError sends a standardized validation error response.
func (v *ValidationHelper) SendValidationError(c *gin.Context, field, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":   "Validation failed",
		"field":   field,
		"message": message,
	})
}

// SendInternalError sends a standardized internal error response.
func (v *ValidationHelper) SendInternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   "Internal server error",
		"message": message,
	})
}
