package match

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/DhavalSuthar-24/miow/config"
	"github.com/DhavalSuthar-24/miow/internal/team"
	responses "github.com/DhavalSuthar-24/miow/pkg/matchresponse"
	"github.com/gin-gonic/gin"
)

// MatchController handles match-related HTTP requests
type MatchController struct {
	repo      MatchRepository
	teamRepo  team.TeamRepository
	appConfig *config.Config
}

// NewMatchController creates a new match controller
func NewMatchController(repo MatchRepository, teamRepo team.TeamRepository, appConfig *config.Config) *MatchController {
	return &MatchController{
		repo:      repo,
		teamRepo:  teamRepo,
		appConfig: appConfig,
	}
}

// --- Helper Functions for Auth ---

func getCurrentUserID(c *gin.Context) (uint, bool) {
	userIDVal, exists := c.Get("currentUserID")
	if !exists {
		return 0, false
	}
	userID, ok := userIDVal.(uint)
	return userID, ok
}

// isTeamMember checks if the user is a member of the team
func (mc *MatchController) isTeamMember(teamID, userID uint) (bool, error) {
	member, err := mc.teamRepo.GetTeamMember(teamID, userID)
	if err != nil {
		return false, err
	}
	return member != nil && member.IsActive, nil
}

// isTeamManager checks if the user is creator, captain, vice_captain or moderator of the team
func (mc *MatchController) isTeamManager(teamID, userID uint) (bool, error) {
	isCreator, err := mc.teamRepo.IsUserTeamCreator(teamID, userID)
	if err != nil {
		return false, err
	}
	if isCreator {
		return true, nil
	}

	member, err := mc.teamRepo.GetTeamMember(teamID, userID)
	if err != nil {
		return false, err
	}
	if member == nil || !member.IsActive {
		return false, nil
	}
	return member.Role == "captain" || member.Role == "vice_captain" || member.Role == "moderator" || member.IsCaptain, nil
}

// --- DTOs for requests ---

// CreateChallengeRequest defines the request payload for creating a challenge
type CreateChallengeRequest struct {
	Title            string        `json:"title" binding:"required,min=3,max=200"`
	Description      string        `json:"description" binding:"max=2000"`
	SportID          uint          `json:"sport_id" binding:"required"`
	ChallengeType    ChallengeType `json:"challenge_type" binding:"required,oneof=open_team open_individual direct_team direct_individual"`
	SenderTeamID     *uint         `json:"sender_team_id,omitempty"`
	ReceiverTeamID   *uint         `json:"receiver_team_id,omitempty"`
	SenderUserID     *uint         `json:"sender_user_id,omitempty"`
	ReceiverUserID   *uint         `json:"receiver_user_id,omitempty"`
	ProposedDateTime time.Time     `json:"proposed_date_time" binding:"required"`
	VenueID          *uint         `json:"venue_id,omitempty"`
	VenueDescription string        `json:"venue_description,omitempty"`
	LocationDetails  string        `json:"location_details,omitempty"`
	EntryFee         float64       `json:"entry_fee,omitempty"`
	PrizeDescription string        `json:"prize_description,omitempty"`
	MinSkillLevel    string        `json:"min_skill_level,omitempty"`
	MaxSkillLevel    string        `json:"max_skill_level,omitempty"`
	TeamSize         *int          `json:"team_size,omitempty"`
	AdditionalRules  string        `json:"additional_rules,omitempty"`
	ExpiresAt        *time.Time    `json:"expires_at,omitempty"`
}

