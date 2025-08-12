package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/samber/lo"

	domainerrors "github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *services.UserService
	logger      *slog.Logger
}

// NewUserHandler creates a new UserHandler with dependency injection
func NewUserHandler(userService *services.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}


// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	ID    string `json:"id" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required"`
}

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required"`
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(c *gin.Context) {
	h.logger.Info("Creating user", "remote_addr", c.ClientIP())

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request payload", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Record validation success for request payload

	// Create user using service layer
	userID, err := values.NewUserID(req.ID)
	if err != nil {
		h.logger.Warn("Invalid user ID format", "error", err, "user_id", req.ID)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID format",
			"details": err.Error(),
		})
		return
	}

	// Record validation success for user ID

	// Validate email format
	if !strings.Contains(req.Email, "@") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
		})
		return
	}

	// Validate name
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name cannot be empty",
		})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), userID, req.Email, req.Name)
	if err != nil {
		h.logger.Error("Failed to create user", "error", err, "user_id", req.ID)

		if errors.Is(err, repositories.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "User or email already exists",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	// Record successful user creation

	h.logger.Info("User created successfully", "user_id", user.ID, "email", user.Email)
	c.JSON(http.StatusCreated, user)
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := values.NewUserID(idStr)
	if err != nil {
		h.logger.Warn("Invalid user ID format", "error", err, "user_id", idStr)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID format",
			"details": err.Error(),
		})
		return
	}

	h.logger.Debug("Getting user", "user_id", id)

	user, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get user", "error", err, "user_id", id)

		if errors.Is(err, repositories.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser updates an existing user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := values.NewUserID(idStr)
	if err != nil {
		h.logger.Warn("Invalid user ID format", "error", err, "user_id", idStr)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID format",
			"details": err.Error(),
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Updating user", "user_id", id)

	// Update user using service layer
	user, err := h.userService.UpdateUser(c.Request.Context(), id, req.Email, req.Name)
	if err != nil {
		h.logger.Error("Failed to update user", "error", err, "user_id", id)

		if errors.Is(err, repositories.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}

		if errors.Is(err, repositories.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Email already exists",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to update user",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("User updated successfully", "user_id", user.ID)
	c.JSON(http.StatusOK, user)
}

// DeleteUser removes a user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := values.NewUserID(idStr)
	if err != nil {
		h.logger.Warn("Invalid user ID format", "error", err, "user_id", idStr)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID format",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Deleting user", "user_id", id)

	// Delete user using service layer
	if err := h.userService.DeleteUser(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete user", "error", err, "user_id", id)

		if errors.Is(err, repositories.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete user",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("User deleted successfully", "user_id", id)
	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// ListUsers retrieves all users
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse optional query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	h.logger.Debug("Listing users", "limit", limit, "offset", offset)

	// Get all users from service layer
	users, err := h.userService.ListUsers(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list users", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve users",
			"details": err.Error(),
		})
		return
	}

	// Apply pagination
	total := len(users)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	paginatedUsers := users[start:end]

	h.logger.Info("Users listed successfully", "total", total, "returned", len(paginatedUsers))
	c.JSON(http.StatusOK, gin.H{
		"users":  paginatedUsers,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// handleError processes errors and returns appropriate HTTP responses using typed errors
func (h *UserHandler) handleError(c *gin.Context, err error, operation string) {
	h.logger.Error("Operation failed", "operation", operation, "error", err)

	// Handle domain errors with proper types
	if domainErr, ok := err.(domainerrors.DomainError); ok {
		c.JSON(domainErr.HTTPStatus(), gin.H{
			"error":   domainErr.Code(),
			"message": domainErr.Error(),
			"details": domainErr.Details(),
		})
		return
	}

	// Handle specific error types
	if validationErr, ok := domainerrors.AsValidationError(err); ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": validationErr.Error(),
			"field":   validationErr.Field(),
			"details": validationErr.Details(),
		})
		return
	}

	if notFoundErr, ok := domainerrors.AsNotFoundError(err); ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error":    "not_found",
			"message":  notFoundErr.Error(),
			"resource": notFoundErr.Resource(),
			"id":       notFoundErr.ID(),
		})
		return
	}

	if conflictErr, ok := domainerrors.AsConflictError(err); ok {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "conflict",
			"message": conflictErr.Error(),
			"details": conflictErr.Details(),
		})
		return
	}

	// Fallback for unknown errors
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   "internal_error",
		"message": "An internal server error occurred",
		"details": err.Error(),
	})
}

