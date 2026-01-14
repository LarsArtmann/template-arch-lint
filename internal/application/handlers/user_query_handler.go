package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	pkgerrors "github.com/LarsArtmann/template-arch-lint/pkg/errors"
)

// UserQueryHandler handles read operations for users using CQRS pattern.
type UserQueryHandler struct {
	userQueryService services.UserQueryService
}

// NewUserQueryHandler creates a new user query handler.
func NewUserQueryHandler(userQueryService services.UserQueryService) *UserQueryHandler {
	return &UserQueryHandler{
		userQueryService: userQueryService,
	}
}

// GetUser retrieves a user by ID using query service.
func (h *UserQueryHandler) GetUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := values.NewUserID(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})

		return
	}

	user, err := h.userQueryService.GetUser(c.Request.Context(), userID)
	if err != nil {
		_, isNotFound := pkgerrors.AsNotFoundError(err)
		if isNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})

			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// ListUsers retrieves all users using query service.
func (h *UserQueryHandler) ListUsers(c *gin.Context) {
	users, err := h.userQueryService.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

// SearchUsers searches for users by email using query service.
func (h *UserQueryHandler) SearchUsers(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email query parameter is required"})

		return
	}

	user, err := h.userQueryService.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		_, isNotFound := pkgerrors.AsNotFoundError(err)
		if isNotFound {
			c.JSON(http.StatusOK, gin.H{"data": []any{}})

			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": []*entities.User{user}})
}

// GetUsersByDomain retrieves users by email domain using query service.
func (h *UserQueryHandler) GetUsersByDomain(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain parameter is required"})

		return
	}

	users, err := h.userQueryService.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})

		return
	}

	// Filter users by email domain
	filteredUsers := lo.Filter(users, func(user *entities.User, _ int) bool {
		userEmail := user.GetEmail().String()
		// Simple domain extraction - can be improved with proper regex
		if strings.Contains(userEmail, "@") {
			parts := strings.Split(userEmail, "@")

			return len(parts) == 2 && parts[1] == domain
		}

		return false
	})

	c.JSON(http.StatusOK, gin.H{"data": filteredUsers})
}

// GetUserStats retrieves user statistics using query service.
func (h *UserQueryHandler) GetUserStats(c *gin.Context) {
	stats, err := h.userQueryService.GetUserStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user statistics"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetActiveUsers retrieves active users using query service.
func (h *UserQueryHandler) GetActiveUsers(c *gin.Context) {
	// Use filters to get active users
	users, err := h.userQueryService.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve active users"})

		return
	}

	// Filter active users manually for now
	activeUsers := lo.Filter(users, func(user *entities.User, _ int) bool {
		return true // All users are considered active for now
	})

	c.JSON(http.StatusOK, gin.H{"data": activeUsers})
}

// GetUsersWithPagination retrieves users with pagination using query service.
func (h *UserQueryHandler) GetUsersWithPagination(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	users, err := h.userQueryService.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})

		return
	}

	// Apply pagination
	total := len(users)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		c.JSON(http.StatusOK, gin.H{
			"data":       []*entities.User{},
			"pagination": gin.H{"page": page, "limit": limit, "total": total},
		})

		return
	}

	if end > total {
		end = total
	}

	paginatedUsers := users[start:end]

	c.JSON(http.StatusOK, gin.H{
		"data":       paginatedUsers,
		"pagination": gin.H{"page": page, "limit": limit, "total": total},
	})
}
