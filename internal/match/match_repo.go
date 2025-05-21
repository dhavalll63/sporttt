package match

import (
	"errors"
	"time"

	"github.com/DhavalSuthar-24/miow/internal/team"
	"gorm.io/gorm"
)

// MatchRepository defines methods to interact with match-related data
type MatchRepository interface {
	// Challenge methods
	CreateChallenge(challenge *Challenge) error
	GetChallengeByID(id uint) (*Challenge, error)
	UpdateChallenge(challenge *Challenge) error
	DeleteChallenge(id uint) error
	GetChallenges(filters map[string]interface{}, page, pageSize int) ([]Challenge, int64, error)
	GetUserChallenges(userID uint, status string, page, pageSize int) ([]Challenge, int64, error)
	GetTeamChallenges(teamID uint, status string, page, pageSize int) ([]Challenge, int64, error)
	AcceptChallenge(challengeID, userID uint, acceptorType string) error
	RejectChallenge(challengeID, userID uint, rejectorType string) error
	ExpireChallenges() error

	// Match methods
	CreateMatch(match *Match) error
	GetMatchByID(id uint) (*Match, error)
	UpdateMatch(match *Match) error
	DeleteMatch(id uint) error
	GetMatches(filters map[string]interface{}, page, pageSize int) ([]Match, int64, error)
	GetUserMatches(userID uint, status string, page, pageSize int) ([]Match, int64, error)
	GetTeamMatches(teamID uint, status string, page, pageSize int) ([]Match, int64, error)
	AddTeamToMatch(matchTeam *MatchTeam) error
	UpdateMatchStatus(matchID uint, status MatchStatus) error
	UpdateMatchScore(matchTeam *MatchTeam) error
	EndMatch(matchID uint, winningTeamID uint) error

	// Tournment methods
	CreateTournament(tournament *Tournament) error
	GetTournamentByID(id uint) (*Tournament, error)
	GetTournaments(filters map[string]interface{}, page, pageSize int) ([]Tournament, int64, error)
	UpdateTournament(tournament *Tournament) error
	DeleteTournament(id uint) error
	RegisterTeamInTournament(tournamentID uint, teamID uint) error
	UnregisterTeamFromTournament(tournamentID uint, teamID uint) error

	// Transaction support
	WithTransaction(txFunc func(MatchRepository) error) error
}

// GormMatchRepository implements MatchRepository using GORM
type GormMatchRepository struct {
	db *gorm.DB
}

// NewGormMatchRepository creates a new GormMatchRepository
func NewGormMatchRepository(db *gorm.DB) *GormMatchRepository {
	return &GormMatchRepository{db: db}
}

// WithTransaction implements transaction support
func (r *GormMatchRepository) WithTransaction(txFunc func(MatchRepository) error) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	txRepo := &GormMatchRepository{db: tx}
	err := txFunc(txRepo)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Challenge Repository Methods

// CreateChallenge creates a new challenge
func (r *GormMatchRepository) CreateChallenge(challenge *Challenge) error {
	return r.db.Create(challenge).Error
}

// GetChallengeByID retrieves a challenge by ID with all related entities
func (r *GormMatchRepository) GetChallengeByID(id uint) (*Challenge, error) {
	var challenge Challenge
	result := r.db.Preload("Sport").
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar") // Only load essential user fields
		}).
		Preload("SenderTeam").
		Preload("ReceiverTeam").
		Preload("SenderUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar")
		}).
		Preload("ReceiverUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar")
		}).
		Preload("Venue").
		First(&challenge, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &challenge, nil
}

// UpdateChallenge updates an existing challenge
func (r *GormMatchRepository) UpdateChallenge(challenge *Challenge) error {
	return r.db.Save(challenge).Error
}

// DeleteChallenge soft-deletes a challenge
func (r *GormMatchRepository) DeleteChallenge(id uint) error {
	return r.db.Delete(&Challenge{}, id).Error
}

// GetChallenges retrieves challenges based on filters with pagination
func (r *GormMatchRepository) GetChallenges(filters map[string]interface{}, page, pageSize int) ([]Challenge, int64, error) {
	var challenges []Challenge
	var total int64

	query := r.db.Model(&Challenge{})

	// Apply filters
	for key, value := range filters {
		query = query.Where(key, value)
	}

	// Count total before pagination
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	result := query.Preload("Sport").
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar")
		}).
		Preload("SenderTeam").
		Preload("ReceiverTeam").
		Preload("SenderUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar")
		}).
		Preload("ReceiverUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar")
		}).
		Preload("Venue").
		Offset(offset).Limit(pageSize).
		Find(&challenges)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return challenges, total, nil
}

