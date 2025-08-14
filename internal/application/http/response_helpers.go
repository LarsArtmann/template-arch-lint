// Package http provides HTTP utilities for handlers.
// These utilities standardize response handling and correlation ID management.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/LarsArtmann/template-arch-lint/internal/application/dto"
)

// CorrelationIDKey is the key used for correlation ID in context and headers.
const CorrelationIDKey = "X-Correlation-ID"

// GetCorrelationID extracts correlation ID from gin context.
func GetCorrelationID(c *gin.Context) string {
	if id := c.GetString(CorrelationIDKey); id != "" {
		return id
	}
	// Fallback to header if not set in context
	return c.GetHeader(CorrelationIDKey)
}

// RespondSuccess sends a successful response with data.
func RespondSuccess[T any](c *gin.Context, statusCode int, data T, message string) {
	correlationID := GetCorrelationID(c)
	response := dto.SuccessResponse(data, message, correlationID)

	// Set correlation ID in response header
	if correlationID != "" {
		c.Header(CorrelationIDKey, correlationID)
	}

	c.JSON(statusCode, response)
}

// RespondError sends an error response.
func RespondError(c *gin.Context, statusCode int, code, message, errorType string, details map[string]string) {
	correlationID := GetCorrelationID(c)
	response := dto.ErrorResponse(code, message, errorType, correlationID, details)

	// Set correlation ID in response header
	if correlationID != "" {
		c.Header(CorrelationIDKey, correlationID)
	}

	c.JSON(statusCode, response)
}

// RespondValidationError sends a validation error response.
func RespondValidationError(c *gin.Context, details map[string]string) {
	correlationID := GetCorrelationID(c)
	response := dto.ValidationErrorResponse(details, correlationID)

	// Set correlation ID in response header
	if correlationID != "" {
		c.Header(CorrelationIDKey, correlationID)
	}

	c.JSON(http.StatusBadRequest, response)
}

// RespondNotFound sends a not found error response.
func RespondNotFound(c *gin.Context, resource, id string) {
	correlationID := GetCorrelationID(c)
	response := dto.NotFoundErrorResponse(resource, id, correlationID)

	// Set correlation ID in response header
	if correlationID != "" {
		c.Header(CorrelationIDKey, correlationID)
	}

	c.JSON(http.StatusNotFound, response)
}

// RespondInternalError sends an internal server error response.
func RespondInternalError(c *gin.Context) {
	correlationID := GetCorrelationID(c)
	response := dto.InternalErrorResponse(correlationID)

	// Set correlation ID in response header
	if correlationID != "" {
		c.Header(CorrelationIDKey, correlationID)
	}

	c.JSON(http.StatusInternalServerError, response)
}

// RespondCreated sends a successful creation response.
func RespondCreated[T any](c *gin.Context, data T, message string) {
	RespondSuccess(c, http.StatusCreated, data, message)
}

// RespondOK sends a successful OK response.
func RespondOK[T any](c *gin.Context, data T, message string) {
	RespondSuccess(c, http.StatusOK, data, message)
}

// RespondNoContent sends a successful no content response.
func RespondNoContent(c *gin.Context) {
	correlationID := GetCorrelationID(c)
	if correlationID != "" {
		c.Header(CorrelationIDKey, correlationID)
	}
	c.Status(http.StatusNoContent)
}
