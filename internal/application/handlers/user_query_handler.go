package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	pkgerrors "github.com/LarsArtmann/template-arch-lint/pkg/errors"
	"github.com/samber/lo"
)

type UserQueryHandler struct {
	userQueryService services.UserQueryService
}

func NewUserQueryHandler(userQueryService services.UserQueryService) *UserQueryHandler {
	return &UserQueryHandler{
		userQueryService: userQueryService,
	}
}

func (h *UserQueryHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/users/query/{id}", h.GetUser)
	mux.HandleFunc("GET /api/v1/users/query", h.ListUsers)
	mux.HandleFunc("GET /api/v1/users/search", h.SearchUsers)
	mux.HandleFunc("GET /api/v1/users/domain/{domain}", h.GetUsersByDomain)
	mux.HandleFunc("GET /api/v1/users/stats", h.GetUserStats)
	mux.HandleFunc("GET /api/v1/users/active", h.GetActiveUsers)
	mux.HandleFunc("GET /api/v1/users/paginated", h.GetUsersWithPagination)
}

func (h *UserQueryHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")

	userID, err := values.NewUserID(idParam)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid user ID format")

		return
	}

	user, err := h.userQueryService.GetUser(r.Context(), userID)
	if err != nil {
		_, isNotFound := pkgerrors.AsNotFoundError(err)
		if isNotFound {
			sendErrorResponse(w, http.StatusNotFound, "User not found")

			return
		}

		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve user")

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": user})
}

func (h *UserQueryHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userQueryService.ListUsers(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve users")

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": users})
}

func (h *UserQueryHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Email query parameter is required")

		return
	}

	user, err := h.userQueryService.GetUserByEmail(r.Context(), email)
	if err != nil {
		_, isNotFound := pkgerrors.AsNotFoundError(err)
		if isNotFound {
			writeJSON(w, http.StatusOK, map[string]any{"data": []any{}})

			return
		}

		sendErrorResponse(w, http.StatusInternalServerError, "Failed to search users")

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": []*entities.User{user}})
}

func (h *UserQueryHandler) GetUsersByDomain(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	if domain == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Domain parameter is required")

		return
	}

	users, err := h.userQueryService.ListUsers(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve users")

		return
	}

	filteredUsers := lo.Filter(users, func(user *entities.User, _ int) bool {
		userEmail := user.GetEmail().String()
		if strings.Contains(userEmail, "@") {
			parts := strings.Split(userEmail, "@")

			return len(parts) == 2 && parts[1] == domain
		}

		return false
	})

	writeJSON(w, http.StatusOK, map[string]any{"data": filteredUsers})
}

func (h *UserQueryHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.userQueryService.GetUserStats(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve user statistics")

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": stats})
}

func (h *UserQueryHandler) GetActiveUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userQueryService.ListUsers(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve active users")

		return
	}

	activeUsers := lo.Filter(users, func(user *entities.User, _ int) bool {
		return true
	})

	writeJSON(w, http.StatusOK, map[string]any{"data": activeUsers})
}

func (h *UserQueryHandler) GetUsersWithPagination(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	users, err := h.userQueryService.ListUsers(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve users")

		return
	}

	total := len(users)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		writeJSON(w, http.StatusOK, map[string]any{
			"data": []*entities.User{},
			"pagination": map[string]any{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		})

		return
	}

	if end > total {
		end = total
	}

	paginatedUsers := users[start:end]

	writeJSON(w, http.StatusOK, map[string]any{
		"data": paginatedUsers,
		"pagination": map[string]any{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