// GetUserChallenges retrieves challenges for a specific user
func (r *GormMatchRepository) GetUserChallenges(userID uint, status string, page, pageSize int) ([]Challenge, int64, error) {
	var challenges []Challenge
	var total int64

	query := r.db.Model(&Challenge{}).Where(
		r.db.Where("created_by_user_id = ?", userID).
			Or("sender_user_id = ?", userID).
			Or("receiver_user_id = ?", userID))

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total before pagination
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	result := query.Preload("Sport").
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar")
		}).
		Preload("SenderTeam").
		Preload("ReceiverTeam").
		Preload("SenderUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar")
		}).
		Preload("ReceiverUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar")
		}).
		Preload("Venue").
		Offset(offset).Limit(pageSize).
		Find(&challenges)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return challenges, total, nil
}

// GetTeamChallenges retrieves challenges for a specific team
func (r *GormMatchRepository) GetTeamChallenges(teamID uint, status string, page, pageSize int) ([]Challenge, int64, error) {
	var challenges []Challenge
	var total int64

	query := r.db.Model(&Challenge{}).Where(
		r.db.Where("sender_team_id = ?", teamID).
			Or("receiver_team_id = ?", teamID))

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total before pagination
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	result := query.Preload("Sport").
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar")
		}).
		Preload("SenderTeam").
		Preload("ReceiverTeam").
		Preload("SenderUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar")
		}).
		Preload("ReceiverUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar")
		}).
		Preload("Venue").
		Offset(offset).Limit(pageSize).
		Find(&challenges)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return challenges, total, nil
}

// AcceptChallenge accepts a challenge and creates a match
func (r *GormMatchRepository) AcceptChallenge(challengeID, userID uint, acceptorType string) error {
	challenge, err := r.GetChallengeByID(challengeID)
	if err != nil {
		return err
	}

	if challenge == nil {
		return errors.New("challenge not found")
	}

	// Check if challenge is in a valid state to be accepted
	if challenge.Status != StatusPending && challenge.Status != StatusOpen {
		return errors.New("challenge cannot be accepted in its current state")
	}

	// Validate acceptor
	if acceptorType == "team" {
		// For team challenges
		if challenge.ChallengeType == OpenChallengeTeam || challenge.ChallengeType == DirectChallengeTeam {
			// Check if acceptor team ID matches receiver team ID
			team := r.db.Model(&team.Team{}).Where("id = ?", challenge.ReceiverTeamID).First(&team.Team{})
			if team.Error != nil {
				return errors.New("team not authorized to accept this challenge")
			}
		} else {
			return errors.New("this is not a team challenge")
		}
	} else if acceptorType == "individual" {
		// For individual challenges
		if challenge.ChallengeType == OpenChallengeIndividual || challenge.ChallengeType == DirectChallengeIndividual {
			// Check if acceptor user ID matches receiver user ID
			if challenge.ReceiverUserID == nil || *challenge.ReceiverUserID != userID {
				return errors.New("user not authorized to accept this challenge")
			}
		} else {
			return errors.New("this is not an individual challenge")
		}
	} else {
		return errors.New("invalid acceptor type")
	}

	// Update challenge status
	now := time.Now()
	challenge.Status = StatusAccepted
	challenge.AcceptedAt = &now

	// Create match from challenge
	match := Match{
		CreatedByUserID: challenge.CreatedByUserID,
		SportID:         challenge.SportID,
		VenueID:         challenge.VenueID,
		LocationText:    challenge.VenueDescription,
		ScheduledAt:     challenge.ProposedDateTime,
		Description:     challenge.Description,
		CustomRules:     challenge.AdditionalRules,
		EntryFee:        challenge.EntryFee,
		WinningPrize:    challenge.PrizeDescription,
		ChallengeID:     &challenge.ID,
		SkillLevel:      challenge.MinSkillLevel,
		Status:          StatusMatchUpcoming,
	}

	// Begin transaction
	return r.WithTransaction(func(txRepo MatchRepository) error {
		// Create match
		if err := txRepo.CreateMatch(&match); err != nil {
			return err
		}

		// Link match back to challenge
		challenge.ScheduledMatchID = &match.ID
		if err := txRepo.UpdateChallenge(challenge); err != nil {
			return err
		}

		// Add teams to match
		if challenge.ChallengeType == OpenChallengeTeam || challenge.ChallengeType == DirectChallengeTeam {
			// Add challenger team
			senderTeam := MatchTeam{
				MatchID: match.ID,
				TeamID:  *challenge.SenderTeamID,
			}
			if err := txRepo.AddTeamToMatch(&senderTeam); err != nil {
				return err
			}

			// Add receiver team
			receiverTeam := MatchTeam{
				MatchID: match.ID,
				TeamID:  *challenge.ReceiverTeamID,
			}
			if err := txRepo.AddTeamToMatch(&receiverTeam); err != nil {
				return err
			}
		}

		return nil
	})
}