// UpdateChallengeRequest defines the request payload for updating a challenge
type UpdateChallengeRequest struct {
	Title            *string    `json:"title,omitempty" binding:"omitempty,min=3,max=200"`
	Description      *string    `json:"description,omitempty" binding:"omitempty,max=2000"`
	ProposedDateTime *time.Time `json:"proposed_date_time,omitempty"`
	VenueID          *uint      `json:"venue_id,omitempty"`
	VenueDescription *string    `json:"venue_description,omitempty"`
	LocationDetails  *string    `json:"location_details,omitempty"`
	EntryFee         *float64   `json:"entry_fee,omitempty"`
	PrizeDescription *string    `json:"prize_description,omitempty"`
	MinSkillLevel    *string    `json:"min_skill_level,omitempty"`
	MaxSkillLevel    *string    `json:"max_skill_level,omitempty"`
	TeamSize         *int       `json:"team_size,omitempty"`
	AdditionalRules  *string    `json:"additional_rules,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
}

// CreateDirectMatchRequest defines the request payload for creating a match directly
type CreateDirectMatchRequest struct {
	Title        string    `json:"title" binding:"required,min=3,max=200"`
	Description  string    `json:"description" binding:"max=2000"`
	SportID      uint      `json:"sport_id" binding:"required"`
	Team1ID      uint      `json:"team1_id" binding:"required"`
	Team2ID      uint      `json:"team2_id" binding:"required"`
	ScheduledAt  time.Time `json:"scheduled_at" binding:"required"`
	Duration     int       `json:"duration,omitempty"`
	VenueID      *uint     `json:"venue_id,omitempty"`
	LocationText string    `json:"location_text,omitempty"`
	EntryFee     float64   `json:"entry_fee,omitempty"`
	WinningPrize string    `json:"winning_prize,omitempty"`
	SkillLevel   string    `json:"skill_level,omitempty"`
	CustomRules  string    `json:"custom_rules,omitempty"`
	Visibility   string    `json:"visibility" binding:"omitempty,oneof=public private unlisted"`
}

// UpdateMatchRequest defines the request payload for updating a match
type UpdateMatchRequest struct {
	Description  *string    `json:"description,omitempty"`
	ScheduledAt  *time.Time `json:"scheduled_at,omitempty"`
	Duration     *int       `json:"duration,omitempty"`
	VenueID      *uint      `json:"venue_id,omitempty"`
	LocationText *string    `json:"location_text,omitempty"`
	CustomRules  *string    `json:"custom_rules,omitempty"`
	SkillLevel   *string    `json:"skill_level,omitempty"`
	Visibility   *string    `json:"visibility,omitempty" binding:"omitempty,oneof=public private unlisted"`
	StreamURL    *string    `json:"stream_url,omitempty"`
	VodURL       *string    `json:"vod_url,omitempty"`
}

// UpdateMatchScoreRequest defines the request payload for updating match scores
type UpdateMatchScoreRequest struct {
	TeamID       uint   `json:"team_id" binding:"required"`
	Score        int    `json:"score" binding:"required"`
	ResultStatus string `json:"result_status,omitempty"`
}

// CreateTournamentRequest defines the request payload for creating a tournament
type CreateTournamentRequest struct {
	Name                 string    `json:"name" binding:"required,min=3,max=200"`
	Description          string    `json:"description" binding:"max=2000"`
	SportID              uint      `json:"sport_id" binding:"required"`
	StartDate            time.Time `json:"start_date" binding:"required"`
	EndDate              time.Time `json:"end_date" binding:"required"`
	RegistrationDeadline time.Time `json:"registration_deadline" binding:"required"`
	Format               string    `json:"format" binding:"required,oneof=knockout round-robin league"`
	FormatDetails        string    `json:"format_details,omitempty"`
	PrizeDescription     string    `json:"prize_description,omitempty"`
	PrizePool            float64   `json:"prize_pool,omitempty"`
	EntryFee             float64   `json:"entry_fee,omitempty"`
	MaxTeams             int       `json:"max_teams" binding:"required,min=2"`
}

// UpdateTournamentRequest defines the request payload for updating a tournament
type UpdateTournamentRequest struct {
	Name                 *string    `json:"name,omitempty" binding:"omitempty,min=3,max=200"`
	Description          *string    `json:"description,omitempty" binding:"omitempty,max=2000"`
	StartDate            *time.Time `json:"start_date,omitempty"`
	EndDate              *time.Time `json:"end_date,omitempty"`
	RegistrationDeadline *time.Time `json:"registration_deadline,omitempty"`
	Format               *string    `json:"format,omitempty" binding:"omitempty,oneof=knockout round-robin league"`
	FormatDetails        *string    `json:"format_details,omitempty"`
	PrizeDescription     *string    `json:"prize_description,omitempty"`
	PrizePool            *float64   `json:"prize_pool,omitempty"`
	EntryFee             *float64   `json:"entry_fee,omitempty"`
	MaxTeams             *int       `json:"max_teams,omitempty" binding:"omitempty,min=2"`
	Status               *string    `json:"status,omitempty" binding:"omitempty,oneof=registration_open upcoming ongoing completed cancelled"`
}

// --- Challenge Controller Methods ---

// CreateChallenge handles the creation of a new challenge
func (mc *MatchController) CreateChallenge(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationErrorResponse(c, err)
		return
	}

	// Validate challenge type and required fields
	if err := mc.validateChallengeRequest(req, userID); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Create challenge object
	challenge := Challenge{
		Title:            req.Title,
		Description:      req.Description,
		SportID:          req.SportID,
		CreatedByUserID:  userID,
		ChallengeType:    req.ChallengeType,
		SenderTeamID:     req.SenderTeamID,
		ReceiverTeamID:   req.ReceiverTeamID,
		SenderUserID:     req.SenderUserID,
		ReceiverUserID:   req.ReceiverUserID,
		ProposedDateTime: req.ProposedDateTime,
		VenueID:          req.VenueID,
		VenueDescription: req.VenueDescription,
		LocationDetails:  req.LocationDetails,
		EntryFee:         req.EntryFee,
		PrizeDescription: req.PrizeDescription,
		MinSkillLevel:    req.MinSkillLevel,
		MaxSkillLevel:    req.MaxSkillLevel,
		TeamSize:         req.TeamSize,
		AdditionalRules:  req.AdditionalRules,
		ExpiresAt:        req.ExpiresAt,
	}

	// Set challenge status based on type
	if req.ChallengeType == OpenChallengeTeam || req.ChallengeType == OpenChallengeIndividual {
		challenge.Status = StatusOpen
	} else {
		challenge.Status = StatusPending
	}

	// Save challenge
	if err := mc.repo.CreateChallenge(&challenge); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to create challenge: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusCreated, gin.H{
		"message":   "Challenge created successfully",
		"challenge": challenge,
	})
}

// validateChallengeRequest validates the challenge request based on challenge type
func (mc *MatchController) validateChallengeRequest(req CreateChallengeRequest, userID uint) error {
	// Validate common fields

	// Check authorization and validate team-specific fields
	switch req.ChallengeType {
	case OpenChallengeTeam:
		// For open team challenges, sender team must be provided and user must be a manager
		if req.SenderTeamID == nil {
			return errors.New("sender team ID is required for team challenges")
		}
		isManager, err := mc.isTeamManager(*req.SenderTeamID, userID)
		if err != nil {
			return err
		}
		if !isManager {
			return errors.New("you must be a team manager to create team challenges")
		}
	case DirectChallengeTeam:
		// For direct team challenges, both sender and receiver team must be provided
		if req.SenderTeamID == nil {
			return errors.New("sender team ID is required for team challenges")
		}
		if req.ReceiverTeamID == nil {
			return errors.New("receiver team ID is required for direct team challenges")
		}
		isManager, err := mc.isTeamManager(*req.SenderTeamID, userID)
		if err != nil {
			return err
		}
		if !isManager {
			return errors.New("you must be a team manager to create team challenges")
		}
	case OpenChallengeIndividual:
		// For open individual challenges, sender user must be the current user
		if req.SenderUserID == nil {
			req.SenderUserID = &userID
		} else if *req.SenderUserID != userID {
			return errors.New("sender user ID must be your user ID")
		}
	case DirectChallengeIndividual:
		// For direct individual challenges, sender must be current user and receiver must be provided
		if req.SenderUserID == nil {
			req.SenderUserID = &userID
		} else if *req.SenderUserID != userID {
			return errors.New("sender user ID must be your user ID")
		}
		if req.ReceiverUserID == nil {
			return errors.New("receiver user ID is required for direct individual challenges")
		}
	}
	return nil
}

// GetChallenges retrieves challenges based on filters
func (mc *MatchController) GetChallenges(c *gin.Context) {
	// Parse query parameters for filters
	sportID := c.Query("sport_id")
	status := c.Query("status")
	challengeType := c.Query("type")

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Build filters
	filters := make(map[string]interface{})
	if sportID != "" {
		sportIDInt, err := strconv.Atoi(sportID)
		if err == nil {
			filters["sport_id"] = sportIDInt
		}
	}
	if status != "" {
		filters["status"] = status
	}
	if challengeType != "" {
		filters["challenge_type"] = challengeType
	}

	// Get challenges
	challenges, total, err := mc.repo.GetChallenges(filters, page, pageSize)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch challenges: "+err.Error())
		return
	}

	responses.PaginatedResponse(c, http.StatusOK, challenges, page, pageSize, total)
}

// GetChallengeByID retrieves a specific challenge by ID
func (mc *MatchController) GetChallengeByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid challenge ID")
		return
	}

	challenge, err := mc.repo.GetChallengeByID(uint(id))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch challenge: "+err.Error())
		return
	}

	if challenge == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Challenge not found")
		return
	}

	responses.SuccessResponse(c, http.StatusOK, challenge)
}

// UpdateChallenge updates an existing challenge
func (mc *MatchController) UpdateChallenge(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid challenge ID")
		return
	}

	// Get existing challenge
	challenge, err := mc.repo.GetChallengeByID(uint(id))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch challenge: "+err.Error())
		return
	}

	if challenge == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Challenge not found")
		return
	}

	// Check authorization
	if challenge.CreatedByUserID != userID {
		isAuthorized := false

		// Check if user is a team manager for the sender team
		if challenge.SenderTeamID != nil {
			isManager, err := mc.isTeamManager(*challenge.SenderTeamID, userID)
			if err == nil && isManager {
				isAuthorized = true
			}
		}

		if !isAuthorized {
			responses.ErrorResponse(c, http.StatusForbidden, "You are not authorized to update this challenge")
			return
		}
	}

	// Check if challenge is in a valid state to be updated
	if challenge.Status != StatusOpen && challenge.Status != StatusPending {
		responses.ErrorResponse(c, http.StatusBadRequest, "Cannot update challenge in its current state")
		return
	}

	var req UpdateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationErrorResponse(c, err)
		return
	}

	// Update challenge fields
	if req.Title != nil {
		challenge.Title = *req.Title
	}
	if req.Description != nil {
		challenge.Description = *req.Description
	}
	if req.ProposedDateTime != nil {
		challenge.ProposedDateTime = *req.ProposedDateTime
	}
	if req.VenueID != nil {
		challenge.VenueID = req.VenueID
	}
	if req.VenueDescription != nil {
		challenge.VenueDescription = *req.VenueDescription
	}
	if req.LocationDetails != nil {
		challenge.LocationDetails = *req.LocationDetails
	}
	if req.EntryFee != nil {
		challenge.EntryFee = *req.EntryFee
	}
	if req.PrizeDescription != nil {
		challenge.PrizeDescription = *req.PrizeDescription
	}
	if req.MinSkillLevel != nil {
		challenge.MinSkillLevel = *req.MinSkillLevel
	}
	if req.MaxSkillLevel != nil {
		challenge.MaxSkillLevel = *req.MaxSkillLevel
	}
	if req.TeamSize != nil {
		challenge.TeamSize = req.TeamSize
	}
	if req.AdditionalRules != nil {
		challenge.AdditionalRules = *req.AdditionalRules
	}
	if req.ExpiresAt != nil {
		challenge.ExpiresAt = req.ExpiresAt
	}

	if err := mc.repo.UpdateChallenge(challenge); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to update challenge: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{
		"message":   "Challenge updated successfully",
		"challenge": challenge,
	})
}

// DeleteChallenge handles deleting a challenge
func (mc *MatchController) DeleteChallenge(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid challenge ID")
		return
	}

	// Get existing challenge
	challenge, err := mc.repo.GetChallengeByID(uint(id))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch challenge: "+err.Error())
		return
	}

	if challenge == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Challenge not found")
		return
	}

	// Check authorization
	if challenge.CreatedByUserID != userID {
		isAuthorized := false

		// Check if user is a team manager for the sender team
		if challenge.SenderTeamID != nil {
			isManager, err := mc.isTeamManager(*challenge.SenderTeamID, userID)
			if err == nil && isManager {
				isAuthorized = true
			}
		}

		if !isAuthorized {
			responses.ErrorResponse(c, http.StatusForbidden, "You are not authorized to delete this challenge")
			return
		}
	}

	// Check if challenge is in a valid state to be deleted
	if challenge.Status == StatusAccepted && challenge.ScheduledMatchID != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Cannot delete a challenge that has been accepted and has a scheduled match")
		return
	}

	if err := mc.repo.DeleteChallenge(uint(id)); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete challenge: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Challenge deleted successfully",
	})
}

// GetUserChallenges retrieves all challenges related to the current user
func (mc *MatchController) GetUserChallenges(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	challenges, total, err := mc.repo.GetUserChallenges(userID, status, page, pageSize)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch challenges: "+err.Error())
		return
	}

	responses.PaginatedResponse(c, http.StatusOK, challenges, page, pageSize, total)
}

// GetTeamChallenges retrieves all challenges related to a specific team
func (mc *MatchController) GetTeamChallenges(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	teamIDStr := c.Param("teamId")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	// Check if user is a member of the team
	isMember, err := mc.isTeamMember(uint(teamID), userID)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to check team membership: "+err.Error())
		return
	}
	if !isMember {
		responses.ErrorResponse(c, http.StatusForbidden, "You must be a member of the team to view its challenges")
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	challenges, total, err := mc.repo.GetTeamChallenges(uint(teamID), status, page, pageSize)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch challenges: "+err.Error())
		return
	}

	responses.PaginatedResponse(c, http.StatusOK, challenges, page, pageSize, total)
}

// AcceptChallenge handles accepting a challenge
func (mc *MatchController) AcceptChallenge(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid challenge ID")
		return
	}

	// Get challenge
	challenge, err := mc.repo.GetChallengeByID(uint(id))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch challenge: "+err.Error())
		return
	}

	if challenge == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Challenge not found")
		return
	}

	// Determine acceptor type based on challenge type
	acceptorType := ""
	if challenge.ChallengeType == OpenChallengeTeam || challenge.ChallengeType == DirectChallengeTeam {
		acceptorType = "team"

		// Check if user is a team manager for the receiving team
		if challenge.ReceiverTeamID != nil {
			isManager, err := mc.isTeamManager(*challenge.ReceiverTeamID, userID)
			if err != nil {
				responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to check team management: "+err.Error())
				return
			}
			if !isManager {
				responses.ErrorResponse(c, http.StatusForbidden, "You must be a team manager to accept challenges")
				return
			}
		} else {
			responses.ErrorResponse(c, http.StatusBadRequest, "Invalid challenge: no receiver team specified")
			return
		}
	} else if challenge.ChallengeType == OpenChallengeIndividual || challenge.ChallengeType == DirectChallengeIndividual {
		acceptorType = "individual"

		// Check if user is the receiver
		if challenge.ReceiverUserID == nil || *challenge.ReceiverUserID != userID {
			responses.ErrorResponse(c, http.StatusForbidden, "You are not authorized to accept this challenge")
			return
		}
	} else {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid challenge type")
		return
	}

	// Accept challenge
	if err := mc.repo.AcceptChallenge(uint(id), userID, acceptorType); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to accept challenge: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Challenge accepted successfully",
	})
}

// --- Missing Controller Methods ---

// RejectChallenge handles rejecting a challenge
func (mc *MatchController) RejectChallenge(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid challenge ID")
		return
	}

	// Get challenge
	challenge, err := mc.repo.GetChallengeByID(uint(id))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch challenge: "+err.Error())
		return
	}

	if challenge == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Challenge not found")
		return
	}

	// Determine rejector type based on challenge type
	rejectorType := ""
	if challenge.ChallengeType == OpenChallengeTeam || challenge.ChallengeType == DirectChallengeTeam {
		rejectorType = "team"

		// Check if user is a team manager for the receiving team
		if challenge.ReceiverTeamID != nil {
			isManager, err := mc.isTeamManager(*challenge.ReceiverTeamID, userID)
			if err != nil {
				responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to check team management: "+err.Error())
				return
			}
			if !isManager {
				responses.ErrorResponse(c, http.StatusForbidden, "You must be a team manager to reject challenges")
				return
			}
		} else {
			responses.ErrorResponse(c, http.StatusBadRequest, "Invalid challenge: no receiver team specified")
			return
		}
	} else if challenge.ChallengeType == OpenChallengeIndividual || challenge.ChallengeType == DirectChallengeIndividual {
		rejectorType = "individual"

		// Check if user is the receiver
		if challenge.ReceiverUserID == nil || *challenge.ReceiverUserID != userID {
			responses.ErrorResponse(c, http.StatusForbidden, "You are not authorized to reject this challenge")
			return
		}
	} else {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid challenge type")
		return
	}

	// Reject challenge
	if err := mc.repo.RejectChallenge(uint(id), userID, rejectorType); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to reject challenge: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Challenge rejected successfully",
	})
}

// CancelChallenge handles canceling a challenge
func (mc *MatchController) CancelChallenge(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid challenge ID")
		return
	}

	// Get challenge
	challenge, err := mc.repo.GetChallengeByID(uint(id))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch challenge: "+err.Error())
		return
	}

	if challenge == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Challenge not found")
		return
	}

	// Check authorization - only creator or team manager can cancel
	if challenge.CreatedByUserID != userID {
		isAuthorized := false

		// Check if user is a team manager for the sender team
		if challenge.SenderTeamID != nil {
			isManager, err := mc.isTeamManager(*challenge.SenderTeamID, userID)
			if err != nil {
				responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to check team management: "+err.Error())
				return
			}
			if isManager {
				isAuthorized = true
			}
		}

		if !isAuthorized {
			responses.ErrorResponse(c, http.StatusForbidden, "You are not authorized to cancel this challenge")
			return
		}
	}

	// Update challenge status
	challenge.Status = StatusCancelled
	if err := mc.repo.UpdateChallenge(challenge); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to cancel challenge: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Challenge cancelled successfully",
	})
}

// CreateDirectMatch handles creating a match directly without a challenge
func (mc *MatchController) CreateDirectMatch(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateDirectMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationErrorResponse(c, err)
		return
	}

	// Validate teams
	// Check if user is a manager for both teams
	isTeam1Manager, err := mc.isTeamManager(req.Team1ID, userID)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to validate team 1: "+err.Error())
		return
	}
	if !isTeam1Manager {
		responses.ErrorResponse(c, http.StatusForbidden, "You must be a manager of team 1 to create a match")
		return
	}

	isTeam2Manager, err := mc.isTeamManager(req.Team2ID, userID)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to validate team 2: "+err.Error())
		return
	}
	if !isTeam2Manager {
		responses.ErrorResponse(c, http.StatusForbidden, "You must be a manager of team 2 to create a match")
		return
	}

	// Create match
	match := Match{
		CreatedByUserID: userID,
		SportID:         req.SportID,
		VenueID:         req.VenueID,
		LocationText:    req.LocationText,
		ScheduledAt:     req.ScheduledAt,
		Duration:        req.Duration,
		Description:     req.Description,
		CustomRules:     req.CustomRules,
		EntryFee:        req.EntryFee,
		WinningPrize:    req.WinningPrize,
		SkillLevel:      req.SkillLevel,
		Status:          StatusMatchUpcoming,
		Visibility:      req.Visibility,
	}

	// Begin transaction to create match and add teams
	err = mc.repo.WithTransaction(func(txRepo MatchRepository) error {
		// Create match
		if err := txRepo.CreateMatch(&match); err != nil {
			return err
		}

		// Add team 1
		team1 := MatchTeam{
			MatchID: match.ID,
			TeamID:  req.Team1ID,
		}
		if err := txRepo.AddTeamToMatch(&team1); err != nil {
			return err
		}

		// Add team 2
		team2 := MatchTeam{
			MatchID: match.ID,
			TeamID:  req.Team2ID,
		}
		if err := txRepo.AddTeamToMatch(&team2); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to create match: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusCreated, gin.H{
		"message": "Match created successfully",
		"match":   match,
	})
}

// GetMatchByID retrieves a specific match by ID
func (mc *MatchController) GetMatchByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid match ID")
		return
	}

	match, err := mc.repo.GetMatchByID(uint(id))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch match: "+err.Error())
		return
	}

	if match == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Match not found")
		return
	}

	responses.SuccessResponse(c, http.StatusOK, match)
}

// UpdateMatch updates an existing match
func (mc *MatchController) UpdateMatch(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid match ID")
		return
	}

	// Get existing match
	match, err := mc.repo.GetMatchByID(uint(id))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch match: "+err.Error())
		return
	}

	if match == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Match not found")
		return
	}

	// Check authorization - only creator or team manager can update
	if match.CreatedByUserID != userID {
		isAuthorized := false

		// Check if user is a manager for any of the participating teams
		for _, matchTeam := range match.MatchTeams {
			isManager, err := mc.isTeamManager(matchTeam.TeamID, userID)
			if err != nil {
				responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to check team management: "+err.Error())
				return
			}
			if isManager {
				isAuthorized = true
				break
			}
		}

		if !isAuthorized {
			responses.ErrorResponse(c, http.StatusForbidden, "You are not authorized to update this match")
			return
		}
	}

	var req UpdateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationErrorResponse(c, err)
		return
	}

	// Update match fields
	if req.Description != nil {
		match.Description = *req.Description
	}
	if req.ScheduledAt != nil {
		match.ScheduledAt = *req.ScheduledAt
	}
	if req.Duration != nil {
		match.Duration = *req.Duration
	}
	if req.VenueID != nil {
		match.VenueID = req.VenueID
	}
	if req.LocationText != nil {
		match.LocationText = *req.LocationText
	}
	if req.CustomRules != nil {
		match.CustomRules = *req.CustomRules
	}
	if req.SkillLevel != nil {
		match.SkillLevel = *req.SkillLevel
	}
	if req.Visibility != nil {
		match.Visibility = *req.Visibility
	}
	if req.StreamURL != nil {
		match.StreamURL = *req.StreamURL
	}
	if req.VodURL != nil {
		match.VodURL = *req.VodURL
	}

	if err := mc.repo.UpdateMatch(match); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to update match: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Match updated successfully",
		"match":   match,
	})
}

// DeleteMatch handles deleting a match
func (mc *MatchController) DeleteMatch(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid match ID")
		return
	}

	// Get existing match
	match, err := mc.repo.GetMatchByID(uint(id))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch match: "+err.Error())
		return
	}

	if match == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Match not found")
		return
	}

	// Check authorization - only creator or admin can delete
	if match.CreatedByUserID != userID {
		responses.ErrorResponse(c, http.StatusForbidden, "You are not authorized to delete this match")
		return
	}

	// Check if match can be deleted (not started or completed)
	if match.Status == StatusMatchLive || match.Status == StatusMatchCompleted {
		responses.ErrorResponse(c, http.StatusBadRequest, "Cannot delete a match that is live or completed")
		return
	}

	if err := mc.repo.DeleteMatch(uint(id)); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete match: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Match deleted successfully",
	})
}

// GetMatches retrieves matches based on filters
func (mc *MatchController) GetMatches(c *gin.Context) {
	// Parse query parameters for filters
	sportID := c.Query("sport_id")
	status := c.Query("status")
	visibility := c.Query("visibility")

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Build filters
	filters := make(map[string]interface{})
	if sportID != "" {
		sportIDInt, err := strconv.Atoi(sportID)
		if err == nil {
			filters["sport_id"] = sportIDInt
		}
	}
	if status != "" {
		filters["status"] = status
	}
	if visibility != "" {
		filters["visibility"] = visibility
	}

	// Get matches
	matches, total, err := mc.repo.GetMatches(filters, page, pageSize)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch matches: "+err.Error())
		return
	}

	responses.PaginatedResponse(c, http.StatusOK, matches, page, pageSize, total)
}

// GetUserMatches retrieves all matches related to the current user
func (mc *MatchController) GetUserMatches(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	matches, total, err := mc.repo.GetUserMatches(userID, status, page, pageSize)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch matches: "+err.Error())
		return
	}

	responses.PaginatedResponse(c, http.StatusOK, matches, page, pageSize, total)
}

// GetTeamMatches retrieves all matches related to a specific team
func (mc *MatchController) GetTeamMatches(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	teamIDStr := c.Param("teamId")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	// Check if user is a member of the team
	isMember, err := mc.isTeamMember(uint(teamID), userID)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to check team membership: "+err.Error())
		return
	}
	if !isMember {
		responses.ErrorResponse(c, http.StatusForbidden, "You must be a member of the team to view its matches")
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	matches, total, err := mc.repo.GetTeamMatches(uint(teamID), status, page, pageSize)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch matches: "+err.Error())
		return
	}

	responses.PaginatedResponse(c, http.StatusOK, matches, page, pageSize, total)
}

// StartMatch handles starting a match
func (mc *MatchController) StartMatch(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid match ID")
		return
	}

	// Get match
	match, err := mc.repo.GetMatchByID(uint(id))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch match: "+err.Error())
		return
	}

	if match == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Match not found")
		return
	}

	// Check authorization - only creator or team manager can start match
	if match.CreatedByUserID != userID {
		isAuthorized := false

		// Check if user is a manager for any of the participating teams
		for _, matchTeam := range match.MatchTeams {
			isManager, err := mc.isTeamManager(matchTeam.TeamID, userID)
			if err != nil {
				responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to check team management: "+err.Error())
				return
			}
			if isManager {
				isAuthorized = true
				break
			}
		}

		if !isAuthorized {
			responses.ErrorResponse(c, http.StatusForbidden, "You are not authorized to start this match")
			return
		}
	}

	// Check if match can be started
	if match.Status != StatusMatchUpcoming {
		responses.ErrorResponse(c, http.StatusBadRequest, "Match cannot be started in its current state")
		return
	}

	// Update match status
	if err := mc.repo.UpdateMatchStatus(match.ID, StatusMatchLive); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to start match: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Match started successfully",
	})
}

// EndMatch handles ending a match and setting the winner
func (mc *MatchController) EndMatch(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid match ID")
		return
	}

	// Get match
	match, err := mc.repo.GetMatchByID(uint(id))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch match: "+err.Error())
		return
	}

	if match == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Match not found")
		return
	}

	// Check authorization - only creator or team manager can end match
	if match.CreatedByUserID != userID {
		isAuthorized := false

		// Check if user is a manager for any of the participating teams
		for _, matchTeam := range match.MatchTeams {
			isManager, err := mc.isTeamManager(matchTeam.TeamID, userID)
			if err != nil {
				responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to check team management: "+err.Error())
				return
			}
			if isManager {
				isAuthorized = true
				break
			}
		}

		if !isAuthorized {
			responses.ErrorResponse(c, http.StatusForbidden, "You are not authorized to end this match")
			return
		}
	}

	// Check if match can be ended
	if match.Status != StatusMatchLive {
		responses.ErrorResponse(c, http.StatusBadRequest, "Match cannot be ended in its current state")
		return
	}

	// Parse winning team ID from request
	var req struct {
		WinningTeamID uint `json:"winning_team_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationErrorResponse(c, err)
		return
	}

	// Validate winning team is part of the match
	isValidTeam := false
	for _, matchTeam := range match.MatchTeams {
		if matchTeam.TeamID == req.WinningTeamID {
			isValidTeam = true
			break
		}
	}
	if !isValidTeam {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid winning team - team must be part of the match")
		return
	}

	// End match
	if err := mc.repo.EndMatch(match.ID, req.WinningTeamID); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to end match: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Match ended successfully",
	})
}

