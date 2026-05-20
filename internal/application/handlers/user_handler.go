package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"charm.land/log/v2"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

const userIDByteLength = 8

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func generateUserID() string {
	bytes := make([]byte, userIDByteLength)
	_, _ = rand.Read(bytes)

	return hex.EncodeToString(bytes)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, status int, errCode, message string) {
	writeJSON(w, status, map[string]string{
		"error":   errCode,
		"message": message,
	})
}

func bindRequest[T any](r *http.Request, req *T) bool {
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		log.Error("Invalid request format", "error", err)
		return false
	}

	return true
}

func userToMap(user *entities.User) map[string]any {
	return map[string]any{
		"id":        user.ID.String(),
		"email":     user.GetEmail().String(),
		"name":      user.GetUserName().String(),
		"createdAt": user.GetCreatedAt(),
		"updatedAt": user.GetUpdatedAt(),
	}
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/users", h.CreateUser)
	mux.HandleFunc("GET /api/v1/users/{id}", h.GetUser)
	mux.HandleFunc("PUT /api/v1/users/{id}", h.UpdateUser)
	mux.HandleFunc("DELETE /api/v1/users/{id}", h.DeleteUser)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if !bindRequest(r, &req) {
		errorResponse(w, http.StatusBadRequest, "invalid_request_format", "Invalid request body")
		return
	}

	userID, err := values.NewUserID(generateUserID())
	if err != nil {
		log.Error("Failed to generate user ID", "error", err)
		errorResponse(w, http.StatusInternalServerError, "user_id_generation_failed", "Failed to generate user ID")
		return
	}

	user, err := h.userService.CreateUser(r.Context(), userID, req.Email, req.Name)
	if err != nil {
		log.Error("Failed to create user", "error", err)
		errorResponse(w, http.StatusInternalServerError, "user_creation_failed", "Failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, userToMap(user))
}

func parseUserID(r *http.Request) (values.UserID, bool) {
	idStr := r.PathValue("id")

	userID, err := values.NewUserID(idStr)
	if err != nil {
		log.Error("Invalid user ID format", "error", err)
		return values.UserID{}, false
	}

	return userID, true
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(r)
	if !ok {
		errorResponse(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID format")
		return
	}

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		log.Error("Failed to get user", "error", err)
		errorResponse(w, http.StatusNotFound, "user_not_found", "User not found")
		return
	}

	writeJSON(w, http.StatusOK, userToMap(user))
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(r)
	if !ok {
		errorResponse(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID format")
		return
	}

	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if !bindRequest(r, &req) {
		errorResponse(w, http.StatusBadRequest, "invalid_request_format", "Invalid request body")
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), userID, req.Email, req.Name)
	if err != nil {
		log.Error("Failed to update user", "error", err)
		errorResponse(w, http.StatusInternalServerError, "user_update_failed", "Failed to update user")
		return
	}

	writeJSON(w, http.StatusOK, userToMap(user))
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(r)
	if !ok {
		errorResponse(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID format")
		return
	}

	err := h.userService.DeleteUser(r.Context(), userID)
	if err != nil {
		log.Error("Failed to delete user", "error", err)
		errorResponse(w, http.StatusInternalServerError, "user_deletion_failed", "Failed to delete user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}
