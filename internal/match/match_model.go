package match

import (
	"time"

	"github.com/DhavalSuthar-24/miow/internal/sport"
	"github.com/DhavalSuthar-24/miow/internal/team"
	"github.com/DhavalSuthar-24/miow/internal/user"
	"github.com/DhavalSuthar-24/miow/internal/venue"
	"gorm.io/gorm"
)

// --- Existing Enums (ChallengeType, ChallengeStatus, MatchStatus) remain the same ---
type ChallengeType string

const (
	OpenChallengeTeam         ChallengeType = "open_team"
	OpenChallengeIndividual   ChallengeType = "open_individual"
	DirectChallengeTeam       ChallengeType = "direct_team"
	DirectChallengeIndividual ChallengeType = "direct_individual"
)

type ChallengeStatus string

const (
	StatusOpen      ChallengeStatus = "open"
	StatusPending   ChallengeStatus = "pending"
	StatusAccepted  ChallengeStatus = "accepted"
	StatusRejected  ChallengeStatus = "rejected"
	StatusExpired   ChallengeStatus = "expired"
	StatusCancelled ChallengeStatus = "cancelled"
	StatusCompleted ChallengeStatus = "completed"
)

type MatchStatus string

const (
	StatusMatchPending   MatchStatus = "pending"
	StatusMatchUpcoming  MatchStatus = "upcoming"
	StatusMatchPreToss   MatchStatus = "pre_toss"  // Added: Teams decided, waiting for toss
	StatusMatchTossDone  MatchStatus = "toss_done" // Added: Toss done, waiting for play to start
	StatusMatchLive      MatchStatus = "live"
	StatusMatchCompleted MatchStatus = "completed"
	StatusMatchCancelled MatchStatus = "cancelled"
	StatusMatchPostponed MatchStatus = "postponed"
	StatusMatchForfeited MatchStatus = "forfeited"
	StatusMatchAbandoned MatchStatus = "abandoned" // Added: Match abandoned (e.g. rain)
)

// DismissalType for cricket wickets
type DismissalType string

const (
	DismissalTypeBowled      DismissalType = "bowled"
	DismissalTypeCaught      DismissalType = "caught"
	DismissalTypeLBW         DismissalType = "lbw"
	DismissalTypeRunOut      DismissalType = "run_out"
	DismissalTypeStumped     DismissalType = "stumped"
	DismissalTypeHitWicket   DismissalType = "hit_wicket"
	DismissalTypeHandledBall DismissalType = "handled_ball"
	DismissalTypeObstructing DismissalType = "obstructing_the_field"
	DismissalTypeTimedOut    DismissalType = "timed_out"
	DismissalTypeRetiredHurt DismissalType = "retired_hurt"
	DismissalTypeRetiredOut  DismissalType = "retired_out" // Different from retired hurt
	DismissalTypeNotOut      DismissalType = "not_out"     // For batsmen remaining at the end
)

// ExtraType for runs not scored off the bat
type ExtraType string

const (
	ExtraWide    ExtraType = "wide"
	ExtraNoBall  ExtraType = "no_ball"
	ExtraBye     ExtraType = "bye"
	ExtraLegBye  ExtraType = "leg_bye"
	ExtraPenalty ExtraType = "penalty"
)

