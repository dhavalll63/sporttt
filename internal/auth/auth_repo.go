package auth

import (
	"time"

	"github.com/DhavalSuthar-24/miow/internal/user" // Assuming your user model is here
	"gorm.io/gorm"
)

type AuthRepository interface {
	CreateUser(u *user.User) error
	GetUserByEmail(email string) (*user.User, error)
	GetUserByPhone(phone string) (*user.User, error)
	GetUserByID(id uint) (*user.User, error)
	UpdateUser(u *user.User) error
	GetUserByResetToken(token string) (*user.User, error)
	GetUserByVerifyToken(token string) (*user.User, error)

	SaveOTP(otp *OTP) error
	GetOTP(phone, code string) (*OTP, error)
	UpdateOTP(otp *OTP) error
	GetLatestOTP(phone string) (*OTP, error)

	SaveRefreshToken(token *user.RefreshToken) error
	GetRefreshToken(tokenString string) (*user.RefreshToken, error)
	InvalidateRefreshToken(tokenString string) error
	DeleteRefreshToken(tokenString string) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) CreateUser(u *user.User) error {
	return r.db.Create(u).Error
}

func (r *authRepository) GetUserByEmail(email string) (*user.User, error) {
	var u user.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *authRepository) GetUserByPhone(phone string) (*user.User, error) {
	var u user.User
	if err := r.db.Where("phone = ?", phone).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *authRepository) GetUserByID(id uint) (*user.User, error) {
	var u user.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *authRepository) UpdateUser(u *user.User) error {
	return r.db.Save(u).Error
}

func (r *authRepository) GetUserByResetToken(token string) (*user.User, error) {
	var u user.User
	if err := r.db.Where("reset_token = ? AND reset_expires > ?", token, time.Now()).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *authRepository) GetUserByVerifyToken(token string) (*user.User, error) {
	var u user.User
	if err := r.db.Where("verify_token = ? AND verify_expires > ?", token, time.Now()).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *authRepository) SaveOTP(otp *OTP) error {
	return r.db.Create(otp).Error
}

func (r *authRepository) GetOTP(phone, code string) (*OTP, error) {
	var otp OTP
	if err := r.db.Where("phone = ? AND code = ? AND expires_at > ? AND verified = ?", phone, code, time.Now(), false).First(&otp).Error; err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *authRepository) UpdateOTP(otp *OTP) error {
	return r.db.Save(otp).Error
}

func (r *authRepository) GetLatestOTP(phone string) (*OTP, error) {
	var otp OTP
	if err := r.db.Where("phone = ? AND expires_at > ?", phone, time.Now()).Order("created_at DESC").First(&otp).Error; err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *authRepository) SaveRefreshToken(token *user.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *authRepository) GetRefreshToken(tokenString string) (*user.RefreshToken, error) {
	var rt user.RefreshToken
	if err := r.db.Where("token = ? AND expires_at > ? AND revoked = ?", tokenString, time.Now(), false).First(&rt).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *authRepository) InvalidateRefreshToken(tokenString string) error {
	return r.db.Model(&user.RefreshToken{}).Where("token = ?", tokenString).Update("revoked", true).Error
}

func (r *authRepository) DeleteRefreshToken(tokenString string) error {
	return r.db.Where("token = ?", tokenString).Delete(&user.RefreshToken{}).Error
}

// You might also need a way to get Roles if not handled elsewhere
func (r *authRepository) GetRoleByName(roleName string) (*user.Role, error) {
	var role user.Role
	if err := r.db.Where("name = ?", roleName).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
