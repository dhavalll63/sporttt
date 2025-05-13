// sport/model.go
package sport

import (
	"gorm.io/gorm"
)

// Sport represents a type of sport/game
type Sport struct {
	gorm.Model
	Name           string `json:"name" gorm:"unique;not null"`
	Description    string `json:"description"`
	Rules          string `json:"rules" gorm:"type:json"`
	Icon           string `json:"icon"`
	Positions      string `json:"positions" gorm:"type:json"`
	RequiredSkills string `json:"required_skills" gorm:"type:json"`
	Equipment      string `json:"equipment" gorm:"type:json"`
}

// Skill represents abilities within sports
type Skill struct {
	gorm.Model
	Name        string  `json:"name" gorm:"unique"`
	Description string  `json:"description"`
	Sport       string  `json:"sport" gorm:"index"`
	Weight      float64 `json:"weight" gorm:"default:1.0"`
}

// UserSport tracks user's sports preferences
type UserSport struct {
	gorm.Model
	UserID   uint   `json:"user_id" gorm:"index"`
	SportID  uint   `json:"sport_id" gorm:"index"`
	Position string `json:"position"`
	Level    string `json:"level"`
}