// --- Existing Challenge model (seems okay for setting up matches) ---
type Challenge struct {
	gorm.Model
	Title           string      `json:"title" gorm:"not null"`
	Description     string      `json:"description" gorm:"type:text"`
	SportID         uint        `json:"sport_id" gorm:"index;not null"`
	Sport           sport.Sport `gorm:"foreignKey:SportID"`
	CreatedByUserID uint        `json:"created_by_user_id" gorm:"index;not null"`
	CreatedByUser   user.User   `gorm:"foreignKey:CreatedByUserID"`

	ChallengeType ChallengeType   `json:"challenge_type" gorm:"index;not null;default:'open_team'"`
	Status        ChallengeStatus `json:"status" gorm:"index;not null;default:'open'"`

	SenderTeamID   *uint      `json:"sender_team_id,omitempty" gorm:"index"`
	SenderTeam     *team.Team `gorm:"foreignKey:SenderTeamID"`
	ReceiverTeamID *uint      `json:"receiver_team_id,omitempty" gorm:"index"`
	ReceiverTeam   *team.Team `gorm:"foreignKey:ReceiverTeamID"`

	SenderUserID   *uint      `json:"sender_user_id,omitempty" gorm:"index"`
	SenderUser     *user.User `gorm:"foreignKey:SenderUserID"`
	ReceiverUserID *uint      `json:"receiver_user_id,omitempty" gorm:"index"`
	ReceiverUser   *user.User `gorm:"foreignKey:ReceiverUserID"`

	ProposedDateTime time.Time    `json:"proposed_date_time"`
	VenueID          *uint        `json:"venue_id,omitempty" gorm:"index"`
	Venue            *venue.Venue `gorm:"foreignKey:VenueID"`
	VenueDescription string       `json:"venue_description,omitempty" gorm:"type:text"`
	LocationDetails  string       `json:"location_details,omitempty" gorm:"type:text"`

	EntryFee         float64 `json:"entry_fee,omitempty"`
	PrizeDescription string  `json:"prize_description,omitempty" gorm:"type:text"`
	MinSkillLevel    string  `json:"min_skill_level,omitempty"`
	MaxSkillLevel    string  `json:"max_skill_level,omitempty"`
	TeamSize         *int    `json:"team_size,omitempty"`
	AdditionalRules  string  `json:"additional_rules,omitempty" gorm:"type:json"`

	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	AcceptedAt       *time.Time `json:"accepted_at,omitempty"`
	ScheduledMatchID *uint      `json:"scheduled_match_id,omitempty" gorm:"index;unique"`
}

// Match represents a sports game. Enhanced for pre-toss and live scoring.
type Match struct {
	gorm.Model
	CreatedByUserID uint         `json:"created_by_user_id" gorm:"index"`
	CreatedByUser   user.User    `gorm:"foreignKey:CreatedByUserID"`
	SportID         uint         `json:"sport_id" gorm:"index;not null"`
	Sport           sport.Sport  `gorm:"foreignKey:SportID"`
	VenueID         *uint        `json:"venue_id,omitempty" gorm:"index"`
	Venue           *venue.Venue `gorm:"foreignKey:VenueID"`
	LocationText    string       `json:"location_text,omitempty"`

	ScheduledAt time.Time  `json:"scheduled_at" gorm:"index"`
	StartedAt   *time.Time `json:"started_at,omitempty"`   // Actual start time
	CompletedAt *time.Time `json:"completed_at,omitempty"` // Actual completion time
	Duration    int        `json:"duration,omitempty"`     // Planned duration in minutes

	Description   string      `json:"description,omitempty" gorm:"type:text"`
	CustomRules   string      `json:"custom_rules,omitempty" gorm:"type:json"` // e.g., overs per innings
	HighlightsURL string      `json:"highlights_url,omitempty"`
	EntryFee      float64     `json:"entry_fee,omitempty"`
	WinningPrize  string      `json:"winning_prize,omitempty"`
	ChallengeID   *uint       `json:"challenge_id,omitempty" gorm:"unique;index"`
	Challenge     *Challenge  `gorm:"foreignKey:ChallengeID"`
	SkillLevel    string      `json:"skill_level,omitempty"`
	Visibility    string      `json:"visibility" gorm:"default:'public'"`
	AutoMatch     bool        `json:"auto_match" gorm:"default:false"`
	Status        MatchStatus `json:"status" gorm:"index;default:'pending'"`
	StreamURL     string      `json:"stream_url,omitempty"`
	VodURL        string      `json:"vod_url,omitempty"`
	TournamentID  *uint       `json:"tournament_id,omitempty" gorm:"index"`
	// Tournament      *Tournament  `gorm:"foreignKey:TournamentID"`

	// Toss Information
	TossWinnerTeamID *uint      `json:"toss_winner_team_id,omitempty" gorm:"index"`
	TossWinnerTeam   *team.Team `gorm:"foreignKey:TossWinnerTeamID"`
	TossDecision     string     `json:"toss_decision,omitempty"` // "bat" or "bowl"

	// Match Result
	WinningTeamID   *uint      `json:"winning_team_id,omitempty" gorm:"index"`
	WinningTeam     *team.Team `gorm:"foreignKey:WinningTeamID"`
	ResultSummary   string     `json:"result_summary,omitempty" gorm:"type:text"` // e.g., "Team A won by 5 wickets"
	ManOfTheMatchID *uint      `json:"man_of_the_match_id,omitempty" gorm:"index"`
	ManOfTheMatch   *user.User `gorm:"foreignKey:ManOfTheMatchID"`

	// Scorecard and Live Data
	MatchTeams       []MatchTeam `json:"match_teams,omitempty" gorm:"foreignKey:MatchID"`
	Innings          []Inning    `json:"innings_data,omitempty" gorm:"foreignKey:MatchID"` // Detailed innings data
	CurrentInningsID *uint       `json:"current_innings_id,omitempty"`                     // To quickly identify the active innings
	// Scoreboard field (JSON) can be kept for a quick summary or derived from Innings.
	// For live updates, Innings and BallDelivery are the source of truth.
	// Scoreboard    string      `json:"scoreboard,omitempty" gorm:"type:json"`
}

