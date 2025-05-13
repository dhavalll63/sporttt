// user/model.go
package user

import (
	"time"

	"gorm.io/gorm"
)

// User represents the core user entity
type User struct {
	gorm.Model
	Name          string    `json:"name" gorm:"not null"`
	Username      string    `json:"username" gorm:"unique"`
	Email         string    `json:"email" gorm:"uniqueIndex;not null"`
	Password      string    `json:"-" gorm:"not null"` // Password is not exposed in JSON
	Role          string    `json:"role" gorm:"type:varchar(20);default:'player'"`
	Phone         string    `json:"phone" gorm:"uniqueIndex;not null"`
	PhoneVerified bool      `json:"phone_verified" gorm:"default:false"`
	ProfileImage  string    `json:"profile_image"`
	EmailVerified bool      `json:"email_verified" gorm:"default:false"`
	Verified      bool      `json:"verified" gorm:"default:false"`
	Address       string    `json:"address"`
	City          string    `json:"city"`
	District      string    `json:"district"`
	State         string    `json:"state"`
	Country       string    `json:"country"`
	PostalCode    string    `json:"postal_code"`
	Coordinates   string    `json:"coordinates" gorm:"type:json"`
	Bio           string    `json:"bio"`
	LastActive    time.Time `json:"last_active"`
	// Reset and verification fields
	ResetToken      string     `json:"-"`
	ResetExpires    *time.Time `json:"-"`
	VerifyToken     string     `json:"-"`
	VerifyExpires   *time.Time `json:"-"`
	PreferredSports string     `json:"preferred_sports" gorm:"type:json"`
	SocialMedia     string     `json:"social_media" gorm:"type:json"`
}

// Role defines different user roles in the system
type Role struct {
	gorm.Model
	Name        string `gorm:"unique;not null"`
	Description string
}

// OTP model for phone verification
type OTP struct {
	gorm.Model
	Phone     string    `gorm:"not null;index"`
	Code      string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null" json:"expired_at"`
	Verified  bool      `gorm:"default:false"`
}

// Token model for storing refresh tokens
type RefreshToken struct {
	gorm.Model
	UserID    uint      `gorm:"not null;index"`
	Token     string    `gorm:"not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null"`
	Revoked   bool      `gorm:"default:false"`
}
