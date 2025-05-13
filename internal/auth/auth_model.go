package auth

import (
	"time"

	"gorm.io/gorm"
)

// RegisterRequest represents the payload for registering a new user
type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
	Phone    string `json:"phone" binding:"required" example:"+919876543210"`
	Roles    string `json:"roles" example:"player"`
}

// LoginRequest represents the payload for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// OTPReq represents a request for OTP verification
type OTPReq struct {
	Phone string `json:"phone" binding:"required" example:"+919876543210"`
}

// OTPVerifyRequest represents the payload for OTP verification
type OTPVerifyRequest struct {
	Phone string `json:"phone" binding:"required" example:"+919876543210"`
	Code  string `json:"code" binding:"required" example:"123456"`
}

// AccessReq represents a request for refreshing an access token
type AccessReq struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// ForgotPasswordRequest represents a request to initiate password reset
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
}

// ResetPasswordRequest represents a request to reset password with token
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required" example:"reset-token-123456"`
	Password string `json:"password" binding:"required,min=6" example:"newpassword123"`
}

// OTP model for storing one-time passwords
type OTP struct {
	gorm.Model
	Phone     string    `gorm:"not null;index"`
	Code      string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Verified  bool      `gorm:"default:false"`
}