// MatchTeam represents a team participating in a match.
// Lineup now references MatchPlayer for structured player info.
type MatchTeam struct {
	gorm.Model
	MatchID    uint      `json:"match_id" gorm:"index;not null"`
	Match      Match     `gorm:"foreignKey:MatchID"`
	TeamID     uint      `json:"team_id" gorm:"index;not null"`
	Team       team.Team `gorm:"foreignKey:TeamID"`
	IsHomeTeam bool      `json:"is_home_team" gorm:"default:false"`

	// Players in this match for this team
	Players []MatchPlayer `json:"players,omitempty" gorm:"foreignKey:MatchTeamID"`

	// Summary scores can be here, but detailed scores are in Innings
	// Score        int       `json:"score" gorm:"default:0"` // This might be total score if innings not used, otherwise derive
	// Wickets      int       `json:"wickets" gorm:"default:0"` // This might be total wickets if innings not used, otherwise derive
	// OversPlayed  float32   `json:"overs_played" gorm:"default:0.0"` // This might be total overs if innings not used

	ResultStatus string `json:"result_status,omitempty"`                 // "win", "loss", "draw", "tie", "no_result"
	TeamDetails  string `json:"team_details,omitempty" gorm:"type:json"` // e.g., captain for the match if different
}

// MatchPlayer defines a player's role and participation in a specific match for a team.
type MatchPlayer struct {
	gorm.Model
	MatchTeamID  uint      `json:"match_team_id" gorm:"index;not null"` // Link to MatchTeam
	MatchTeam    MatchTeam `gorm:"foreignKey:MatchTeamID"`
	UserID       uint      `json:"user_id" gorm:"index;not null"`
	User         user.User `gorm:"foreignKey:UserID"`
	Role         string    `json:"role,omitempty"` // e.g., "Captain", "WicketKeeper", "Substitute"
	IsPlayingXI  bool      `json:"is_playing_xi" gorm:"default:false"`
	IsSubstitute bool      `json:"is_substitute" gorm:"default:false"`
	BattingOrder *int      `json:"batting_order,omitempty"` // Nullable, 1-indexed
	BowlingOrder *int      `json:"bowling_order,omitempty"` // Nullable, 1-indexed
}

// Inning represents one team's batting session in a match.
type Inning struct {
	gorm.Model
	MatchID       uint      `json:"match_id" gorm:"index;not null"`
	Match         Match     `gorm:"foreignKey:MatchID"`
	InningsNumber int       `json:"innings_number" gorm:"not null"` // 1, 2, 3, 4 (e.g., for Test matches)
	BattingTeamID uint      `json:"batting_team_id" gorm:"index;not null"`
	BattingTeam   team.Team `gorm:"foreignKey:BattingTeamID"`
	BowlingTeamID uint      `json:"bowling_team_id" gorm:"index;not null"`
	BowlingTeam   team.Team `gorm:"foreignKey:BowlingTeamID"`

	Score       int     `json:"score" gorm:"default:0"`
	Wickets     int     `json:"wickets" gorm:"default:0"`
	Overs       float32 `json:"overs" gorm:"default:0.0"`            // e.g., 10.2 for 10 overs and 2 balls
	Balls       int     `json:"balls" gorm:"default:0"`              // Total legal balls bowled in this innings
	Status      string  `json:"status" gorm:"default:'not_started'"` // "not_started", "in_progress", "completed", "declared", "forfeited"
	TargetScore *int    `json:"target_score,omitempty"`              // For chasing team
	Declared    bool    `json:"declared" gorm:"default:false"`
	FollowOn    bool    `json:"follow_on" gorm:"default:false"`

	// Breakdown of Extras
	WideRuns    int `json:"wide_runs" gorm:"default:0"`
	NoBallRuns  int `json:"no_ball_runs" gorm:"default:0"`  // Runs scored off no-ball (bat or byes)
	NoBallCount int `json:"no_ball_count" gorm:"default:0"` // Actual no-balls bowled
	ByeRuns     int `json:"bye_runs" gorm:"default:0"`
	LegByeRuns  int `json:"leg_bye_runs" gorm:"default:0"`
	PenaltyRuns int `json:"penalty_runs" gorm:"default:0"`

	Deliveries    []BallDelivery `json:"deliveries,omitempty" gorm:"foreignKey:InningID"`
	FallOfWickets []FallOfWicket `json:"fall_of_wickets,omitempty" gorm:"foreignKey:InningID"`
}

