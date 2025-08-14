// Package middleware provides HTTP middleware for error handling
package middleware

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	domainerrors "github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
)

// ErrorResponse represents a structured error response.
type ErrorResponse struct {
	Error   string                    `json:"error"`
	Code    domainerrors.ErrorCode    `json:"code"`
	Message string                    `json:"message"`
	Details domainerrors.ErrorDetails `json:"details,omitempty"`
}

// ErrorHandler wraps a handler and provides structured error handling.
type ErrorHandler func(w http.ResponseWriter, r *http.Request) error

// ServeHTTP implements the http.Handler interface for ErrorHandler.
func (eh ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := eh(w, r); err != nil {
		HandleError(w, r, err)
	}
}

// HandleError processes errors and returns appropriate HTTP responses.
func HandleError(w http.ResponseWriter, _ *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	// Check if it's a domain error
	var domainErr domainerrors.DomainError
	if errors.As(err, &domainErr) {
		status := domainErr.HTTPStatus()
		w.WriteHeader(status)

		response := ErrorResponse{
			Error:   "domain_error",
			Code:    domainErr.Code(),
			Message: domainErr.Error(),
			Details: domainErr.Details(),
		}

		if jsonErr := json.NewEncoder(w).Encode(response); jsonErr != nil {
			slog.Error("Error encoding error response", "error", jsonErr)
		}
		return
	}

	// Handle non-domain errors as internal server errors
	w.WriteHeader(http.StatusInternalServerError)
	response := ErrorResponse{
		Error:   "internal_error",
		Code:    domainerrors.InternalErrorCode,
		Message: "An internal server error occurred",
		Details: domainerrors.ErrorDetails{
			Reason: err.Error(),
		},
	}

	if jsonErr := json.NewEncoder(w).Encode(response); jsonErr != nil {
		slog.Error("Error encoding error response", "error", jsonErr)
	}

	// Log internal errors for debugging
	slog.Error("Internal server error", "error", err)
}

// ValidationErrorMiddleware specifically handles validation errors.
func ValidationErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a custom response writer to capture the response
		recorder := &responseRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		// If there was an error, it should have been handled by ErrorHandler
		// This middleware provides additional validation-specific logic if needed
	})
}

// responseRecorder captures response details for middleware processing.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	r.body = body
	return r.ResponseWriter.Write(body)
}

// RecoveryMiddleware handles panics and converts them to errors.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("Panic recovered", "error", err)

				// Convert panic to internal error
				internalErr := domainerrors.NewInternalError("panic recovered", nil)
				HandleError(w, r, internalErr)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs requests and errors with structured logging.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("HTTP request started",
			"method", r.Method,
			"path", r.URL.Path,
		)

		recorder := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)

		if recorder.statusCode >= 400 {
			slog.Warn("HTTP error response",
				"status_code", recorder.statusCode,
				"method", r.Method,
				"path", r.URL.Path,
			)
		}
	})
}
