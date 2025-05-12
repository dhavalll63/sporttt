package auth

import (
	"time"

	"gorm.io/gorm"
)

type OTP struct {
	gorm.Model
	Phone     string    `gorm:"not null;index"`
	Code      string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null" json:"expired_at"`
	Verified  bool      `gorm:"default:false"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
	Phone    string   `json:"phone"`
}

type AccessReq struct {
	RefreshToken string `json:"refresh_token"`
}

type OTPReq struct {
	Phone string `json:"phone" binding:"required"`
}
type OTPVerifyRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}