// BallDelivery records every ball bowled in an innings. THIS IS THE CORE OF LIVE SCORING.
type BallDelivery struct {
	gorm.Model
	InningID         uint   `json:"inning_id" gorm:"index;not null"`
	Inning           Inning `gorm:"foreignKey:InningID"`
	OverNumber       int    `json:"over_number" gorm:"not null"`         // 1-indexed
	BallNumberInOver int    `json:"ball_number_in_over" gorm:"not null"` // 1-indexed for legal deliveries
	DeliveryInOver   int    `json:"delivery_in_over" gorm:"not null"`    // Actual delivery sequence if no-balls/wides (can be > 6)

	BowlerID     uint      `json:"bowler_id" gorm:"index;not null"`
	Bowler       user.User `gorm:"foreignKey:BowlerID"`
	StrikerID    uint      `json:"striker_id" gorm:"index;not null"`
	Striker      user.User `gorm:"foreignKey:StrikerID"`
	NonStrikerID uint      `json:"non_striker_id" gorm:"index;not null"`
	NonStriker   user.User `gorm:"foreignKey:NonStrikerID"`

	RunsScored int        `json:"runs_scored" gorm:"default:0"` // Runs scored by batsman off this ball
	IsFour     bool       `json:"is_four" gorm:"default:false"`
	IsSix      bool       `json:"is_six" gorm:"default:false"`
	IsWicket   bool       `json:"is_wicket" gorm:"default:false"`
	IsExtra    bool       `json:"is_extra" gorm:"default:false"`
	ExtraType  *ExtraType `json:"extra_type,omitempty"`        // "wide", "no_ball", "bye", "leg_bye", "penalty"
	ExtraRuns  int        `json:"extra_runs" gorm:"default:0"` // Runs from extras only
	// For no-balls, ExtraRuns = 1 (for the no-ball itself), RunsScored = runs hit by batsman. Total runs from ball = RunsScored + ExtraRuns.

	// Wicket Details (if IsWicket is true)
	DismissalType *DismissalType `json:"dismissal_type,omitempty"`
	PlayerOutID   *uint          `json:"player_out_id,omitempty" gorm:"index"`
	PlayerOut     *user.User     `gorm:"foreignKey:PlayerOutID"`
	Fielder1ID    *uint          `json:"fielder1_id,omitempty" gorm:"index"` // Catch, Runout (thrower), Stumping
	Fielder1      *user.User     `gorm:"foreignKey:Fielder1ID"`
	Fielder2ID    *uint          `json:"fielder2_id,omitempty" gorm:"index"` // Runout (receiver at stumps)
	Fielder2      *user.User     `gorm:"foreignKey:Fielder2ID"`

	Commentary      string    `json:"commentary,omitempty" gorm:"type:text"`
	Timestamp       time.Time `json:"timestamp" gorm:"autoCreateTime"`
	IsLegalDelivery bool      `json:"is_legal_delivery" gorm:"default:true"` // False for wide/no-ball (doesn't count towards 6 balls in over)
}

// FallOfWicket records when and how a wicket fell.
type FallOfWicket struct {
	gorm.Model
	InningID       uint         `json:"inning_id" gorm:"index;not null"`
	Inning         Inning       `gorm:"foreignKey:InningID"`
	PlayerOutID    uint         `json:"player_out_id" gorm:"index;not null"`
	PlayerOut      user.User    `gorm:"foreignKey:PlayerOutID"`
	ScoreAtWicket  int          `json:"score_at_wicket" gorm:"not null"`
	OversAtWicket  float32      `json:"overs_at_wicket" gorm:"not null"`
	WicketNumber   int          `json:"wicket_number" gorm:"not null"`        // 1st wicket, 2nd wicket etc.
	BallDeliveryID uint         `json:"ball_delivery_id" gorm:"index;unique"` // The ball that caused the wicket
	BallDelivery   BallDelivery `gorm:"foreignKey:BallDeliveryID"`
}

