package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Auth is a placeholder JWT authentication middleware.
// In Phase 3 this will be expanded with proper JWT validation.
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Authorization header required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Invalid authorization format, expected: Bearer <token>",
			})
			return
		}

		token := parts[1]
		_ = token // TODO Phase 3: validate JWT, extract claims

		// Placeholder: set user context
		c.Set("user_id", 1)
		c.Set("user_role", "admin")

		c.Next()
	}
}

// RequireRole returns middleware that checks if the authenticated user
// has one of the specified roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  false,
				"message": "Access denied",
			})
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  false,
				"message": "Invalid user role",
			})
			return
		}

		for _, r := range roles {
			if r == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"status":  false,
			"message": "Insufficient permissions",
		})
	}
}
