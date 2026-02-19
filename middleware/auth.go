package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Auth validates the JWT token from the Authorization header or cookie
// and sets user_id, user_email, user_role, user_name in the Gin context.
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Try Authorization header first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"status":  false,
					"message": "Invalid authorization format, expected: Bearer <token>",
				})
				return
			}
			tokenString = parts[1]
		}

		// Fall back to cookie (for SSR requests)
		if tokenString == "" {
			if cookie, err := c.Cookie("token"); err == nil && cookie != "" {
				tokenString = cookie
			}
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Authorization required",
			})
			return
		}

		// Parse and validate JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Invalid token claims",
			})
			return
		}

		// Extract claims and set in context
		if userID, ok := claims["user_id"].(float64); ok {
			c.Set("user_id", int(userID))
		}
		if email, ok := claims["email"].(string); ok {
			c.Set("user_email", email)
		}
		if role, ok := claims["role"].(string); ok {
			c.Set("user_role", role)
		}
		if name, ok := claims["name"].(string); ok {
			c.Set("user_name", name)
		}

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