// PlayerMatchStat tracks individual player statistics for a specific match (CRICKET SPECIFIC).
// Your existing PlayerStat was generic JSON. This is more structured for cricket.
type PlayerMatchStat struct {
	gorm.Model
	UserID        uint      `json:"user_id" gorm:"index;not null"`
	User          user.User `gorm:"foreignKey:UserID"`
	MatchID       uint      `json:"match_id" gorm:"index;not null"`
	Match         Match     `gorm:"foreignKey:MatchID"`
	TeamID        uint      `json:"team_id" gorm:"index;not null"`
	Team          team.Team `gorm:"foreignKey:TeamID"`
	InningsNumber int       `json:"innings_number" gorm:"not null;default:1"` // For matches with multiple innings per player (e.g. Test)

	// Batting Stats
	RunsScored          int            `json:"runs_scored" gorm:"default:0"`
	BallsFaced          int            `json:"balls_faced" gorm:"default:0"`
	Fours               int            `json:"fours" gorm:"default:0"`
	Sixes               int            `json:"sixes" gorm:"default:0"`
	StrikeRate          float32        `json:"strike_rate" gorm:"default:0.0"` // (RunsScored / BallsFaced) * 100
	HowOut              *DismissalType `json:"how_out,omitempty"`              // If dismissed
	DismissedByBowlerID *uint          `json:"dismissed_by_bowler_id,omitempty" gorm:"index"`
	DismissedByBowler   *user.User     `gorm:"foreignKey:DismissedByBowlerID"`
	FielderID           *uint          `json:"fielder_id,omitempty" gorm:"index"` // For caught, run out, stumped
	Fielder             *user.User     `gorm:"foreignKey:FielderID"`
	NotOut              bool           `json:"not_out" gorm:"default:false"`

	// Bowling Stats
	OversBowled  float32 `json:"overs_bowled" gorm:"default:0.0"` // e.g. 4.2 overs
	BallsBowled  int     `json:"balls_bowled" gorm:"default:0"`   // Total legal balls bowled
	Maidens      int     `json:"maidens" gorm:"default:0"`
	RunsConceded int     `json:"runs_conceded" gorm:"default:0"`
	WicketsTaken int     `json:"wickets_taken" gorm:"default:0"`
	EconomyRate  float32 `json:"economy_rate" gorm:"default:0.0"` // RunsConceded / (BallsBowled/6)
	NoBalls      int     `json:"no_balls" gorm:"default:0"`
	Wides        int     `json:"wides" gorm:"default:0"`
	DotsBowled   int     `json:"dots_bowled" gorm:"default:0"`

	// Fielding Stats
	Catches         int `json:"catches" gorm:"default:0"`
	Stumpings       int `json:"stumpings" gorm:"default:0"`
	RunOutsDirect   int `json:"run_outs_direct" gorm:"default:0"`
	RunOutsAssisted int `json:"run_outs_assisted" gorm:"default:0"`

	// Other
	PlayedInMatch bool `json:"played_in_match" gorm:"default:false"` // True if part of playing XI or substituted in
}

// --- Existing Tournament & TournamentTeam models (seem okay) ---
type Tournament struct {
	gorm.Model
	Name                 string      `json:"name" gorm:"not null"`
	Description          string      `json:"description" gorm:"type:text"`
	CreatedByUserID      uint        `json:"created_by_user_id" gorm:"index"`
	CreatedByUser        user.User   `gorm:"foreignKey:CreatedByUserID"`
	SportID              uint        `json:"sport_id" gorm:"index;not null"`
	Sport                sport.Sport `gorm:"foreignKey:SportID"`
	StartDate            time.Time   `json:"start_date"`
	EndDate              time.Time   `json:"end_date"`
	RegistrationDeadline time.Time   `json:"registration_deadline"`
	Format               string      `json:"format" gorm:"default:'knockout'"`
	FormatDetails        string      `json:"format_details" gorm:"type:json"`
	PrizeDescription     string      `json:"prize_description" gorm:"type:text"`
	PrizePool            float64     `json:"prize_pool,omitempty"`
	EntryFee             float64     `json:"entry_fee,omitempty"`
	MaxTeams             int         `json:"max_teams,omitempty"`
	CurrentTeams         int         `json:"current_teams" gorm:"default:0"`
	Status               string      `json:"status" gorm:"default:'registration_open'"`
	Bracket              string      `json:"bracket,omitempty" gorm:"type:json"`
}

