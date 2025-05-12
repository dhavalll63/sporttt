package user

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique" json:"username"`
	Email    string `gorm:"unique" json:"email"`
	Password string `json:"-"`
	Roles    []Role `gorm:"many2many:user_roles" json:"roles"`
	Phone    string `gorm:"uniqueIndex;not null" json:"phone"`
}

type Role struct {
	gorm.Model
	Name string `gorm:"unique;not null"`
}
