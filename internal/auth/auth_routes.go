package auth

import (
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.RouterGroup) {
	r.POST("/register", Register)
	r.POST("/login", Login)
	r.POST("/refresh-token", RefreshToken)
	r.POST("/request-otp", RequestOTP)
	r.POST("/verify-otp", VerifyOTP)
	r.POST("/resend-otp", ResendOTP)
}