type TournamentTeam struct {
	gorm.Model
	TournamentID uint       `json:"tournament_id" gorm:"index;not null;uniqueIndex:idx_tournament_team_unique"`
	Tournament   Tournament `gorm:"foreignKey:TournamentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TeamID       uint       `json:"team_id" gorm:"index;not null;uniqueIndex:idx_tournament_team_unique"`
	Team         team.Team  `gorm:"foreignKey:TeamID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RegisteredAt time.Time  `json:"registered_at"`
	Status       string     `json:"status" gorm:"default:'approved'"`
}

type PlayerOverallCricketStat struct {
	gorm.Model
	UserID uint      `json:"user_id" gorm:"uniqueIndex:idx_user_sport_overall;not null"` // Link to user.User
	User   user.User `gorm:"foreignKey:UserID"`
	// SportID uint      `json:"sport_id" gorm:"uniqueIndex:idx_user_sport_overall;not null"` // If you want generic overall stats
	// Sport   sport.Sport `gorm:"foreignKey:SportID"`

	// Batting Career Stats
	BattingMatchesPlayed      int     `json:"batting_matches_played" gorm:"default:0"`
	BattingInnings            int     `json:"batting_innings" gorm:"default:0"`
	BattingRunsScored         int     `json:"batting_runs_scored" gorm:"default:0"`
	BattingBallsFaced         int     `json:"batting_balls_faced" gorm:"default:0"`
	BattingHighestScore       int     `json:"batting_highest_score" gorm:"default:0"`
	BattingHighestScoreNotOut bool    `json:"batting_highest_score_not_out" gorm:"default:false"`
	BattingAverage            float32 `json:"batting_average" gorm:"default:0.0"`
	BattingStrikeRate         float32 `json:"batting_strike_rate" gorm:"default:0.0"`
	BattingNotOuts            int     `json:"batting_not_outs" gorm:"default:0"`
	BattingFours              int     `json:"batting_fours" gorm:"default:0"`
	BattingSixes              int     `json:"batting_sixes" gorm:"default:0"`
	BattingHundreds           int     `json:"batting_hundreds" gorm:"default:0"`
	BattingFifties            int     `json:"batting_fifties" gorm:"default:0"`
	BattingDucks              int     `json:"batting_ducks" gorm:"default:0"`

	// Bowling Career Stats
	BowlingMatchesPlayed      int     `json:"bowling_matches_played" gorm:"default:0"`
	BowlingInnings            int     `json:"bowling_innings" gorm:"default:0"`
	BowlingBallsBowled        int     `json:"bowling_balls_bowled" gorm:"default:0"` // Total legal balls
	BowlingRunsConceded       int     `json:"bowling_runs_conceded" gorm:"default:0"`
	BowlingWicketsTaken       int     `json:"bowling_wickets_taken" gorm:"default:0"`
	BowlingAverage            float32 `json:"bowling_average" gorm:"default:0.0"`
	BowlingEconomyRate        float32 `json:"bowling_economy_rate" gorm:"default:0.0"`
	BowlingStrikeRate         float32 `json:"bowling_strike_rate" gorm:"default:0.0"` // Balls per wicket
	BowlingMaidens            int     `json:"bowling_maidens" gorm:"default:0"`
	BestBowlingInningsRuns    int     `json:"best_bowling_innings_runs,omitempty"`    // e.g. for 5/25, this is 25
	BestBowlingInningsWickets int     `json:"best_bowling_innings_wickets,omitempty"` // e.g. for 5/25, this is 5
	FiveWicketHauls           int     `json:"five_wicket_hauls" gorm:"default:0"`
	TenWicketHaulsMatch       int     `json:"ten_wicket_hauls_match" gorm:"default:0"` // 10 wickets in a match (sum of 2 innings)
	BowlingNoBalls            int     `json:"bowling_no_balls" gorm:"default:0"`
	BowlingWides              int     `json:"bowling_wides" gorm:"default:0"`

	// Fielding Career Stats
	FieldingMatchesPlayed int `json:"fielding_matches_played" gorm:"default:0"`
	Catches               int `json:"catches" gorm:"default:0"`
	Stumpings             int `json:"stumpings" gorm:"default:0"`
	RunOutsDirect         int `json:"run_outs_direct" gorm:"default:0"`
	RunOutsAssisted       int `json:"run_outs_assisted" gorm:"default:0"`

	// General
	MatchesPlayedTotal  int       `json:"matches_played_total" gorm:"default:0"`
	ManOfTheMatchAwards int       `json:"man_of_the_match_awards" gorm:"default:0"`
	LastUpdatedAt       time.Time `json:"last_updated_at"`
}
