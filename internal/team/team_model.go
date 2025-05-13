// team/model.go
package team

import (
	"time"

	"gorm.io/gorm"
)

// Team represents a sports team
type Team struct {
	gorm.Model
	Name           string    `json:"name" gorm:"not null"`
	Description    string    `json:"description"`
	Logo           string    `json:"logo"`
	CreatedByID    uint      `json:"created_by_id" gorm:"index"`
	Sport          string    `json:"sport" gorm:"index"`
	MinPlayers     int       `json:"min_players"`
	MaxPlayers     int       `json:"max_players"`
	Requirements   string    `json:"requirements" gorm:"type:json"`
	Level          string    `json:"level"`
	Achievements   string    `json:"achievements" gorm:"type:json"`
	TrophyCount    int       `json:"trophy_count" gorm:"default:0"`
	SocialLinks    string    `json:"social_links" gorm:"type:json"`
	MatchHistory   string    `json:"match_history" gorm:"type:json"`
	RegionalRank   int       `json:"regional_rank"`
	LastRankUpdate time.Time `json:"last_rank_update"`
	Rating         float64   `json:"rating" gorm:"default:1000.0"`
	IsDeleted      bool      `json:"is_deleted" gorm:"default:false"`
}

// TeamMember represents a user's membership in a team
type TeamMember struct {
	gorm.Model
	TeamID       uint      `json:"team_id" gorm:"index"`
	UserID       uint      `json:"user_id" gorm:"index"`
	Role         string    `json:"role" gorm:"default:'player'"`
	Position     string    `json:"position"`
	JoinedAt     time.Time `json:"joined_at"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	IsCaptain    bool      `json:"is_captain" gorm:"default:false"`
	JerseyNumber int       `json:"jersey_number"`
	Stats        string    `json:"stats" gorm:"type:json"`
}

// TeamInvitation for inviting users to join teams
type TeamInvitation struct {
	gorm.Model
	TeamID    uint      `json:"team_id" gorm:"index"`
	UserID    uint      `json:"user_id" gorm:"index"`
	Role      string    `json:"role" gorm:"default:'player'"`
	Position  string    `json:"position"`
	Message   string    `json:"message"`
	Status    string    `json:"status" gorm:"default:'pending'"`
	ExpiresAt time.Time `json:"expires_at"`
}

// JoinRequest for users requesting to join teams
type JoinRequest struct {
	gorm.Model
	TeamID    uint      `json:"team_id" gorm:"index"`
	UserID    uint      `json:"user_id" gorm:"index"`
	Message   string    `json:"message"`
	Status    string    `json:"status" gorm:"default:'pending'"`
	Position  string    `json:"position"`
	Skills    string    `json:"skills" gorm:"type:json"`
	ExpiresAt time.Time `json:"expires_at"`
}
