package auth

import (
	"mime/multipart"
	"time"

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

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=30" example:"john_doe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=8,max=72" example:"password123"`
	Phone    string `json:"phone" binding:"required,e164" example:"+919876543210"`                // e164 for international phone number format
	Role     string `json:"role" binding:"omitempty,oneof=player admin manager" example:"player"` // Specify allowed roles if fixed
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
	Name            *string `json:"name,omitempty" example:"John Doe"`
	Username        *string `json:"username,omitempty" binding:"omitempty,min=3,max=30" example:"john_doe_new"`
	Bio             *string `json:"bio,omitempty" example:"Avid cricketer and developer."`
	Address         *string `json:"address,omitempty" example:"123 Main St"`
	City            *string `json:"city,omitempty" example:"Mumbai"`
	District        *string `json:"district,omitempty" example:"Mumbai Suburban"`
	State           *string `json:"state,omitempty" example:"Maharashtra"`
	Country         *string `json:"country,omitempty" example:"India"`
	PostalCode      *string `json:"postal_code,omitempty" example:"400001"`
	PreferredSports *string `json:"preferred_sports,omitempty" binding:"omitempty,json" example:"[\"cricket\",\"football\"]"`
	SocialMedia     *string `json:"social_media,omitempty" binding:"omitempty,json" example:"{\"twitter\":\"@johndoe\"}"`
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

type UserResponse struct {
	ID              uint      `json:"id"`
	Name            string    `json:"name"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	Role            string    `json:"role"`
	Phone           string    `json:"phone"`
	PhoneVerified   bool      `json:"phone_verified"`
	ProfileImage    string    `json:"profile_image"`
	EmailVerified   bool      `json:"email_verified"`
	Verified        bool      `json:"verified"`
	Address         string    `json:"address,omitempty"`
	City            string    `json:"city,omitempty"`
	District        string    `json:"district,omitempty"`
	State           string    `json:"state,omitempty"`
	Country         string    `json:"country,omitempty"`
	PostalCode      string    `json:"postal_code,omitempty"`
	Coordinates     string    `json:"coordinates,omitempty"` // Assuming JSON string
	Bio             string    `json:"bio,omitempty"`
	LastActive      time.Time `json:"last_active,omitempty"`
	PreferredSports string    `json:"preferred_sports,omitempty"` // Assuming JSON string
	SocialMedia     string    `json:"social_media,omitempty"`     // Assuming JSON string
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func FilterUserRecord(u *user.User) UserResponse {
	return UserResponse{
		ID:              u.ID,
		Name:            u.Name,
		Username:        u.Username,
		Email:           u.Email,
		Role:            u.Role,
		Phone:           u.Phone,
		PhoneVerified:   u.PhoneVerified,
		ProfileImage:    u.ProfileImage,
		EmailVerified:   u.EmailVerified,
		Verified:        u.Verified,
		Address:         u.Address,
		City:            u.City,
		District:        u.District,
		State:           u.State,
		Country:         u.Country,
		PostalCode:      u.PostalCode,
		Coordinates:     u.Coordinates, // Ensure this is a string or handle conversion
		Bio:             u.Bio,
		LastActive:      u.LastActive,
		PreferredSports: u.PreferredSports, // Ensure this is a string or handle conversion
		SocialMedia:     u.SocialMedia,     // Ensure this is a string or handle conversion
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}