// GetUserStats demonstrates functional operations endpoint
func (h *UserHandler) GetUserStats(c *gin.Context) {
	h.logger.Debug("Getting user statistics")

	stats, err := h.userService.GetUserStats(c.Request.Context())
	if err != nil {
		h.handleError(c, err, "get_user_stats")
		return
	}

	h.logger.Info("User statistics retrieved successfully", "stats", stats)
	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GetActiveUsers demonstrates functional programming endpoint
func (h *UserHandler) GetActiveUsers(c *gin.Context) {
	h.logger.Debug("Getting active users")

	users, err := h.userService.FilterActiveUsers(c.Request.Context())
	if err != nil {
		h.handleError(c, err, "get_active_users")
		return
	}

	h.logger.Info("Active users retrieved successfully", "count", len(users))
	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// GetUserEmails demonstrates Result pattern endpoint
func (h *UserHandler) GetUserEmails(c *gin.Context) {
	h.logger.Debug("Getting user emails")

	result := h.userService.GetUserEmailsWithResult(c.Request.Context())
	if result.IsError() {
		h.handleError(c, result.Error(), "get_user_emails")
		return
	}

	emails := result.OrElse([]string{})
	h.logger.Info("User emails retrieved successfully", "count", len(emails))
	c.JSON(http.StatusOK, gin.H{
		"emails": emails,
		"count":  len(emails),
	})
}

// CreateUserFunctional demonstrates functional programming with Result pattern
func (h *UserHandler) CreateUserFunctional(c *gin.Context) {
	h.logger.Info("Creating user functionally", "remote_addr", c.ClientIP())

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload", 
			"details": err.Error(),
		})
		return
	}

	// Use functional approach with Result pattern
	userID, err := values.NewUserID(req.ID)
	if err != nil {
		h.logger.Warn("Invalid user ID format", "error", err, "user_id", req.ID)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID format",
			"details": err.Error(),
		})
		return
	}

	// Use the functional Result-based method
	result := h.userService.CreateUserWithResult(c.Request.Context(), userID, req.Email, req.Name)
	
	if result.IsError() {
		h.handleError(c, result.Error(), "create_user_functional")
		return
	}

	user := result.MustGet()
	h.logger.Info("User created successfully (functional)", "user_id", user.ID, "email", user.Email)
	c.JSON(http.StatusCreated, user)
}

// GetUsersFiltered demonstrates functional filtering with lo operations
func (h *UserHandler) GetUsersFiltered(c *gin.Context) {
	h.logger.Debug("Getting filtered users")

	// Parse query parameters functionally
	filters := make(map[string]interface{})
	
	// Use lo.Ternary for conditional parameter parsing
	if domain := c.Query("domain"); domain != "" {
		filters["domain"] = domain
	}
	
	if active := c.Query("active"); active != "" {
		filters["active"] = active == "true"
	}

	users, err := h.userService.GetUsersWithFilters(c.Request.Context(), filters)
	if err != nil {
		h.handleError(c, err, "get_users_filtered")
		return
	}

	h.logger.Info("Filtered users retrieved successfully", "count", len(users), "filters", filters)
	c.JSON(http.StatusOK, gin.H{
		"users":   users,
		"count":   len(users),
		"filters": filters,
	})
}

// GetUsersByDomains demonstrates complex functional operations
func (h *UserHandler) GetUsersByDomains(c *gin.Context) {
	h.logger.Debug("Getting users by domains")

	// Parse domains from query parameter
	domainsParam := c.Query("domains")
	if domainsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "domains parameter is required",
		})
		return
	}

	domains := strings.Split(domainsParam, ",")
	
	usersByDomain, err := h.userService.GetUsersByEmailDomains(c.Request.Context(), domains)
	if err != nil {
		h.handleError(c, err, "get_users_by_domains")
		return
	}

	// Calculate totals functionally using lo
	totals := lo.MapValues(usersByDomain, func(users []*entities.User, _ string) int {
		return len(users)
	})

	h.logger.Info("Users by domains retrieved successfully", "domains", totals)
	c.JSON(http.StatusOK, gin.H{
		"users_by_domain": usersByDomain,
		"totals":          totals,
		"requested_domains": domains,
	})
}

// ValidateUsersBatch demonstrates Either pattern for batch operations
func (h *UserHandler) ValidateUsersBatch(c *gin.Context) {
	h.logger.Debug("Validating users batch")

	var users []*entities.User
	if err := c.ShouldBindJSON(&users); err != nil {
		h.logger.Warn("Invalid request payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Use Either pattern for batch validation
	result := h.userService.ValidateUserBatchWithEither(users)
	
	if result.IsLeft() {
		// Validation errors occurred
		errors := result.MustLeft()
		h.logger.Warn("Batch validation failed", "error_count", len(errors))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Batch validation failed", 
			"errors": lo.Map(errors, func(err error, _ int) string {
				return err.Error()
			}),
		})
		return
	}

	// All users valid
	validUserIDs := result.MustRight()
	h.logger.Info("Batch validation succeeded", "valid_count", len(validUserIDs))
	c.JSON(http.StatusOK, gin.H{
		"message":        "All users are valid",
		"valid_user_ids": validUserIDs,
		"count":          len(validUserIDs),
	})
}
