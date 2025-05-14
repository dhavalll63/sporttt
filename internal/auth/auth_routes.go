package auth

import (
	"github.com/DhavalSuthar-24/miow/config"              // For DB and App Config
	"github.com/DhavalSuthar-24/miow/internal/middleware" // Your auth middleware
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterAuthRoutes(router *gin.RouterGroup, db *gorm.DB, appConfig *config.Config) {
	// Initialize repository and controller
	// mailerService := services.NewSESMailer(appConfig) // Example
	// smsService := services.NewTwilioSMS(appConfig)  // Example

	authRepo := NewAuthRepository(db)
	authController := NewAuthController(authRepo, appConfig /* mailerService, smsService */)

	// Public routes
	authPublic := router.Group("/auth")
	{
		authPublic.POST("/register", authController.Register)
		authPublic.POST("/login", authController.Login)
		authPublic.POST("/refresh-token", authController.RefreshToken)

		authPublic.POST("/request-otp", authController.RequestOTP)
		authPublic.POST("/verify-otp", authController.VerifyOTP)
		// Resend OTP might be similar to request-otp, or have its own logic
		// authPublic.POST("/resend-otp", authController.ResendOTP) // Assuming ResendOTP exists

		authPublic.POST("/forgot-password", authController.ForgotPassword)
		authPublic.POST("/reset-password", authController.ResetPassword)

		authPublic.GET("/verify-email", authController.VerifyEmail) // Changed to GET as it's usually a link
		authPublic.POST("/resend-verification", authController.ResendVerificationEmail)
	}

	// Authenticated routes (protected by auth middleware)
	authProtected := router.Group("/auth")
	authProtected.Use(middleware.AuthMiddleware(appConfig.JWT.AccessTokenSecret)) // Apply your auth middleware
	{
		authProtected.GET("/me", authController.GetProfile)
		authProtected.PUT("/me", authController.UpdateProfile)
		authProtected.PUT("/me/profile-image", authController.UpdateProfileImage)
		authProtected.POST("/change-password", authController.ChangePassword)
		authProtected.POST("/logout", authController.Logout) // Changed to POST
	}
}
