package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/DhavalSuthar-24/miow/internal/user"
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
	GetUserByUsername(username string) (*user.User, error)

	SaveOTP(otp *OTP) error
	GetOTP(phone, code string) (*OTP, error)
	UpdateOTP(otp *OTP) error
	GetLatestOTP(phone string) (*OTP, error)

	SaveRefreshToken(token *user.RefreshToken) error
	GetRefreshToken(tokenString string) (*user.RefreshToken, error)
	InvalidateRefreshToken(tokenString string) error
	InvalidateAllRefreshTokensForUser(UserID uint) error
	DeleteRefreshToken(tokenString string) error

	AssignRoleToUser(userID uint, role string) error
	GetUserRoles(userID uint) ([]string, error)
	RemoveRoleFromUser(userID uint, role string) error
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

func (r *authRepository) GetUserByUsername(username string) (*user.User, error) {
	var u user.User
	if err := r.db.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
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

func (r *authRepository) InvalidateAllRefreshTokensForUser(userID uint) error {
	result := r.db.Model(&user.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true)

	if result.Error != nil {
		return fmt.Errorf("failed to invalidate all refresh tokens: %w", result.Error)
	}
	return nil
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

func (r *authRepository) AssignRoleToUser(userID uint, roleName string) error {
	var role user.Role
	var existingUserRole user.UserRole
	userRole := user.UserRole{
		UserID: userID,
		RoleID: role.ID,
	}
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var user user.User
	if err := tx.First(&user, userID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("role '%s' not found", roleName)
		}
		return fmt.Errorf("failed to find role: %w", err)
	}

	if err := tx.Where("user_id = ? AND role_id = ?", userID, role.ID).First(&existingUserRole).Error; err == nil {
		tx.Rollback()
		return nil // User already has this role, no error but no action needed
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return fmt.Errorf("failed to check existing user role: %w", err)
	}

	if err := tx.Create(&userRole).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
func (r *authRepository) GetUserRoles(userID uint) ([]string, error) {
	var roles []string
	err := r.db.Model(&user.UserRole{}).
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Pluck("roles.name", &roles).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}
func (r *authRepository) RemoveRoleFromUser(userID uint, roleName string) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var role user.Role
	if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("role '%s' not found", roleName)
		}
		return fmt.Errorf("failed to find role: %w", err)
	}

	result := tx.Where("user_id = ? AND role_id = ?", userID, role.ID).Delete(&user.UserRole{})
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove role from user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("user did not have the specified role")
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
