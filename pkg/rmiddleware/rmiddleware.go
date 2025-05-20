package rmiddleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/DhavalSuthar-24/miow/internal/auth"
	"github.com/DhavalSuthar-24/miow/internal/middleware"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

func RoleMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := middleware.GetUserIDFromContext(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
			return
		}

		// Get the auth repository from context or create a new one
		authRepo, ok := c.MustGet("auth_repo").(auth.AuthRepository)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Get user's roles
		userRoles, err := authRepo.GetUserRoles(userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User roles not found"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
			return
		}

		// Check if user has any of the required roles
		hasRequiredRole := false
		for _, userRole := range userRoles {
			for _, requiredRole := range requiredRoles {
				if strings.EqualFold(userRole, requiredRole) {
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}

		if !hasRequiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":      "Forbidden",
				"message":    "You don't have permission to access this resource",
				"required":   requiredRoles,
				"user_roles": userRoles,
			})
			return
		}

		// Add roles to context for downstream handlers
		c.Set("user_roles", userRoles)
		c.Next()
	}
}

// AdminMiddleware is a convenience middleware for admin-only access
func AdminMiddleware() gin.HandlerFunc {
	return RoleMiddleware("admin")
}

// PlayerMiddleware is a convenience middleware for player-only access
func PlayerMiddleware() gin.HandlerFunc {
	return RoleMiddleware("player")
}

// CoachOrAdminMiddleware is a convenience middleware for coach or admin access
func CoachOrAdminMiddleware() gin.HandlerFunc {
	return RoleMiddleware("coach", "admin")
}
func VenueManagerhOrAdminMiddleware() gin.HandlerFunc {
	return RoleMiddleware("venue_manager", "admin")
}
