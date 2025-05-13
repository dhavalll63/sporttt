package auth

import (
	"time"

	"github.com/DhavalSuthar-24/miow/config"
	"github.com/DhavalSuthar-24/miow/internal/user"
)

// Repository interface for auth operations
type AuthRepository interface {
	CreateUser(user *user.User) error
	GetUserByEmail(email string) (*user.User, error)
	GetUserByPhone(phone string) (*user.User, error)
	SaveOTP(otp *OTP) error
	GetOTP(phone, code string) (*OTP, error)
	GetLatestOTP(phone string) (*OTP, error)
	UpdateOTP(otp *OTP) error
	SaveRefreshToken(token *user.RefreshToken) error
	GetRefreshToken(token string) (*user.RefreshToken, error)
	InvalidateRefreshToken(token string) error
	UpdateUser(user *user.User) error
}

type authRepo struct{}

// NewAuthRepository creates a new auth repository
func NewAuthRepository() AuthRepository {
	return &authRepo{}
}

// CreateUser creates a new user in the database
func (r *authRepo) CreateUser(user *user.User) error {
	return config.DB.Create(user).Error
}

// GetUserByEmail retrieves a user by email
func (r *authRepo) GetUserByEmail(email string) (*user.User, error) {
	var u user.User
	err := config.DB.Where("email = ?", email).Preload("Roles").First(&u).Error
	return &u, err
}

// GetUserByPhone retrieves a user by phone number
func (r *authRepo) GetUserByPhone(phone string) (*user.User, error) {
	var u user.User
	err := config.DB.Where("phone = ?", phone).Preload("Roles").First(&u).Error
	return &u, err
}

// SaveOTP saves an OTP to the database
func (r *authRepo) SaveOTP(otp *OTP) error {
	return config.DB.Create(otp).Error
}

// GetOTP retrieves an OTP by phone and code
func (r *authRepo) GetOTP(phone, code string) (*OTP, error) {
	var otp OTP
	err := config.DB.Where("phone = ? AND code = ? AND verified = false AND expires_at > ?",
		phone, code, time.Now()).First(&otp).Error
	return &otp, err
}

// GetLatestOTP retrieves the latest unexpired OTP for a phone number
func (r *authRepo) GetLatestOTP(phone string) (*OTP, error) {
	var otp OTP
	err := config.DB.Where("phone = ? AND verified = false AND expires_at > ?",
		phone, time.Now()).Order("created_at desc").First(&otp).Error
	return &otp, err
}

// UpdateOTP updates an OTP in the database
func (r *authRepo) UpdateOTP(otp *OTP) error {
	return config.DB.Save(otp).Error
}

// SaveRefreshToken saves a refresh token to the database
func (r *authRepo) SaveRefreshToken(token *user.RefreshToken) error {
	return config.DB.Create(token).Error
}

// GetRefreshToken retrieves a refresh token by token string
func (r *authRepo) GetRefreshToken(token string) (*user.RefreshToken, error) {
	var t user.RefreshToken
	err := config.DB.Where("token = ? AND revoked = false AND expires_at > ?",
		token, time.Now()).First(&t).Error
	return &t, err
}

// InvalidateRefreshToken invalidates a refresh token
func (r *authRepo) InvalidateRefreshToken(token string) error {
	return config.DB.Model(&user.RefreshToken{}).Where("token = ?", token).
		Update("revoked", true).Error
}

// UpdateUser updates a user in the database
func (r *authRepo) UpdateUser(user *user.User) error {
	return config.DB.Save(user).Error
}
