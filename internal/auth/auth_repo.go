package auth

import (
	"github.com/DhavalSuthar-24/miow/config"
	"github.com/DhavalSuthar-24/miow/internal/user"
)

func CreateUser(u *user.User) error {
	return config.DB.Create(u).Error
}

func GetUserByEmail(email string) (*user.User, error) {
	var u user.User
	err := config.DB.Where("email= ?", email).First(&u).Error
	return &u, err
}

func getAllRoles() ([]user.Role, error) {
	var roles []user.Role
	err := config.DB.Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}
