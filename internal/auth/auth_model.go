package auth

import (
	"mime/multipart"
	"time"

	"github.com/DhavalSuthar-24/miow/internal/models"
	"github.com/DhavalSuthar-24/miow/internal/user"
	"gorm.io/gorm"
)

type OTP struct {
	gorm.Model
	Phone     string    `gorm:"not null;index"`
	Code      string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Verified  bool      `gorm:"default:false"`
	Attempt   int       `gorm:"default:0"`
}

type LoginRequest struct {
	LoginIdentifier string `json:"login_identifier" binding:"required" example:"john@example.com"` // Can be email or username
	Password        string `json:"password" binding:"required" example:"password123"`
}

type OTPRequest struct {
	Phone string `json:"phone" binding:"required,e164" example:"+919876543210"`
}

type VerifyOTPRequest struct {
	Phone string `json:"phone" binding:"required,e164" example:"+919876543210"`
	Code  string `json:"code" binding:"required,len=6" example:"123456"` // Assuming 6 digit OTP
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token" binding:"required" example:"reset-token-123456"`
	Password        string `json:"password" binding:"required,min=8,max=72" example:"newpassword123"`
	PasswordConfirm string `json:"password_confirm" binding:"required,eqfield=Password" example:"newpassword123"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required" example:"email-verify-token-123"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email" example:"johndoe@example.com"`
}

type UpdateProfileRequest struct {
	Name            *string             `json:"name,omitempty" example:"John Doe"`
	Username        *string             `json:"username,omitempty" binding:"omitempty,min=3,max=30" example:"john_doe_new"`
	Bio             *string             `json:"bio,omitempty" example:"Avid cricketer and developer."`
	Address         *string             `json:"address,omitempty" example:"123 Main St"`
	City            *string             `json:"city,omitempty" example:"Mumbai"`
	District        *string             `json:"district,omitempty" example:"Mumbai Suburban"`
	State           *string             `json:"state,omitempty" example:"Maharashtra"`
	Country         *string             `json:"country,omitempty" example:"India"`
	PostalCode      *string             `json:"postal_code,omitempty" example:"400001"`
	PreferredSports []string            `json:"preferred_sports,omitempty"`
	Coordinates     *models.Coordinates `json:"coordinates,omitempty"`
	SocialMedia     *models.SocialMedia `json:"social_media,omitempty"`
}

type UpdateProfileImageRequest struct {
	Image *multipart.FileHeader `form:"image" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=72"`
	PasswordConfirm string `json:"password_confirm" binding:"required,eqfield=NewPassword"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type RegisterRequest struct {
	Name            string              `json:"name" binding:"required"`
	Username        string              `json:"username" binding:"required"`
	Email           string              `json:"email" binding:"required,email"`
	Password        string              `json:"password" binding:"required,min=8"`
	Phone           string              `json:"phone" binding:"required"`
	Address         string              `json:"address,omitempty"`
	City            string              `json:"city,omitempty"`
	Roles           []string            `json:"role,omitempty"`
	District        string              `json:"district,omitempty"`
	State           string              `json:"state,omitempty"`
	Country         string              `json:"country,omitempty"`
	PostalCode      string              `json:"postal_code,omitempty"`
	Bio             string              `json:"bio,omitempty"`
	PreferredSports []string            `json:"preferred_sports,omitempty"`
	SocialMedia     *models.SocialMedia `json:"social_media,omitempty"`
	Coordinates     *models.Coordinates `json:"coordinates,omitempty"`
}

type UserResponse struct {
	ID              uint               `json:"id"`
	Name            string             `json:"name"`
	Username        string             `json:"username"`
	Email           string             `json:"email"`
	Phone           string             `json:"phone"`
	PhoneVerified   bool               `json:"phone_verified"`
	ProfileImage    string             `json:"profile_image"`
	EmailVerified   bool               `json:"email_verified"`
	Verified        bool               `json:"verified"`
	Address         string             `json:"address"`
	City            string             `json:"city"`
	District        string             `json:"district"`
	State           string             `json:"state"`
	Country         string             `json:"country"`
	PostalCode      string             `json:"postal_code"`
	Bio             string             `json:"bio"`
	LastActive      time.Time          `json:"last_active"`
	Coordinates     models.Coordinates `json:"coordinates"`
	PreferredSports []string           `json:"preferred_sports"`
	SocialMedia     models.SocialMedia `json:"social_media"`
	Roles           []string           `json:"roles"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

type LogoutRequest struct {
	RefreshToken          string `json:"refresh_token"`           // Optional: specific token to invalidate
	InvalidateAllSessions bool   `json:"invalidate_all_sessions"` // If true, invalidate all user's sessions
}

func FilterUserRecord(user *user.User) UserResponse {
	var roles []string
	for _, userRole := range user.UserRoles {
		roles = append(roles, userRole.Role.Name)
	}

	return UserResponse{
		ID:              user.ID,
		Name:            user.Name,
		Username:        user.Username,
		Email:           user.Email,
		Phone:           user.Phone,
		PhoneVerified:   user.PhoneVerified,
		ProfileImage:    user.ProfileImage,
		EmailVerified:   user.EmailVerified,
		Verified:        user.Verified,
		Address:         user.Address,
		City:            user.City,
		District:        user.District,
		State:           user.State,
		Country:         user.Country,
		PostalCode:      user.PostalCode,
		Bio:             user.Bio,
		LastActive:      user.LastActive,
		Coordinates:     user.Coordinates,
		PreferredSports: user.PreferredSports,
		SocialMedia:     user.SocialMedia,
		Roles:           roles,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}
