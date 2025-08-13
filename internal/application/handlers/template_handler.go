// Package handlers provides HTTP request handlers for the web application.
package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/a-h/templ"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/LarsArtmann/template-arch-lint/web/templates/pages"
	"github.com/LarsArtmann/template-arch-lint/web/templates/components"
)

// TemplateHandler handles template-based HTTP requests with HTMX support
type TemplateHandler struct {
	userService *services.UserService
	logger      *slog.Logger
}

// NewTemplateHandler creates a new TemplateHandler
func NewTemplateHandler(userService *services.UserService, logger *slog.Logger) *TemplateHandler {
	return &TemplateHandler{
		userService: userService,
		logger:      logger,
	}
}

// renderTemplate is a helper function to render templ components
func (h *TemplateHandler) renderTemplate(c *gin.Context, component templ.Component) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.Error("Failed to render template", "error", err)
		c.String(http.StatusInternalServerError, "Template rendering error")
	}
}

// showErrorTemplate renders error messages for HTMX responses
func (h *TemplateHandler) showErrorTemplate(c *gin.Context, message, details string) {
	h.renderTemplate(c, components.ErrorMessage(message, details))
}

// UsersPage renders the main users list page
func (h *TemplateHandler) UsersPage(c *gin.Context) {
	h.logger.Debug("Rendering users page")

	// Get users from service
	users, err := h.userService.ListUsers(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list users", "error", err)
		h.renderTemplate(c, pages.UsersPage([]*entities.User{}, make(map[string]int)))
		return
	}

	// Get user stats
	stats, err := h.userService.GetUserStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get user stats", "error", err)
		stats = make(map[string]int)
	}

	h.renderTemplate(c, pages.UsersPage(users, stats))
}

// SearchUsers handles HTMX search requests
func (h *TemplateHandler) SearchUsers(c *gin.Context) {
	h.logger.Debug("Searching users")

	query := strings.TrimSpace(c.Query("search"))
	filters := h.buildSearchFilters(c)

	users, err := h.performUserSearch(c, query, filters)
	if err != nil {
		h.logger.Error("Failed to search users", "error", err)
		h.showErrorTemplate(c, "Search failed", err.Error())
		return
	}

	h.renderTemplate(c, pages.SearchUsersContent(users))
}

// buildSearchFilters constructs filter map from query parameters
func (h *TemplateHandler) buildSearchFilters(c *gin.Context) map[string]interface{} {
	filters := make(map[string]interface{})
	
	if domain := c.Query("domain"); domain != "" {
		filters["domain"] = domain
	}
	
	if activeParam := c.Query("active"); activeParam != "" {
		filters["active"] = activeParam == "true"
	}
	
	return filters
}

// performUserSearch executes the search with given query and filters
func (h *TemplateHandler) performUserSearch(c *gin.Context, query string, filters map[string]interface{}) ([]*entities.User, error) {
	if query != "" || len(filters) > 0 {
		return h.searchWithFilters(c, query, filters)
	}
	
	return h.userService.ListUsers(c.Request.Context())
}

// searchWithFilters performs filtered search and applies text filtering if needed
func (h *TemplateHandler) searchWithFilters(c *gin.Context, query string, filters map[string]interface{}) ([]*entities.User, error) {
	users, err := h.userService.GetUsersWithFilters(c.Request.Context(), filters)
	if err != nil {
		return nil, err
	}

	if query != "" {
		return h.filterUsersByText(users, query), nil
	}
	
	return users, nil
}

// filterUsersByText filters users by text search in name and email
func (h *TemplateHandler) filterUsersByText(users []*entities.User, query string) []*entities.User {
	filteredUsers := make([]*entities.User, 0)
	queryLower := strings.ToLower(query)
	
	for _, user := range users {
		if h.userMatchesQuery(user, queryLower) {
			filteredUsers = append(filteredUsers, user)
		}
	}
	
	return filteredUsers
}

// userMatchesQuery checks if user matches the search query
func (h *TemplateHandler) userMatchesQuery(user *entities.User, queryLower string) bool {
	return strings.Contains(strings.ToLower(user.Name), queryLower) ||
		strings.Contains(strings.ToLower(user.Email), queryLower)
}

// CreateUserPage renders the create user form
func (h *TemplateHandler) CreateUserPage(c *gin.Context) {
	h.logger.Debug("Rendering create user page")
	h.renderTemplate(c, pages.CreateUserPage())
}

