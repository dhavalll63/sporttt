package auth

import (
	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers all auth-related routes
// @Summary Register auth routes
// @Description Register all authentication and authorization routes
func RegisterAuthRoutes(router *gin.RouterGroup) {
	// Initialize repository and controller
	repo := NewAuthRepository()
	controller := NewAuthController(repo)

	// Register routes
	router.POST("/register", controller.Register)
	router.POST("/login", controller.Login)
	router.POST("/refresh-token", controller.RefreshToken)
	router.POST("/request-otp", controller.RequestOTP)
	router.POST("/verify-otp", controller.VerifyOTP)
	router.POST("/resend-otp", controller.ResendOTP)
	router.POST("/forgot-password", controller.ForgotPassword)
	router.POST("/reset-password", controller.ResetPassword)
}