// CancelMatch handles canceling a match
func (mc *MatchController) CancelMatch(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid match ID")
		return
	}

	// Get match
	match, err := mc.repo.GetMatchByID(uint(id))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch match: "+err.Error())
		return
	}

	if match == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Match not found")
		return
	}

	// Check authorization - only creator or team manager can cancel match
	if match.CreatedByUserID != userID {
		isAuthorized := false

		// Check if user is a manager for any of the participating teams
		for _, matchTeam := range match.MatchTeams {
			isManager, err := mc.isTeamManager(matchTeam.TeamID, userID)
			if err != nil {
				responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to check team management: "+err.Error())
				return
			}
			if isManager {
				isAuthorized = true
				break
			}
		}

		if !isAuthorized {
			responses.ErrorResponse(c, http.StatusForbidden, "You are not authorized to cancel this match")
			return
		}
	}

	// Check if match can be canceled
	if match.Status == StatusMatchCompleted || match.Status == StatusMatchCancelled {
		responses.ErrorResponse(c, http.StatusBadRequest, "Match cannot be canceled in its current state")
		return
	}

	// Update match status
	if err := mc.repo.UpdateMatchStatus(match.ID, StatusMatchCancelled); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to cancel match: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Match cancelled successfully",
	})
}

// PostponeMatch handles postponing a match
func (mc *MatchController) PostponeMatch(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid match ID")
		return
	}

	// Get match
	match, err := mc.repo.GetMatchByID(uint(id))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch match: "+err.Error())
		return
	}

	if match == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Match not found")
		return
	}

	// Check authorization - only creator or team manager can postpone match
	if match.CreatedByUserID != userID {
		isAuthorized := false

		// Check if user is a manager for any of the participating teams
		for _, matchTeam := range match.MatchTeams {
			isManager, err := mc.isTeamManager(matchTeam.TeamID, userID)
			if err != nil {
				responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to check team management: "+err.Error())
				return
			}
			if isManager {
				isAuthorized = true
				break
			}
		}

		if !isAuthorized {
			responses.ErrorResponse(c, http.StatusForbidden, "You are not authorized to postpone this match")
			return
		}
	}

	// Check if match can be postponed
	if match.Status != StatusMatchUpcoming {
		responses.ErrorResponse(c, http.StatusBadRequest, "Match cannot be postponed in its current state")
		return
	}

	// Parse new scheduled time from request
	var req struct {
		NewScheduledAt time.Time `json:"new_scheduled_at" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationErrorResponse(c, err)
		return
	}

	// Update match scheduled time
	match.ScheduledAt = req.NewScheduledAt
	match.Status = StatusMatchPostponed

	if err := mc.repo.UpdateMatch(match); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to postpone match: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Match postponed successfully",
		"match":   match,
	})
}

// UpdateMatchScore updates the score for a team in a match
func (mc *MatchController) UpdateMatchScore(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	matchID, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid match ID")
		return
	}

	// Get match
	match, err := mc.repo.GetMatchByID(uint(matchID))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch match: "+err.Error())
		return
	}

	if match == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Match not found")
		return
	}

	// Check authorization - only creator or team manager can update score
	if match.CreatedByUserID != userID {
		isAuthorized := false

		// Check if user is a manager for any of the participating teams
		for _, matchTeam := range match.MatchTeams {
			isManager, err := mc.isTeamManager(matchTeam.TeamID, userID)
			if err != nil {
				responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to check team management: "+err.Error())
				return
			}
			if isManager {
				isAuthorized = true
				break
			}
		}

		if !isAuthorized {
			responses.ErrorResponse(c, http.StatusForbidden, "You are not authorized to update scores for this match")
			return
		}
	}

	// Check if match is in progress
	if match.Status != StatusMatchLive {
		responses.ErrorResponse(c, http.StatusBadRequest, "Scores can only be updated for live matches")
		return
	}

	var req UpdateMatchScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationErrorResponse(c, err)
		return
	}

	// Validate team is part of the match
	isValidTeam := false
	for _, matchTeam := range match.MatchTeams {
		if matchTeam.TeamID == req.TeamID {
			isValidTeam = true
			break
		}
	}
	if !isValidTeam {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid team - team must be part of the match")
		return
	}

	// Update match team score
	matchTeam := MatchTeam{
		MatchID: uint(matchID),
		TeamID:  req.TeamID,

		ResultStatus: req.ResultStatus,
	}

	if err := mc.repo.UpdateMatchScore(&matchTeam); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to update match score: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Match score updated successfully",
	})
}

// --- Tournament Controller Methods ---

// CreateTournament handles creating a new tournament
func (mc *MatchController) CreateTournament(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateTournamentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationErrorResponse(c, err)
		return
	}

	// Validate dates
	if req.StartDate.Before(time.Now()) {
		responses.ErrorResponse(c, http.StatusBadRequest, "Start date must be in the future")
		return
	}
	if req.EndDate.Before(req.StartDate) {
		responses.ErrorResponse(c, http.StatusBadRequest, "End date must be after start date")
		return
	}
	if req.RegistrationDeadline.After(req.StartDate) {
		responses.ErrorResponse(c, http.StatusBadRequest, "Registration deadline must be before start date")
		return
	}

	// Create tournament
	tournament := Tournament{
		Name:                 req.Name,
		Description:          req.Description,
		CreatedByUserID:      userID,
		SportID:              req.SportID,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		RegistrationDeadline: req.RegistrationDeadline,
		Format:               req.Format,
		FormatDetails:        req.FormatDetails,
		PrizeDescription:     req.PrizeDescription,
		PrizePool:            req.PrizePool,
		EntryFee:             req.EntryFee,
		MaxTeams:             req.MaxTeams,
		Status:               "registration_open",
	}

	if err := mc.repo.CreateTournament(&tournament); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to create tournament: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusCreated, gin.H{
		"message":    "Tournament created successfully",
		"tournament": tournament,
	})
}

// GetTournaments retrieves tournaments based on filters
func (mc *MatchController) GetTournaments(c *gin.Context) {
	// Parse query parameters for filters
	sportID := c.Query("sport_id")
	status := c.Query("status")
	format := c.Query("format")

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Build filters
	filters := make(map[string]interface{})
	if sportID != "" {
		sportIDInt, err := strconv.Atoi(sportID)
		if err == nil {
			filters["sport_id"] = sportIDInt
		}
	}
	if status != "" {
		filters["status"] = status
	}
	if format != "" {
		filters["format"] = format
	}

	// Get tournaments
	tournaments, total, err := mc.repo.GetTournaments(filters, page, pageSize)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tournaments: "+err.Error())
		return
	}

	responses.PaginatedResponse(c, http.StatusOK, tournaments, page, pageSize, total)
}

// GetTournamentByID retrieves a specific tournament by ID
func (mc *MatchController) GetTournamentByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	tournament, err := mc.repo.GetTournamentByID(uint(id))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tournament: "+err.Error())
		return
	}

	if tournament == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Tournament not found")
		return
	}

	responses.SuccessResponse(c, http.StatusOK, tournament)
}
func (mc *MatchController) UpdateTournament(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	tournament, err := mc.repo.GetTournamentByID(uint(id))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tournament: "+err.Error())
		return
	}

	if tournament == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Tournament not found")
		return
	}

	if tournament.CreatedByUserID != userID {
		responses.ErrorResponse(c, http.StatusForbidden, "You are not authorized to update this tournament")
		return
	}

	var req UpdateTournamentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationErrorResponse(c, err)
		return
	}

	if req.Name != nil {
		tournament.Name = *req.Name
	}
	if req.Description != nil {
		tournament.Description = *req.Description
	}
	if req.StartDate != nil {
		tournament.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		tournament.EndDate = *req.EndDate
	}
	if req.RegistrationDeadline != nil {
		tournament.RegistrationDeadline = *req.RegistrationDeadline
	}
	if req.Format != nil {
		tournament.Format = *req.Format
	}
	if req.FormatDetails != nil {
		tournament.FormatDetails = *req.FormatDetails
	}
	if req.PrizeDescription != nil {
		tournament.PrizeDescription = *req.PrizeDescription
	}
	if req.PrizePool != nil {
		tournament.PrizePool = *req.PrizePool
	}
	if req.EntryFee != nil {
		tournament.EntryFee = *req.EntryFee
	}
	if req.MaxTeams != nil {
		tournament.MaxTeams = *req.MaxTeams
	}
	if req.Status != nil {
		tournament.Status = *req.Status
	}

	if err := mc.repo.UpdateTournament(tournament); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to update tournament: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{
		"message":    "Tournament updated successfully",
		"tournament": tournament,
	})
}

func (mc *MatchController) AdminOverrideMatchStatus(c *gin.Context) {
	idStr := c.Param("id")
	matchID, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid match ID")
		return
	}

	var req struct {
		Status MatchStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationErrorResponse(c, err)
		return
	}

	match, err := mc.repo.GetMatchByID(uint(matchID))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch match: "+err.Error())
		return
	}
	if match == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Match not found")
		return
	}

	if err := mc.repo.UpdateMatchStatus(uint(matchID), req.Status); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to override match status: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{"message": "Match status overridden successfully"})
}

func (mc *MatchController) AdminOverrideMatchScore(c *gin.Context) {
	matchIDStr := c.Param("id")
	matchID, err := strconv.Atoi(matchIDStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid match ID")
		return
	}

	var req []UpdateMatchScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationErrorResponse(c, err)
		return
	}

	match, err := mc.repo.GetMatchByID(uint(matchID))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch match: "+err.Error())
		return
	}
	if match == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Match not found")
		return
	}

	if match.Status != StatusMatchCompleted && match.Status != StatusMatchLive {
		responses.ErrorResponse(c, http.StatusBadRequest, "Scores can only be overridden for live or completed matches.")
		return
	}

	err = mc.repo.WithTransaction(func(txRepo MatchRepository) error {
		for _, scoreUpdate := range req {
			matchTeam := MatchTeam{
				MatchID: uint(matchID),
				TeamID:  scoreUpdate.TeamID,

				ResultStatus: scoreUpdate.ResultStatus,
			}
			if err := txRepo.UpdateMatchScore(&matchTeam); err != nil {
				return errors.New("failed to update score for team " + strconv.Itoa(int(scoreUpdate.TeamID)) + ": " + err.Error())
			}
		}
		return nil
	})

	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to override match scores: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{"message": "Match scores overridden successfully"})
}
func (mc *MatchController) ExpireChallenges(c *gin.Context) {
	err := mc.repo.ExpireChallenges()
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to expire challenges: "+err.Error())
		return
	}
	responses.SuccessResponse(c, http.StatusOK, gin.H{"message": "Challenges expired successfully"})
}

func (mc *MatchController) DeleteTournament(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tournamentIDStr := c.Param("id")
	tournamentID, err := strconv.Atoi(tournamentIDStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	tournament, err := mc.repo.GetTournamentByID(uint(tournamentID))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tournament: "+err.Error())
		return
	}

	if tournament == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Tournament not found")
		return
	}

	if tournament.CreatedByUserID != userID {
		responses.ErrorResponse(c, http.StatusForbidden, "You are not authorized to delete this tournament")
		return
	}

	if tournament.Status == "ongoing" || tournament.Status == "completed" {
		responses.ErrorResponse(c, http.StatusBadRequest, "Cannot delete an ongoing or completed tournament")
		return
	}

	if err := mc.repo.DeleteTournament(uint(tournamentID)); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete tournament: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{"message": "Tournament deleted successfully"})
}

type TournamentTeamRegistrationRequest struct {
	TeamID uint `json:"team_id" binding:"required"`
}

func (mc *MatchController) RegisterTeamForTournament(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tournamentIDStr := c.Param("id")
	tournamentID, err := strconv.Atoi(tournamentIDStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	var req TournamentTeamRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationErrorResponse(c, err)
		return
	}

	tournament, err := mc.repo.GetTournamentByID(uint(tournamentID))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tournament: "+err.Error())
		return
	}
	if tournament == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Tournament not found")
		return
	}

	if tournament.Status != "registration_open" {
		responses.ErrorResponse(c, http.StatusBadRequest, "Tournament registration is not open")
		return
	}

	if time.Now().After(tournament.RegistrationDeadline) {
		responses.ErrorResponse(c, http.StatusBadRequest, "Registration deadline has passed")
		return
	}

	if tournament.MaxTeams > 0 && tournament.CurrentTeams >= tournament.MaxTeams {
		responses.ErrorResponse(c, http.StatusBadRequest, "Tournament is full")
		return
	}

	isManager, err := mc.isTeamManager(req.TeamID, userID)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to verify team manager status: "+err.Error())
		return
	}
	if !isManager {
		responses.ErrorResponse(c, http.StatusForbidden, "You must be a manager of the team to register it")
		return
	}

	if err := mc.repo.RegisterTeamInTournament(uint(tournamentID), req.TeamID); err != nil {
		if err.Error() == "team already registered" { // Example specific error check
			responses.ErrorResponse(c, http.StatusConflict, "Team is already registered for this tournament")
			return
		}
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to register team: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{"message": "Team registered successfully for the tournament"})
}

func (mc *MatchController) UnregisterTeamFromTournament(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tournamentIDStr := c.Param("id")
	tournamentID, err := strconv.Atoi(tournamentIDStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	var req TournamentTeamRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationErrorResponse(c, err)
		return
	}

	tournament, err := mc.repo.GetTournamentByID(uint(tournamentID))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tournament: "+err.Error())
		return
	}
	if tournament == nil {
		responses.ErrorResponse(c, http.StatusNotFound, "Tournament not found")
		return
	}

	if tournament.Status != "registration_open" {
		responses.ErrorResponse(c, http.StatusBadRequest, "Cannot unregister team if registration is not open")
		return
	}
	if time.Now().After(tournament.RegistrationDeadline) && tournament.Status == "registration_open" {
		// Allow unregistration after deadline if still in registration_open status by admin,
		// but typically this might be disallowed or have penalties.
		// For now, let's allow it if status is still registration_open.
	} else if tournament.Status != "registration_open" { // Stricter check for other statuses
		responses.ErrorResponse(c, http.StatusBadRequest, "Cannot unregister team from a tournament that is not in registration phase.")
		return
	}

	isManager, err := mc.isTeamManager(req.TeamID, userID)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to verify team manager status: "+err.Error())
		return
	}
	if !isManager {
		responses.ErrorResponse(c, http.StatusForbidden, "You must be a manager of the team to unregister it")
		return
	}

	if err := mc.repo.UnregisterTeamFromTournament(uint(tournamentID), req.TeamID); err != nil {
		if err.Error() == "team not registered" { // Example specific error check
			responses.ErrorResponse(c, http.StatusNotFound, "Team is not registered for this tournament")
			return
		}
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to unregister team: "+err.Error())
		return
	}

	responses.SuccessResponse(c, http.StatusOK, gin.H{"message": "Team unregistered successfully from the tournament"})
}

func (mc *MatchController) GetTournamentMatches(c *gin.Context) {
	tournamentIDStr := c.Param("id")
	tournamentID, err := strconv.Atoi(tournamentIDStr)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	_, err = mc.repo.GetTournamentByID(uint(tournamentID))
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tournament details: "+err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	filters := make(map[string]interface{})
	filters["tournament_id"] = uint(tournamentID)

	status := c.Query("status")
	if status != "" {
		filters["status"] = status
	}

	matches, total, err := mc.repo.GetMatches(filters, page, pageSize)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tournament matches: "+err.Error())
		return
	}

	responses.PaginatedResponse(c, http.StatusOK, matches, page, pageSize, total)
}
