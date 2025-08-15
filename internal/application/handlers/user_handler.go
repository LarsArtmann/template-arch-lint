package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/LarsArtmann/template-arch-lint/internal/application/dto"
	httputil "github.com/LarsArtmann/template-arch-lint/internal/application/http"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	domainerrors "github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	utilsErrors "github.com/LarsArtmann/template-arch-lint/internal/utils/errors"
	utilsValidation "github.com/LarsArtmann/template-arch-lint/internal/utils/validation"
	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userService  *services.UserService
	logger       *slog.Logger
	validators   *utilsValidation.PrebuiltValidators
	errorFactory *utilsErrors.ErrorFactory
}

// NewUserHandler creates a new UserHandler with dependency injection.
func NewUserHandler(userService *services.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		userService:  userService,
		logger:       logger,
		validators:   utilsValidation.NewPrebuiltValidators(),
		errorFactory: utilsErrors.NewErrorFactory(),
	}
}

// Note: Request types moved to dto package for better separation of concerns

// CreateUser creates a new user with comprehensive validation and sanitization.
func (h *UserHandler) CreateUser(c *gin.Context) {
	correlationID := httputil.GetCorrelationID(c)
	h.logger.Info("Creating user",
		"remote_addr", c.ClientIP(),
		"correlation_id", correlationID)

	// Create context with timeout
	ctx := context.WithValue(c.Request.Context(), utilsErrors.CorrelationIDKey, correlationID)
	ctx = context.WithValue(ctx, utilsErrors.OperationKey, "create_user")

	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = h.errorFactory.Validation(err, "request_payload").
			WithContext(timeoutCtx).
			WithOperation("bind_json")

		h.logger.Warn("Invalid request payload",
			"error", err,
			"correlation_id", correlationID)

		httputil.RespondValidationError(c, map[string]string{
			"request": "Invalid JSON payload: " + err.Error(),
		})
		return
	}

	// Sanitize input data
	sanitizedID := h.validators.Sanitization.TrimWhitespace(req.ID)
	sanitizedID = h.validators.Sanitization.StripNonPrintable(sanitizedID)

	sanitizedEmail := h.validators.Sanitization.TrimWhitespace(req.Email)
	sanitizedEmail = h.validators.Sanitization.NormalizeWhitespace(sanitizedEmail)

	sanitizedName := h.validators.Sanitization.TrimWhitespace(req.Name)
	sanitizedName = h.validators.Sanitization.EscapeHTML(sanitizedName)
	sanitizedName = h.validators.Sanitization.NormalizeWhitespace(sanitizedName)

	// Validate input using utility validators
	validationResult := utilsValidation.ValidateWithResult(timeoutCtx, sanitizedID,
		utilsValidation.ToValidator(h.validators.String.NotEmpty("id")),
		utilsValidation.ToValidator(h.validators.String.LengthRange("id", 1, 100)),
		utilsValidation.ToValidator(h.validators.String.NoSpecialChars("id", '-', '_')),
	)

	if !validationResult.IsValid() {
		h.logger.Warn("User ID validation failed",
			"user_id", sanitizedID,
			"errors", validationResult.AllErrors(),
			"correlation_id", correlationID)

		httputil.RespondValidationError(c, map[string]string{
			"id": validationResult.FirstError(),
		})
		return
	}

	// Validate email
	emailValidation := utilsValidation.ValidateWithResult(timeoutCtx, sanitizedEmail,
		utilsValidation.ToValidator(h.validators.String.NotEmpty("email")),
		utilsValidation.ToValidator(h.validators.String.Email("email")),
		utilsValidation.ToValidator(h.validators.String.MaxLength("email", 255)),
	)

	if !emailValidation.IsValid() {
		h.logger.Warn("Email validation failed",
			"email", sanitizedEmail,
			"errors", emailValidation.AllErrors(),
			"correlation_id", correlationID)

		httputil.RespondValidationError(c, map[string]string{
			"email": emailValidation.FirstError(),
		})
		return
	}

	// Validate name
	nameValidation := utilsValidation.ValidateWithResult(timeoutCtx, sanitizedName,
		utilsValidation.ToValidator(h.validators.String.NotEmpty("name")),
		utilsValidation.ToValidator(h.validators.String.LengthRange("name", 1, 255)),
	)

	if !nameValidation.IsValid() {
		h.logger.Warn("Name validation failed",
			"name", sanitizedName,
			"errors", nameValidation.AllErrors(),
			"correlation_id", correlationID)

		httputil.RespondValidationError(c, map[string]string{
			"name": nameValidation.FirstError(),
		})
		return
	}

	// Create user ID using validated input
	userID, err := values.NewUserID(sanitizedID)
	if err != nil {
		_ = h.errorFactory.Validation(err, "user_id").
			WithContext(timeoutCtx).
			WithExtra("user_id", sanitizedID)

		h.logger.Warn("Invalid user ID format",
			"error", err,
			"user_id", sanitizedID,
			"correlation_id", correlationID)

		httputil.RespondValidationError(c, map[string]string{
			"id": "Invalid user ID format: " + err.Error(),
		})
		return
	}

	// Create user using service layer with sanitized data
	user, err := h.userService.CreateUser(timeoutCtx, userID, sanitizedEmail, sanitizedName)
	if err != nil {
		wrappedErr := h.errorFactory.WrapWithContext(timeoutCtx, err, "failed to create user")
		h.logger.Error("Failed to create user",
			"error", wrappedErr,
			"user_id", sanitizedID,
			"correlation_id", correlationID)

		if errors.Is(err, repositories.ErrUserAlreadyExists) {
			httputil.RespondError(c, http.StatusConflict,
				"USER_ALREADY_EXISTS",
				"User or email already exists",
				"conflict",
				map[string]string{
					"id":    sanitizedID,
					"email": sanitizedEmail,
				})
			return
		}

		if validationErr, ok := domainerrors.AsValidationError(err); ok {
			httputil.RespondValidationError(c, map[string]string{
				"validation": validationErr.Error(),
			})
			return
		}

		httputil.RespondInternalError(c)
		return
	}

	// Record successful user creation and return standardized response
	h.logger.Info("User created successfully",
		"user_id", user.ID.String(),
		"email", user.Email,
		"correlation_id", correlationID)

	userResponse := dto.ToUserResponse(user)
	httputil.RespondCreated(c, userResponse, "User created successfully")
}

