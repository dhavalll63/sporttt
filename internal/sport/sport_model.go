package sport

import (
	"time"
)

// BaseModel duplicates gorm.Model so your codegen can see it directly.
type BaseModel struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

type Sport struct {
	BaseModel
	Name           string     `json:"name" gorm:"uniqueIndex;not null;size:100"`
	Description    string     `json:"description" gorm:"type:text"`
	Rules          Rules      `json:"rules" gorm:"type:json"`
	Icon           string     `json:"icon,omitempty" gorm:"size:255"` // URL or path to icon
	Positions      Positions  `json:"positions,omitempty" gorm:"type:json"`
	Equipment      Equipments `json:"equipment,omitempty" gorm:"type:json"`
	RequiredSkills []Skill    `json:"required_skills,omitempty" gorm:"foreignKey:SportID;constraint:OnDelete:CASCADE"` // Skills associated with this sport
	IsActive       bool       `json:"is_active" gorm:"default:true"`                                                   // To soft-enable/disable sport visibility
}

type Rules struct {
	MaxPlayers   int    `json:"max_players,omitempty"`
	MinPlayers   int    `json:"min_players,omitempty"`
	GameDuration string `json:"game_duration,omitempty"` // e.g., "90 minutes", "4 quarters of 12 minutes"
	Other        string `json:"other,omitempty"`
}

// Position defines a player position within a sport.
type Position struct {
	Name         string `json:"name" gorm:"size:100"`
	Description  string `json:"description,omitempty" gorm:"type:text"`
	Abbreviation string `json:"abbreviation,omitempty" gorm:"size:10"`
}

// Equipment defines a piece of equipment used in a sport.
type Equipment struct {
	Name        string `json:"name" gorm:"size:100"`
	Description string `json:"description,omitempty" gorm:"type:text"`
	IsRequired  bool   `json:"is_required,omitempty"`
}

// Positions is a slice of Position, implementing Scanner and Valuer for GORM.
type Positions []Position

// Equipments is a slice of Equipment, implementing Scanner and Valuer for GORM.
type Equipments []Equipment

// Skill represents a specific skill related to a sport.
type Skill struct {
	BaseModel
	Name        string  `json:"name" gorm:"not null;size:100;uniqueIndex:idx_skill_sport"`
	Description string  `json:"description,omitempty" gorm:"type:text"`
	SportID     uint    `json:"sport_id" gorm:"not null;uniqueIndex:idx_skill_sport"` // Foreign key to Sport
	Sport       Sport   `json:"-" gorm:"foreignKey:SportID"`                          // Belongs to Sport (optional, for eager loading if needed)
	Weight      float64 `json:"weight,omitempty" gorm:"default:1.0"`                  // Importance factor of the skill
}

// UserSport links a user to a sport they play, including their position and level.
type UserSport struct {
	BaseModel
	UserID   uint   `json:"user_id" gorm:"not null;index:idx_user_sport,unique"`
	SportID  uint   `json:"sport_id" gorm:"not null;index:idx_user_sport,unique"`
	Sport    Sport  `json:"sport"`                              // Eager load sport details
	Position string `json:"position,omitempty" gorm:"size:100"` // e.g., "Forward", "Goalkeeper"
	Level    string `json:"level,omitempty" gorm:"size:50"`     // e.g., "Beginner", "Intermediate", "Advanced", "Professional"
	// User User `json:"-" gorm:"foreignKey:UserID"` // Belongs to User (if User model is in a different package, manage carefully or use IDs)
}
