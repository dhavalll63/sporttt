package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/DhavalSuthar-24/miow/pkg/token"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	AuthUserIDKey = "auth_user_id"
)

func AuthMiddleware(jwtSecret string, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format. Expected: Bearer <token>"})
			return
		}

		jwtToken, err := token.ValidateToken(bearerToken[1], jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token: " + err.Error()})
			return
		}

		if !jwtToken.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		userID, err := token.ExtractUserID(jwtToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Could not extract user ID from token: " + err.Error()})
			return
		}

		var exists bool
		if err := db.Table("users").Select("1").Where("id = ? AND deleted_at IS NULL", userID).Scan(&exists).Error; err != nil || !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found or inactive"})
			return
		}

		c.Set(AuthUserIDKey, userID)
		c.Next()
	}
}

// GetUserIDFromContext extracts the user ID from the context
func GetUserIDFromContext(c *gin.Context) (uint, error) {
	userID, exists := c.Get(AuthUserIDKey)
	if !exists {
		return 0, errors.New("user ID not found in context")
	}

	uid, ok := userID.(uint)
	if !ok {
		return 0, fmt.Errorf("user ID has unexpected type: %T", userID)
	}

	return uid, nil
}