// CreateUser handles user creation via HTMX form submission
func (h *TemplateHandler) CreateUser(c *gin.Context) {
	h.logger.Info("Creating user via template", "remote_addr", c.ClientIP())

	// Parse form data
	id := strings.TrimSpace(c.PostForm("id"))
	email := strings.TrimSpace(c.PostForm("email"))
	name := strings.TrimSpace(c.PostForm("name"))

	// Validate input
	if id == "" || email == "" || name == "" {
		h.logger.Warn("Missing required fields")
		h.showErrorTemplate(c, "All fields are required", "Please fill in all required fields")
		return
	}

	// Create user ID value object
	userID, err := values.NewUserID(id)
	if err != nil {
		h.logger.Warn("Invalid user ID format", "error", err, "user_id", id)
		h.showErrorTemplate(c, "Invalid user ID", err.Error())
		return
	}

	// Create user using service
	user, err := h.userService.CreateUser(c.Request.Context(), userID, email, name)
	if err != nil {
		h.logger.Error("Failed to create user", "error", err, "user_id", id)
		h.showErrorTemplate(c, "Failed to create user", err.Error())
		return
	}

	h.logger.Info("User created successfully via template", "user_id", user.ID, "email", user.Email)
	h.renderTemplate(c, components.UserFormSuccess(user, "created"))
}

// EditUserPage renders the edit user form
func (h *TemplateHandler) EditUserPage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := values.NewUserID(idStr)
	if err != nil {
		h.logger.Warn("Invalid user ID format", "error", err, "user_id", idStr)
		c.String(http.StatusBadRequest, "Invalid user ID format")
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get user for edit", "error", err, "user_id", id)
		c.String(http.StatusNotFound, "User not found")
		return
	}

	h.renderTemplate(c, pages.EditUserPage(user))
}

// EditUserInline renders inline edit form for HTMX
func (h *TemplateHandler) EditUserInline(c *gin.Context) {
	idStr := c.Param("id")
	id, err := values.NewUserID(idStr)
	if err != nil {
		h.logger.Warn("Invalid user ID format", "error", err, "user_id", idStr)
		h.showErrorTemplate(c, "Invalid user ID", err.Error())
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get user for inline edit", "error", err, "user_id", id)
		h.showErrorTemplate(c, "User not found", err.Error())
		return
	}

	h.renderTemplate(c, components.UserEditRow(user))
}

// CancelUserEdit cancels inline editing and returns to view mode
func (h *TemplateHandler) CancelUserEdit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := values.NewUserID(idStr)
	if err != nil {
		h.logger.Warn("Invalid user ID format", "error", err, "user_id", idStr)
		h.showErrorTemplate(c, "Invalid user ID", err.Error())
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get user for cancel edit", "error", err, "user_id", id)
		h.showErrorTemplate(c, "User not found", err.Error())
		return
	}

	h.renderTemplate(c, components.UserRow(user))
}

// UpdateUser handles user updates via HTMX
func (h *TemplateHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := values.NewUserID(idStr)
	if err != nil {
		h.logger.Warn("Invalid user ID format", "error", err, "user_id", idStr)
		h.showErrorTemplate(c, "Invalid user ID", err.Error())
		return
	}

	// Parse form data
	email := strings.TrimSpace(c.PostForm("email"))
	name := strings.TrimSpace(c.PostForm("name"))

	// Validate input
	if email == "" || name == "" {
		h.logger.Warn("Missing required fields for update")
		h.showErrorTemplate(c, "All fields are required", "Please fill in all required fields")
		return
	}

	// Update user using service
	user, err := h.userService.UpdateUser(c.Request.Context(), id, email, name)
	if err != nil {
		h.logger.Error("Failed to update user", "error", err, "user_id", id)
		h.showErrorTemplate(c, "Failed to update user", err.Error())
		return
	}

	h.logger.Info("User updated successfully via template", "user_id", user.ID)
	
	// Check if this is a form page update or inline update
	if c.GetHeader("HX-Request") != "" && strings.Contains(c.GetHeader("HX-Target"), "user-") {
		// Inline update - return updated row
		h.renderTemplate(c, components.UserRow(user))
	} else {
		// Form page update - return success message
		h.renderTemplate(c, components.UserFormSuccess(user, "updated"))
	}
}

// DeleteUser handles user deletion via HTMX
func (h *TemplateHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := values.NewUserID(idStr)
	if err != nil {
		h.logger.Warn("Invalid user ID format", "error", err, "user_id", idStr)
		h.showErrorTemplate(c, "Invalid user ID", err.Error())
		return
	}

	// Delete user using service
	if err := h.userService.DeleteUser(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete user", "error", err, "user_id", id)
		h.showErrorTemplate(c, "Failed to delete user", err.Error())
		return
	}

	h.logger.Info("User deleted successfully via template", "user_id", id)
	
	// Return empty content to remove the row from DOM
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, "")
}

// UserStatsPartial returns user statistics as HTML partial for HTMX updates
func (h *TemplateHandler) UserStatsPartial(c *gin.Context) {
	stats, err := h.userService.GetUserStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get user stats", "error", err)
		h.showErrorTemplate(c, "Failed to load statistics", err.Error())
		return
	}

	h.renderTemplate(c, components.StatsGrid(stats))
}