// GetUser retrieves a user by ID.
func (h *UserHandler) GetUser(c *gin.Context) {
	correlationID := httputil.GetCorrelationID(c)
	idStr := c.Param("id")

	h.logger.Info("Getting user",
		"user_id", idStr,
		"correlation_id", correlationID)

	id, err := values.NewUserID(idStr)
	if err != nil {
		h.logger.Warn("Invalid user ID format",
			"error", err,
			"user_id", idStr,
			"correlation_id", correlationID)

		httputil.RespondValidationError(c, map[string]string{
			"id": "Invalid user ID format: " + err.Error(),
		})
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get user",
			"error", err,
			"user_id", id.String(),
			"correlation_id", correlationID)

		if errors.Is(err, repositories.ErrUserNotFound) {
			httputil.RespondNotFound(c, "User", id.String())
			return
		}

		httputil.RespondInternalError(c)
		return
	}

	h.logger.Info("User retrieved successfully",
		"user_id", user.ID.String(),
		"correlation_id", correlationID)

	userResponse := dto.ToUserResponse(user)
	httputil.RespondOK(c, userResponse, "User retrieved successfully")
}

// UpdateUser updates an existing user.
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := h.parseAndValidateUserID(c, idStr)
	if err != nil {
		return // Error already handled in parseAndValidateUserID
	}

	var req dto.UpdateUserRequest
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

// DeleteUser removes a user.
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := h.parseAndValidateUserID(c, idStr)
	if err != nil {
		return // Error already handled in parseAndValidateUserID
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

// ListUsers retrieves all users with optimized pagination.
// Memory optimization: Pagination should ideally be done at database level,
// but for demo purposes we'll optimize the slicing logic.
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse optional query parameters with validation
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit := parsePositiveInt(limitStr, 50, 1000) // Cap at 1000 for safety
	offset := parsePositiveInt(offsetStr, 0, 0)   // No upper limit for offset

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

	// Optimized pagination bounds checking
	total := len(users)
	paginatedUsers := paginateSlice(users, offset, limit)

	h.logger.Info("Users listed successfully", "total", total, "returned", len(paginatedUsers))
	c.JSON(http.StatusOK, gin.H{
		"users":  paginatedUsers,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// parsePositiveInt parses a string to positive int with bounds checking.
func parsePositiveInt(s string, defaultVal, maxVal int) int {
	val, err := strconv.Atoi(s)
	if err != nil || val < 0 {
		return defaultVal
	}
	if maxVal > 0 && val > maxVal {
		return maxVal
	}
	return val
}

// paginateSlice efficiently paginates a slice with bounds checking.
func paginateSlice[T any](slice []T, offset, limit int) []T {
	total := len(slice)
	if offset >= total {
		return []T{} // Return empty slice of same type
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return slice[offset:end]
}

// parseAndValidateUserID extracts user ID parsing logic to reduce duplication.
func (h *UserHandler) parseAndValidateUserID(c *gin.Context, idStr string) (values.UserID, error) {
	id, err := values.NewUserID(idStr)
	if err != nil {
		h.logger.Warn("Invalid user ID format", "error", err, "user_id", idStr)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID format",
			"details": err.Error(),
		})
		return values.UserID{}, err
	}
	return id, nil
}

// handleError processes errors and returns appropriate HTTP responses using typed errors.
func (h *UserHandler) handleError(c *gin.Context, err error, operation string) {
	h.logger.Error("Operation failed", "operation", operation, "error", err)

	// Handle domain errors with proper types
	var domainErr domainerrors.DomainError
	if errors.As(err, &domainErr) {
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

// GetUserStats demonstrates functional operations endpoint.
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

// GetActiveUsers demonstrates functional programming endpoint.
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

// GetUserEmails demonstrates Result pattern endpoint.
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

// CreateUserFunctional demonstrates functional programming with Result pattern.
func (h *UserHandler) CreateUserFunctional(c *gin.Context) {
	h.logger.Info("Creating user functionally", "remote_addr", c.ClientIP())

	var req dto.CreateUserRequest
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

// GetUsersFiltered demonstrates functional filtering with lo operations.
func (h *UserHandler) GetUsersFiltered(c *gin.Context) {
	h.logger.Debug("Getting filtered users")

	// Parse query parameters into type-safe filters
	var filters services.UserFilters

	// Use pointer assignment for optional filters
	if domain := c.Query("domain"); domain != "" {
		filters.Domain = &domain
	}

	if active := c.Query("active"); active != "" {
		isActive := active == "true"
		filters.Active = &isActive
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

// GetUsersByDomains demonstrates complex functional operations.
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
		"users_by_domain":   usersByDomain,
		"totals":            totals,
		"requested_domains": domains,
	})
}

// ValidateUsersBatch demonstrates Either pattern for batch operations.
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
			"error": "Batch validation failed",
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
