// match/model.go
package match

import (
	"time"

	"gorm.io/gorm"
)

// Match represents a sports game between teams
type Match struct {
	gorm.Model
	CreatedByID  uint      `json:"created_by_id" gorm:"index"`
	GameID       uint      `json:"game_id" gorm:"index"`
	MatchType    string    `json:"match_type"`
	VenueID      uint      `json:"venue_id" gorm:"index"`
	ScheduledAt  time.Time `json:"scheduled_at"`
	Duration     int       `json:"duration"` // in minutes
	Location     string    `json:"location"`
	CustomRules  string    `json:"custom_rules" gorm:"type:json"`
	Scoreboard   string    `json:"scoreboard" gorm:"type:json"`
	Highlights   string    `json:"highlights"`
	ChallengeID  uint      `json:"challenge_id" gorm:"unique;index"`
	SkillLevel   string    `json:"skill_level"`
	Visibility   string    `json:"visibility" gorm:"default:'public'"`
	AutoMatch    bool      `json:"auto_match" gorm:"default:false"`
	LocationType string    `json:"location_type"`
	Status       string    `json:"status" gorm:"default:'pending'"`
	StreamURL    string    `json:"stream_url"`
	VodURL       string    `json:"vod_url"`
	TournamentID uint      `json:"tournament_id" gorm:"index"`
}

// MatchTeam represents a team participating in a match
type MatchTeam struct {
	gorm.Model
	MatchID uint   `json:"match_id" gorm:"index"`
	TeamID  uint   `json:"team_id" gorm:"index"`
	Score   int    `json:"score"`
	Details string `json:"details" gorm:"type:json"`
	Lineup  string `json:"lineup" gorm:"type:json"`
}

// Challenge represents a challenge between teams or players
type Challenge struct {
	gorm.Model
	Title          string    `json:"title"`
	SenderID       uint      `json:"sender_id" gorm:"index"`
	CreatedBy      uint      `json:"created_by"`
	ReceiverID     uint      `json:"receiver_id" gorm:"index"`
	Description    string    `json:"description"`
	MatchID        uint      `json:"match_id" gorm:"index"`
	SenderTeamID   uint      `json:"sender_team_id" gorm:"index"`
	ReceiverTeamID uint      `json:"receiver_team_id" gorm:"index"`
	TeamSelection  bool      `json:"team_selection" gorm:"default:true"`
	Status         string    `json:"status" gorm:"default:'pending'"`
	ExpiresAt      time.Time `json:"expires_at"`
	CompletedAt    time.Time `json:"completed_at"`
	TeamMatchID    uint      `json:"team_match_id" gorm:"unique"`
}

// PlayerStat tracks individual player statistics
type PlayerStat struct {
	gorm.Model
	ProfileID uint   `json:"profile_id" gorm:"index"`
	MatchID   uint   `json:"match_id" gorm:"index"`
	Stats     string `json:"stats" gorm:"type:json"`
}

// Tournament represents a multi-match competition
type Tournament struct {
	gorm.Model
	Name                 string    `json:"name"`
	Description          string    `json:"description"`
	CreatedByID          uint      `json:"created_by_id" gorm:"index"`
	Sport                string    `json:"sport" gorm:"index"`
	GameID               uint      `json:"game_id" gorm:"index"`
	StartDate            time.Time `json:"start_date"`
	EndDate              time.Time `json:"end_date"`
	RegistrationDeadline time.Time `json:"registration_deadline"`
	Format               string    `json:"format" gorm:"default:'knockout'"`
	FormatDetails        string    `json:"format_details" gorm:"type:json"`
	Prize                string    `json:"prize" gorm:"type:json"`
	PrizePool            float64   `json:"prize_pool"`
	EntryFee             float64   `json:"entry_fee"`
	MaxTeams             int       `json:"max_teams"`
	CurrentTeams         int       `json:"current_teams" gorm:"default:0"`
	Status               string    `json:"status" gorm:"default:'registration_open'"`
	Bracket              string    `json:"bracket" gorm:"type:json"`
}