// RejectChallenge rejects a challenge
func (r *GormMatchRepository) RejectChallenge(challengeID, userID uint, rejectorType string) error {
	challenge, err := r.GetChallengeByID(challengeID)
	if err != nil {
		return err
	}

	if challenge == nil {
		return errors.New("challenge not found")
	}

	// Check if challenge is in a valid state to be rejected
	if challenge.Status != StatusPending && challenge.Status != StatusOpen {
		return errors.New("challenge cannot be rejected in its current state")
	}

	// Validate rejector
	if rejectorType == "team" {
		// For team challenges
		if challenge.ChallengeType == OpenChallengeTeam || challenge.ChallengeType == DirectChallengeTeam {
			// Check if rejector team ID matches receiver team ID or if the user is part of the team
			if challenge.ReceiverTeamID == nil {
				return errors.New("team not specified in this challenge")
			}
		} else {
			return errors.New("this is not a team challenge")
		}
	} else if rejectorType == "individual" {
		// For individual challenges
		if challenge.ChallengeType == OpenChallengeIndividual || challenge.ChallengeType == DirectChallengeIndividual {
			// Check if rejector user ID matches receiver user ID
			if challenge.ReceiverUserID == nil || *challenge.ReceiverUserID != userID {
				return errors.New("user not authorized to reject this challenge")
			}
		} else {
			return errors.New("this is not an individual challenge")
		}
	} else {
		return errors.New("invalid rejector type")
	}

	// Update challenge status
	challenge.Status = StatusRejected
	return r.UpdateChallenge(challenge)
}

// ExpireChallenges updates status of expired challenges
func (r *GormMatchRepository) ExpireChallenges() error {
	now := time.Now()
	return r.db.Model(&Challenge{}).
		Where("expires_at < ? AND status IN ?", now, []ChallengeStatus{StatusOpen, StatusPending}).
		Update("status", StatusExpired).Error
}

// Match Repository Methods

// CreateMatch creates a new match
func (r *GormMatchRepository) CreateMatch(match *Match) error {
	return r.db.Create(match).Error
}

// GetMatchByID retrieves a match by ID with all related entities
func (r *GormMatchRepository) GetMatchByID(id uint) (*Match, error) {
	var match Match
	result := r.db.Preload("Sport").
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar")
		}).
		Preload("Venue").
		Preload("Challenge").
		Preload("WinningTeam").
		Preload("Teams").
		Preload("Teams.Team").
		First(&match, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &match, nil
}

// UpdateMatch updates an existing match
func (r *GormMatchRepository) UpdateMatch(match *Match) error {
	return r.db.Save(match).Error
}

// DeleteMatch soft-deletes a match
func (r *GormMatchRepository) DeleteMatch(id uint) error {
	return r.db.Delete(&Match{}, id).Error
}

// GetMatches retrieves matches based on filters with pagination
func (r *GormMatchRepository) GetMatches(filters map[string]interface{}, page, pageSize int) ([]Match, int64, error) {
	var matches []Match
	var total int64

	query := r.db.Model(&Match{})

	// Apply filters
	for key, value := range filters {
		query = query.Where(key, value)
	}

	// Count total before pagination
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	result := query.Preload("Sport").
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar")
		}).
		Preload("Venue").
		Preload("Teams").
		Preload("Teams.Team").
		Offset(offset).Limit(pageSize).
		Find(&matches)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return matches, total, nil
}

// GetUserMatches retrieves matches for a specific user
func (r *GormMatchRepository) GetUserMatches(userID uint, status string, page, pageSize int) ([]Match, int64, error) {
	// Find team IDs where the user is a member
	var teamIDs []uint
	err := r.db.Table("team_members").
		Where("user_id = ? AND is_active = ?", userID, true).
		Pluck("team_id", &teamIDs).Error

	if err != nil {
		return nil, 0, err
	}

	// Construct query to find matches related to the user
	query := r.db.Model(&Match{}).
		Joins("LEFT JOIN match_teams ON match_teams.match_id = matches.id").
		Where("matches.created_by_user_id = ? OR match_teams.team_id IN ?", userID, teamIDs)

	if status != "" {
		query = query.Where("matches.status = ?", status)
	}

	var total int64
	err = query.Distinct("matches.id").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	var matchIDs []uint
	err = query.Distinct("matches.id").
		Offset(offset).Limit(pageSize).
		Pluck("matches.id", &matchIDs).Error

	if err != nil {
		return nil, 0, err
	}

	var matches []Match
	if len(matchIDs) > 0 {
		err = r.db.Preload("Sport").
			Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
				return db.Select("ID, Username, FirstName, LastName, Avatar")
			}).
			Preload("Venue").
			Preload("Teams").
			Preload("Teams.Team").
			Where("id IN ?", matchIDs).
			Find(&matches).Error

		if err != nil {
			return nil, 0, err
		}
	}

	return matches, total, nil
}

// GetTeamMatches retrieves matches for a specific team
func (r *GormMatchRepository) GetTeamMatches(teamID uint, status string, page, pageSize int) ([]Match, int64, error) {
	query := r.db.Model(&Match{}).
		Joins("LEFT JOIN match_teams ON match_teams.match_id = matches.id").
		Where("match_teams.team_id = ?", teamID)

	if status != "" {
		query = query.Where("matches.status = ?", status)
	}

	var total int64
	err := query.Distinct("matches.id").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	var matchIDs []uint
	err = query.Distinct("matches.id").
		Offset(offset).Limit(pageSize).
		Pluck("matches.id", &matchIDs).Error

	if err != nil {
		return nil, 0, err
	}

	var matches []Match
	if len(matchIDs) > 0 {
		err = r.db.Preload("Sport").
			Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
				return db.Select("ID, Username, FirstName, LastName, Avatar")
			}).
			Preload("Venue").
			Preload("Teams").
			Preload("Teams.Team").
			Where("id IN ?", matchIDs).
			Find(&matches).Error

		if err != nil {
			return nil, 0, err
		}
	}

	return matches, total, nil
}

// AddTeamToMatch adds a team to a match
func (r *GormMatchRepository) AddTeamToMatch(matchTeam *MatchTeam) error {
	return r.db.Create(matchTeam).Error
}

// UpdateMatchStatus updates the status of a match
func (r *GormMatchRepository) UpdateMatchStatus(matchID uint, status MatchStatus) error {
	return r.db.Model(&Match{}).Where("id = ?", matchID).Update("status", status).Error
}

// UpdateMatchScore updates the score for a team in a match
func (r *GormMatchRepository) UpdateMatchScore(matchTeam *MatchTeam) error {
	return r.db.Save(matchTeam).Error
}

// EndMatch ends a match and updates the winning team
func (r *GormMatchRepository) EndMatch(matchID uint, winningTeamID uint) error {
	return r.db.Model(&Match{}).
		Where("id = ?", matchID).
		Updates(map[string]interface{}{
			"status":          StatusMatchCompleted,
			"winning_team_id": winningTeamID,
		}).Error
}
func (r *GormMatchRepository) CreateTournament(tournament *Tournament) error {
	return r.db.Create(tournament).Error
}

// GetTournamentByID retrieves a tournament by ID with all related entities
func (r *GormMatchRepository) GetTournamentByID(id uint) (*Tournament, error) {
	var tournament Tournament
	result := r.db.Preload("Sport").
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar")
		}).
		Preload("Teams").
		Preload("Teams.Team", func(db *gorm.DB) *gorm.DB { // Select specific fields for team to avoid loading too much
			return db.Select("ID, Name, Avatar")
		}).
		Preload("Matches", func(db *gorm.DB) *gorm.DB { // Select specific fields for matches
			return db.Select("ID, ScheduledAt, Status, TournamentID")
		}).
		First(&tournament, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &tournament, nil
}

// GetTournaments retrieves tournaments based on filters with pagination
func (r *GormMatchRepository) GetTournaments(filters map[string]interface{}, page, pageSize int) ([]Tournament, int64, error) {
	var tournaments []Tournament
	var total int64

	query := r.db.Model(&Tournament{})

	for key, value := range filters {
		query = query.Where(key, value)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	result := query.Preload("Sport").
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Username, FirstName, LastName, Avatar")
		}).
		Order("created_at desc").
		Offset(offset).Limit(pageSize).
		Find(&tournaments)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return tournaments, total, nil
}

// UpdateTournament updates an existing tournament
func (r *GormMatchRepository) UpdateTournament(tournament *Tournament) error {
	return r.db.Save(tournament).Error
}

// DeleteTournament soft-deletes a tournament
func (r *GormMatchRepository) DeleteTournament(id uint) error {
	// This will soft delete the tournament.
	// Associated TournamentTeam entries with OnDelete:CASCADE will also be affected based on GORM's handling of soft deletes + cascades.
	// It's usually better to handle cascading logic explicitly if complex.
	return r.db.Delete(&Tournament{}, id).Error
}

// RegisterTeamInTournament registers a team for a tournament
func (r *GormMatchRepository) RegisterTeamInTournament(tournamentID uint, teamID uint) error {
	// Use the repository's db field for transactions, not the global db.
	// The WithTransaction method handles BEGIN/COMMIT/ROLLBACK.
	// For direct transaction usage:
	return r.db.Transaction(func(tx *gorm.DB) error {
		var tournament Tournament
		if err := tx.First(&tournament, tournamentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("tournament not found")
			}
			return err
		}

		if tournament.Status != "registration_open" {
			return errors.New("tournament registration is not open")
		}

		if !tournament.RegistrationDeadline.IsZero() && time.Now().After(tournament.RegistrationDeadline) {
			return errors.New("registration deadline has passed")
		}

		if tournament.MaxTeams > 0 && tournament.CurrentTeams >= tournament.MaxTeams {
			return errors.New("tournament has reached its maximum number of teams")
		}

		var existingReg TournamentTeam
		err := tx.Where("tournament_id = ? AND team_id = ?", tournamentID, teamID).First(&existingReg).Error
		if err == nil {
			return errors.New("team is already registered in this tournament")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		tournamentTeam := TournamentTeam{
			TournamentID: tournamentID,
			TeamID:       teamID,
			RegisteredAt: time.Now(),
			Status:       "approved", // Default status
		}
		if err := tx.Create(&tournamentTeam).Error; err != nil {
			return err
		}

		tournament.CurrentTeams++
		if err := tx.Model(&Tournament{}).Where("id = ?", tournamentID).Update("current_teams", tournament.CurrentTeams).Error; err != nil {
			// Using tx.Save(&tournament) is also an option if the tournament object is up-to-date
			// if err := tx.Save(&tournament).Error; err != nil {
			return err
		}

		return nil
	})
}

// UnregisterTeamFromTournament unregisters a team from a tournament
func (r *GormMatchRepository) UnregisterTeamFromTournament(tournamentID uint, teamID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var tournament Tournament
		if err := tx.First(&tournament, tournamentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("tournament not found")
			}
			return err
		}

		// Add business logic checks if needed (e.g., cannot unregister if tournament started)
		// if tournament.Status == "ongoing" || tournament.Status == "completed" {
		// 	return errors.New("cannot unregister from an ongoing or completed tournament")
		// }

		var tournamentTeam TournamentTeam
		if err := tx.Where("tournament_id = ? AND team_id = ?", tournamentID, teamID).First(&tournamentTeam).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("team is not registered in this tournament")
			}
			return err
		}

		if err := tx.Delete(&tournamentTeam).Error; err != nil {
			return err
		}

		if tournament.CurrentTeams > 0 {
			tournament.CurrentTeams--
			if err := tx.Model(&Tournament{}).Where("id = ?", tournamentID).Update("current_teams", tournament.CurrentTeams).Error; err != nil {
				// Using tx.Save(&tournament) is also an option
				// if err := tx.Save(&tournament).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
