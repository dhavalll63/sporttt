package user

import (
	"time"

	"github.com/DhavalSuthar-24/miow/internal/models"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name            string             `json:"name" gorm:"not null"`
	Username        string             `json:"username" gorm:"unique"`
	Email           string             `json:"email" gorm:"uniqueIndex;not null"`
	Password        string             `json:"-" gorm:"not null"`
	UserRoles       []UserRole         `json:"roles" gorm:"foreignKey:UserID"`
	Phone           string             `json:"phone" gorm:"uniqueIndex;not null"`
	PhoneVerified   bool               `json:"phone_verified" gorm:"default:false"`
	ProfileImage    string             `json:"profile_image"`
	EmailVerified   bool               `json:"email_verified" gorm:"default:false"`
	Verified        bool               `json:"verified" gorm:"default:false"`
	Address         string             `json:"address"`
	City            string             `json:"city"`
	District        string             `json:"district"`
	State           string             `json:"state"`
	Country         string             `json:"country"`
	PostalCode      string             `json:"postal_code"`
	Bio             string             `json:"bio"`
	LastActive      time.Time          `json:"last_active"`
	ResetToken      string             `json:"-"`
	ResetExpires    *time.Time         `json:"-"`
	VerifyToken     string             `json:"-"`
	VerifyExpires   *time.Time         `json:"-"`
	Coordinates     models.Coordinates `json:"coordinates,omitempty" gorm:"type:jsonb;default:'{}'"`
	PreferredSports models.StringSlice `json:"preferred_sports,omitempty" gorm:"type:jsonb;default:'{}'"`
	SocialMedia     models.SocialMedia `json:"social_media,omitempty" gorm:"type:jsonb;default:'{}'"`
	RefreshTokens   []RefreshToken     `json:"-" gorm:"foreignKey:UserID"`
}

type Role struct {
	gorm.Model
	Name        string `gorm:"unique;not null"`
	Description string
	Users       []UserRole `gorm:"foreignKey:RoleID"`
}

type UserRole struct {
	gorm.Model
	UserID uint `gorm:"index"`
	RoleID uint `gorm:"index"`
	User   User
	Role   Role
}

type RefreshToken struct {
	gorm.Model
	UserID     uint      `gorm:"not null;index"`
	Token      string    `gorm:"not null;uniqueIndex"`
	DeviceID   string    `gorm:"size:255;index"`
	DeviceName string    `gorm:"size:255"`
	IPAddress  string    `gorm:"size:45"`
	UserAgent  string    `gorm:"size:512"`
	ExpiresAt  time.Time `gorm:"not null"`
	Revoked    bool      `gorm:"default:false"`
	LastUsedAt time.Time
	User       User `gorm:"foreignKey:UserID"`
}

type UserSkill interface {
	GetUserID() uint
	GetSkillID() uint
	GetSportID() uint
	GetLevel() string
}
