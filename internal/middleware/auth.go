package middleware

import (
	"net/http"
	"strings"

	"github.com/DhavalSuthar-24/miow/internal/common" // Import common package instead of auth
	"github.com/DhavalSuthar-24/miow/internal/user"   // For user.User
	"github.com/DhavalSuthar-24/miow/pkg/token"       // For token.ValidateJWT and claims struct
	"github.com/gin-gonic/gin"
	"gorm.io/gorm" // To pass DB to repository
)

const (
	AuthorizationHeaderKey = "Authorization"
	BearerSchema           = "Bearer "
)

// AuthMiddleware authenticates users based on JWT.
// It fetches the user from the database and sets it in the context.
func AuthMiddleware(jwtSecret string, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeaderKey)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		if !strings.HasPrefix(authHeader, BearerSchema) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization schema is not Bearer"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerSchema)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is missing"})
			return
		}

		claims, err := token.ValidateJWT(tokenString, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token", "details": err.Error()})
			return
		}

		userID := claims.UserID // Assuming claims struct has UserID field of appropriate type (e.g., uint)
		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: User ID not found in claims"})
			return
		}

		// Fetch user from database to ensure they exist and are active
		// And to get the most up-to-date user information, including their role
		authRepo := NewAuthRepository(db)                     // Using local interface, not importing auth package
		userRecord, err := authRepo.GetUserByID(uint(userID)) // Ensure type compatibility for userID
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found or token invalid"})
			return
		}

		// Optional: Check if user is active or not banned
		// if !userRecord.IsActive || userRecord.IsBanned {
		//  c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User account is inactive or banned"})
		//  return
		// }

		c.Set(common.ContextUserKey, userRecord)
		c.Set(common.ContextUserIDKey, userRecord.ID) // Storing ID separately can also be useful

		c.Next()
	}
}

// RoleMiddleware checks if the authenticated user has one of the required roles.
// This middleware should be used AFTER AuthMiddleware.
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, exists := c.Get(common.ContextUserKey)
		if !exists {
			// This should not happen if AuthMiddleware ran successfully
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "User not found in context. AuthMiddleware might be missing."})
			return
		}

		user, ok := currentUser.(*user.User) // Type assertion
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context."})
			return
		}

		isAllowed := false
		for _, role := range allowedRoles {
			if user.Role == role { // Assuming user.Role is a string
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: You do not have the required role(s)."})
			return
		}

		c.Next()
	}
}

// GetCurrentUser retrieves the authenticated user from the Gin context.
func GetCurrentUser(c *gin.Context) (*user.User, bool) {
	userInterface, exists := common.GetCurrentUserInterface(c)
	if !exists {
		return nil, false
	}
	user, ok := userInterface.(*user.User)
	if !ok {
		return nil, false
	}
	return user, true
}

// Now we need a local repository interface to avoid importing auth package

// AuthRepository defines auth-related database operations needed by middleware
type AuthRepository interface {
	GetUserByID(id uint) (*user.User, error)
}

// Implementation of the repository for the middleware
type authRepository struct {
	db *gorm.DB
}

// NewAuthRepository creates a new auth repository for middleware
func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

// GetUserByID fetches a user by their ID
func (r *authRepository) GetUserByID(id uint) (*user.User, error) {
	var user user.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// And finally update the auth controller to use common.GetUserIDFromContext

// File: internal/auth/controller.go (partial)
// Inside auth.GetProfile:
