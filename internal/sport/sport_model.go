package sport

import "gorm.io/gorm"

type Sport struct {
	gorm.Model
	Name string `gorm:"not null;unique"`
	Icon string `json:"icon"`
}
type UserSport struct {
	gorm.Model
	UserID   uint
	SportID  uint
	Position string `json:"position"`
}
