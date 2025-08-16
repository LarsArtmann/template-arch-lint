// Package middleware provides HTTP middleware functionality.
package middleware

import (
	"net/http"
	"strings"

	httputil "github.com/LarsArtmann/template-arch-lint/internal/application/http"
	"github.com/LarsArtmann/template-arch-lint/internal/application/services"
	"github.com/gin-gonic/gin"
)

// JWTMiddleware provides JWT authentication middleware.
type JWTMiddleware struct {
	jwtService *services.JWTService
}

// NewJWTMiddleware creates a new JWT middleware.
func NewJWTMiddleware(jwtService *services.JWTService) *JWTMiddleware {
	return &JWTMiddleware{
		jwtService: jwtService,
	}
}

// AuthenticateJWT validates JWT tokens from Authorization header.
func (m *JWTMiddleware) AuthenticateJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractTokenFromHeader(c)
		if token == "" {
			httputil.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header missing or invalid", "authentication", nil)
			c.Abort()
			return
		}

		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			httputil.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token", "authentication", nil)
			c.Abort()
			return
		}

		// Only allow access tokens for API access
		if claims.Type != "access" {
			httputil.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token type", "authentication", nil)
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// OptionalAuth validates JWT tokens but allows requests without them.
func (m *JWTMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractTokenFromHeader(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			// Don't abort, just continue without user context
			c.Next()
			return
		}

		if claims.Type == "access" {
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("jwt_claims", claims)
		}

		c.Next()
	}
}

// extractTokenFromHeader extracts JWT token from Authorization header.
func (m *JWTMiddleware) extractTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// GetUserIDFromContext retrieves user ID from gin context.
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

// GetUserEmailFromContext retrieves user email from gin context.
func GetUserEmailFromContext(c *gin.Context) (string, bool) {
	email, exists := c.Get("user_email")
	if !exists {
		return "", false
	}

	emailStr, ok := email.(string)
	return emailStr, ok
}

// GetJWTClaimsFromContext retrieves JWT claims from gin context.
func GetJWTClaimsFromContext(c *gin.Context) (*services.Claims, bool) {
	claims, exists := c.Get("jwt_claims")
	if !exists {
		return nil, false
	}

	jwtClaims, ok := claims.(*services.Claims)
	return jwtClaims, ok
}
