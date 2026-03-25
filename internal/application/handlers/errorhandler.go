package handlers

import (
	"github.com/gin-gonic/gin"
)

// sendErrorResponse sends a standardized error response to the HTTP client.
// This helper eliminates code duplication in handler error responses.
func sendErrorResponse(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, gin.H{"error": message})
	return
}
