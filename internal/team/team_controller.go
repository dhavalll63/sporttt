package team

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DhavalSuthar-24/miow/config" // Assuming your config package

	// "github.com/DhavalSuthar-24/miow/internal/user" // Assuming user package for User model if needed for responses
	// Generic response package
	"github.com/DhavalSuthar-24/miow/pkg/responses"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

const (
	RolePlayer       = "player"
	RoleModerator    = "moderator"
	RoleViceCaptain  = "vice_captain"
	RoleCaptain      = "captain"
	StatusPending    = "pending"
	StatusApproved   = "approved"
	StatusAccepted   = "accepted"
	StatusRejected   = "rejected"
	StatusCancelled  = "cancelled"
	DefaultAvatarURL = "path/to/default/team_logo.png" // Placeholder
)

// TeamController handles team-related HTTP requests
type TeamController struct {
	repo      TeamRepository
	appConfig *config.Config
	// userRepo user.UserRepository
}

// NewTeamController creates a new team controller
func NewTeamController(repo TeamRepository, appConfig *config.Config /*, userRepo user.UserRepository*/) *TeamController {
	return &TeamController{
		repo:      repo,
		appConfig: appConfig,
		// userRepo: userRepo,
	}
}

// --- Helper Functions for Auth ---
// These would typically be part of a middleware or a shared auth utility package

func getCurrentUserID(c *gin.Context) (uint, bool) {
	userIDVal, exists := c.Get("currentUserID")
	if !exists {
		return 0, false
	}
	userID, ok := userIDVal.(uint)
	return userID, ok
}

// isTeamManager checks if the user is creator, captain, vice_captain or moderator of the team
func (tc *TeamController) isTeamManager(teamID, userID uint) (bool, error) {
	isCreator, err := tc.repo.IsUserTeamCreator(teamID, userID)
	if err != nil {
		return false, err
	}
	if isCreator {
		return true, nil
	}

	member, err := tc.repo.GetTeamMember(teamID, userID)
	if err != nil {
		return false, err
	}
	if member == nil || !member.IsActive {
		return false, nil
	}
	return member.Role == RoleCaptain || member.Role == RoleViceCaptain || member.Role == RoleModerator || member.IsCaptain, nil
}

// isTeamCreator checks if the user is the creator of the team
func (tc *TeamController) isTeamCreator(teamID, userID uint) (bool, error) {
	return tc.repo.IsUserTeamCreator(teamID, userID)
}

// isAdminUser checks if the current user has admin privileges
func isAdminUser(c *gin.Context) bool {
	rolesVal, exists := c.Get("currentUserRoles")
	if !exists {
		return false
	}
	roles, ok := rolesVal.([]string)
	if !ok {
		return false
	}
	for _, role := range roles {
		if role == "admin" { // Assuming "admin" is the role name for administrators
			return true
		}
	}
	return false
}

// --- DTOs for requests ---

type CreateTeamRequest struct {
	Name         string `json:"name" binding:"required,min=3,max=100"`
	Description  string `json:"description" binding:"max=1000"`
	Logo         string `json:"logo"`
	SportID      uint   `json:"sport_id" binding:"required"`
	MinPlayers   int    `json:"min_players" binding:"gte=1"`
	MaxPlayers   int    `json:"max_players" binding:"gtefield=MinPlayers"`
	Requirements string `json:"requirements"` // JSON string
	Level        string `json:"level"`
	SocialLinks  string `json:"social_links"` // JSON string
}

type UpdateTeamRequest struct {
	Name         *string `json:"name" binding:"omitempty,min=3,max=100"`
	Description  *string `json:"description" binding:"omitempty,max=1000"`
	Logo         *string `json:"logo"`
	MinPlayers   *int    `json:"min_players" binding:"omitempty,gte=1"`
	MaxPlayers   *int    `json:"max_players" binding:"omitempty,gtefield=MinPlayers"` // This validation might need custom logic if MinPlayers is not also updated
	Requirements *string `json:"requirements"`                                        // JSON string
	Level        *string `json:"level"`
	SocialLinks  *string `json:"social_links"` // JSON string
}

type InviteUserRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	Role     string `json:"role" binding:"omitempty,oneof=player moderator vice_captain captain"`
	Position string `json:"position"`
	Message  string `json:"message" binding:"max=500"`
}

type CreateJoinRequest struct {
	Message  string `json:"message" binding:"max=500"`
	Position string `json:"position"`
	Skills   string `json:"skills"` // JSON string
}

type UpdateMemberRoleRequest struct {
	Role      string `json:"role" binding:"required,oneof=player moderator vice_captain captain"`
	IsCaptain *bool  `json:"is_captain"` // Explicitly set captain status
}

// --- Team Handlers ---

// CreateTeam godoc
// @Summary Create a new team
// @Description Creates a new team with the authenticated user as the creator and captain.
// @Tags Teams
// @Accept json
// @Produce json
// @Param team body CreateTeamRequest true "Team Creation Data"
// @Success 201 {object} responses.SuccessResponse{data=Team} "Team created successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid input"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /teams [post]
func (tc *TeamController) CreateTeam(c *gin.Context) {
	userID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Check if team name already exists
	existingTeam, _ := tc.repo.GetTeamByName(req.Name)
	if existingTeam != nil {
		responses.SendError(c, http.StatusConflict, "Team name already exists")
		return
	}

	team := Team{
		Name:         req.Name,
		Description:  req.Description,
		Logo:         req.Logo,
		CreatedByID:  userID,
		SportID:      req.SportID,
		MinPlayers:   req.MinPlayers,
		MaxPlayers:   req.MaxPlayers,
		Requirements: req.Requirements,
		Level:        req.Level,
		SocialLinks:  req.SocialLinks,
		Rating:       1000.0, // Default rating
	}
	if team.Logo == "" {
		team.Logo = DefaultAvatarURL
	}

	// Use a transaction to create team and initial member
	err := tc.repo.WithTransaction(func(repo TeamRepository) error {
		// Create team within transaction using the repo parameter
		if err := repo.Create(&team); err != nil {
			return err
		}

		// Add creator as the first member and captain
		creatorMember := TeamMember{
			TeamID:    team.ID,
			UserID:    userID,
			Role:      RoleCaptain,
			IsCaptain: true,
			JoinedAt:  time.Now(),
			IsActive:  true,
		}
		return repo.Create(&creatorMember)
	})

	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to create team: "+err.Error())
		return
	}

	// Reload team to get Sport populated
	createdTeam, _ := tc.repo.GetTeamByID(team.ID)
	responses.SendSuccess(c, http.StatusCreated, "Team created successfully", createdTeam)
}

// GetTeamByID godoc
// @Summary Get a team by its ID
// @Description Retrieves details of a specific team.
// @Tags Teams
// @Produce json
// @Param team_id path uint true "Team ID"
// @Success 200 {object} responses.SuccessResponse{data=Team} "Team details"
// @Failure 404 {object} responses.ErrorResponse "Team not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /teams/{team_id} [get]
func (tc *TeamController) GetTeamByID(c *gin.Context) {
	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	team, err := tc.repo.GetTeamByID(uint(teamID))
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve team: "+err.Error())
		return
	}
	if team == nil || team.IsDeleted {
		responses.SendError(c, http.StatusNotFound, "Team not found")
		return
	}
	responses.SendSuccess(c, http.StatusOK, "Team retrieved successfully", team)
}

// GetAllTeams godoc
// @Summary Get all teams
// @Description Retrieves a list of all teams with optional filters and pagination.
// @Tags Teams
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param sport_id query int false "Filter by Sport ID"
// @Param level query string false "Filter by team level (e.g., 'Amateur', 'Professional')"
// @Param name query string false "Search by team name (case-insensitive, partial match)"
// @Success 200 {object} responses.PaginatedResponse{data=[]Team} "List of teams"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /teams [get]
func (tc *TeamController) GetAllTeams(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	filters := make(map[string]interface{})
	if sportIDStr := c.Query("sport_id"); sportIDStr != "" {
		sportID, err := strconv.ParseUint(sportIDStr, 10, 32)
		if err == nil {
			filters["sport_id"] = uint(sportID)
		}
	}
	if level := c.Query("level"); level != "" {
		filters["level"] = level
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}

	teams, total, err := tc.repo.GetAllTeams(page, limit, filters)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve teams: "+err.Error())
		return
	}
	responses.SendPaginated(c, http.StatusOK, "Teams retrieved successfully", teams, total, page, limit)
}

// UpdateTeam godoc
// @Summary Update a team
// @Description Updates details of an existing team. Only team creator or captain can update.
// @Tags Teams
// @Accept json
// @Produce json
// @Param team_id path uint true "Team ID"
// @Param team body UpdateTeamRequest true "Team Update Data"
// @Success 200 {object} responses.SuccessResponse{data=Team} "Team updated successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid input or team ID"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden - Not team creator or captain"
// @Failure 404 {object} responses.ErrorResponse "Team not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /teams/{team_id} [put]
func (tc *TeamController) UpdateTeam(c *gin.Context) {
	userID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	team, err := tc.repo.GetTeamByID(uint(teamID))
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve team: "+err.Error())
		return
	}
	if team == nil || team.IsDeleted {
		responses.SendError(c, http.StatusNotFound, "Team not found")
		return
	}

	// Authorization: Only creator or a designated captain can update
	isCreator, _ := tc.isTeamCreator(uint(teamID), userID)
	memberRole, _ := tc.repo.GetUserTeamRole(uint(teamID), userID)

	if !isCreator && memberRole != RoleCaptain {
		responses.SendError(c, http.StatusForbidden, "Only the team creator or captain can update the team")
		return
	}

	var req UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	if req.Name != nil {
		team.Name = *req.Name
	}
	if req.Description != nil {
		team.Description = *req.Description
	}
	if req.Logo != nil {
		team.Logo = *req.Logo
	}
	if req.MinPlayers != nil {
		team.MinPlayers = *req.MinPlayers
	}
	if req.MaxPlayers != nil {
		team.MaxPlayers = *req.MaxPlayers
	}
	if req.Requirements != nil {
		team.Requirements = *req.Requirements
	}
	if req.Level != nil {
		team.Level = *req.Level
	}
	if req.SocialLinks != nil {
		team.SocialLinks = *req.SocialLinks
	}

	if req.MaxPlayers != nil && req.MinPlayers == nil && *req.MaxPlayers < team.MinPlayers {
		responses.SendError(c, http.StatusBadRequest, "Max players cannot be less than current min players without updating min players")
		return
	}
	if req.MinPlayers != nil && req.MaxPlayers == nil && *req.MinPlayers > team.MaxPlayers {
		responses.SendError(c, http.StatusBadRequest, "Min players cannot be greater than current max players without updating max players")
		return
	}
	if req.MinPlayers != nil && req.MaxPlayers != nil && *req.MinPlayers > *req.MaxPlayers {
		responses.SendError(c, http.StatusBadRequest, "Min players cannot be greater than max players")
		return
	}

	if err := tc.repo.UpdateTeam(team); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to update team: "+err.Error())
		return
	}
	// Reload team to get Sport populated if necessary
	updatedTeam, _ := tc.repo.GetTeamByID(team.ID)
	responses.SendSuccess(c, http.StatusOK, "Team updated successfully", updatedTeam)
}

// DeleteTeam godoc
// @Summary Delete a team
// @Description Deletes a team. Only team creator or an admin can delete. Soft delete by default.
// @Tags Teams
// @Produce json
// @Param team_id path uint true "Team ID"
// @Param hard_delete query bool false "Hard delete team and associated data" default(false)
// @Success 200 {object} responses.SuccessResponse "Team deleted successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid team ID"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden - Not team creator or admin"
// @Failure 404 {object} responses.ErrorResponse "Team not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /teams/{team_id} [delete]
func (tc *TeamController) DeleteTeam(c *gin.Context) {
	userID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	hardDelete, _ := strconv.ParseBool(c.DefaultQuery("hard_delete", "false"))

	team, err := tc.repo.GetTeamByID(uint(teamID))
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve team: "+err.Error())
		return
	}
	if team == nil { // Check for nil explicitly, IsDeleted handled by GetTeamByID if strict
		responses.SendError(c, http.StatusNotFound, "Team not found")
		return
	}

	isCreator, _ := tc.isTeamCreator(uint(teamID), userID)
	if !isCreator && !isAdminUser(c) {
		responses.SendError(c, http.StatusForbidden, "Only the team creator or an admin can delete the team")
		return
	}

	if err := tc.repo.DeleteTeam(uint(teamID), hardDelete); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to delete team: "+err.Error())
		return
	}
	responses.SendSuccess(c, http.StatusOK, "Team deleted successfully", nil)
}

// GetMyTeams godoc
// @Summary Get teams for the current user
// @Description Retrieves a list of teams the authenticated user is a member of.
// @Tags Teams
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} responses.PaginatedResponse{data=[]Team} "List of user's teams"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /users/me/teams [get]
func (tc *TeamController) GetMyTeams(c *gin.Context) {
	userID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	teams, total, err := tc.repo.GetTeamsByUserID(userID, page, limit)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve your teams: "+err.Error())
		return
	}
	responses.SendPaginated(c, http.StatusOK, "Your teams retrieved successfully", teams, total, page, limit)
}

// GetTeamsCreatedByMe godoc
// @Summary Get teams created by the current user
// @Description Retrieves a list of teams created by the authenticated user.
// @Tags Teams
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} responses.PaginatedResponse{data=[]Team} "List of teams created by user"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /users/me/teams/created [get]
func (tc *TeamController) GetTeamsCreatedByMe(c *gin.Context) {
	userID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	teams, total, err := tc.repo.GetTeamsCreatedByUserID(userID, page, limit)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve created teams: "+err.Error())
		return
	}
	responses.SendPaginated(c, http.StatusOK, "Teams created by you retrieved successfully", teams, total, page, limit)
}

// --- Team Member Handlers ---

// GetTeamMembers godoc
// @Summary Get team members
// @Description Retrieves a list of members for a specific team.
// @Tags Team Members
// @Produce json
// @Param team_id path uint true "Team ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param role query string false "Filter by member role (e.g., 'player', 'captain')"
// @Success 200 {object} responses.PaginatedResponse{data=[]TeamMember} "List of team members"
// @Failure 400 {object} responses.ErrorResponse "Invalid team ID"
// @Failure 404 {object} responses.ErrorResponse "Team not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /teams/{team_id}/members [get]
func (tc *TeamController) GetTeamMembers(c *gin.Context) {
	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	team, err := tc.repo.GetTeamByID(uint(teamID))
	if err != nil || team == nil || team.IsDeleted {
		responses.SendError(c, http.StatusNotFound, "Team not found")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	roleFilter := c.Query("role")
	var members []TeamMember
	var total int64

	if roleFilter != "" {
		members, total, err = tc.repo.GetTeamMembersByRole(uint(teamID), roleFilter, page, limit)
	} else {
		members, total, err = tc.repo.GetTeamMembers(uint(teamID), page, limit)
	}

	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve team members: "+err.Error())
		return
	}
	responses.SendPaginated(c, http.StatusOK, "Team members retrieved successfully", members, total, page, limit)
}

// RemoveTeamMember godoc
// @Summary Remove a team member
// @Description Removes a member from a team. Only team creator or captain can remove members. Creator cannot be removed this way.
// @Tags Team Members
// @Produce json
// @Param team_id path uint true "Team ID"
// @Param user_id path uint true "User ID of the member to remove"
// @Success 200 {object} responses.SuccessResponse "Member removed successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid ID(s)"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden - Insufficient permissions"
// @Failure 404 {object} responses.ErrorResponse "Team or member not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /teams/{team_id}/members/{user_id} [delete]
func (tc *TeamController) RemoveTeamMember(c *gin.Context) {
	currentUserID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid team ID")
		return
	}
	memberUserID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	team, err := tc.repo.GetTeamByID(uint(teamID))
	if err != nil || team == nil || team.IsDeleted {
		responses.SendError(c, http.StatusNotFound, "Team not found")
		return
	}

	// Authorization
	isCreator, _ := tc.isTeamCreator(uint(teamID), currentUserID)
	currentUserRole, _ := tc.repo.GetUserTeamRole(uint(teamID), currentUserID)

	if !isCreator && currentUserRole != RoleCaptain {
		responses.SendError(c, http.StatusForbidden, "Only team creator or captain can remove members")
		return
	}

	// Prevent creator from being removed by captain through this endpoint
	if team.CreatedByID == uint(memberUserID) && !isCreator {
		responses.SendError(c, http.StatusForbidden, "Captain cannot remove the team creator")
		return
	}
	// Prevent self-removal if you are the creator and sole captain (team would be orphaned)
	if team.CreatedByID == uint(memberUserID) && currentUserID == uint(memberUserID) {
		// Check if other captains exist
		captains, _, _ := tc.repo.GetTeamMembersByRole(uint(teamID), RoleCaptain, 1, 2) // page 1, limit 2
		if len(captains) <= 1 && captains[0].UserID == currentUserID {                  // Only self as captain
			responses.SendError(c, http.StatusForbidden, "Creator cannot leave if they are the sole captain. Promote another captain first or delete the team.")
			return
		}
	}

	memberToRemove, err := tc.repo.GetTeamMember(uint(teamID), uint(memberUserID))
	if err != nil || memberToRemove == nil || !memberToRemove.IsActive {
		responses.SendError(c, http.StatusNotFound, "Member not found in this team or already inactive")
		return
	}

	if err := tc.repo.RemoveTeamMember(uint(teamID), uint(memberUserID)); err != nil { // This now means setting IsActive to false
		responses.SendError(c, http.StatusInternalServerError, "Failed to remove member: "+err.Error())
		return
	}
	responses.SendSuccess(c, http.StatusOK, "Member removed (deactivated) successfully", nil)
}

// UpdateTeamMemberRole godoc
// @Summary Update a team member's role
// @Description Updates the role of a team member. Only team creator or captain can change roles.
// @Description Valid roles: 'player', 'moderator', 'vice_captain', 'captain'.
// @Tags Team Members
// @Accept json
// @Produce json
// @Param team_id path uint true "Team ID"
// @Param user_id path uint true "User ID of the member whose role is to be updated"
// @Param role_update body UpdateMemberRoleRequest true "Role Update Data"
// @Success 200 {object} responses.SuccessResponse{data=TeamMember} "Member role updated successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid input or IDs"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden - Insufficient permissions"
// @Failure 404 {object} responses.ErrorResponse "Team or member not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /teams/{team_id}/members/{user_id}/role [put]
func (tc *TeamController) UpdateTeamMemberRole(c *gin.Context) {
	currentUserID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid team ID")
		return
	}
	memberUserID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	team, err := tc.repo.GetTeamByID(uint(teamID))
	if err != nil || team == nil || team.IsDeleted {
		responses.SendError(c, http.StatusNotFound, "Team not found")
		return
	}

	// Authorization
	isCreator, _ := tc.isTeamCreator(uint(teamID), currentUserID)
	currentUserRole, _ := tc.repo.GetUserTeamRole(uint(teamID), currentUserID)

	if !isCreator && currentUserRole != RoleCaptain {
		responses.SendError(c, http.StatusForbidden, "Only team creator or captain can change member roles")
		return
	}

	memberToUpdate, err := tc.repo.GetTeamMember(uint(teamID), uint(memberUserID))
	if err != nil || memberToUpdate == nil || !memberToUpdate.IsActive {
		responses.SendError(c, http.StatusNotFound, "Member not found in this team or is inactive")
		return
	}

	// Creator's role cannot be changed to non-captain by another captain
	if team.CreatedByID == uint(memberUserID) && req.Role != RoleCaptain && !isCreator {
		responses.SendError(c, http.StatusForbidden, "Cannot change the role of the team creator to non-captain.")
		return
	}
	// Creator can only be 'captain'
	if team.CreatedByID == uint(memberUserID) && req.Role != RoleCaptain {
		responses.SendError(c, http.StatusForbidden, "Team creator must have the 'captain' role.")
		return
	}

	memberToUpdate.Role = req.Role
	if req.IsCaptain != nil { // Allow explicit setting of IsCaptain
		memberToUpdate.IsCaptain = *req.IsCaptain
	} else { // Default IsCaptain based on role
		memberToUpdate.IsCaptain = (req.Role == RoleCaptain)
	}

	if err := tc.repo.UpdateTeamMember(memberToUpdate); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to update member role: "+err.Error())
		return
	}
	responses.SendSuccess(c, http.StatusOK, "Member role updated successfully", memberToUpdate)
}

// LeaveTeam godoc
// @Summary Leave a team
// @Description Allows an authenticated user to leave a team they are a member of.
// @Description Creator cannot leave if they are the sole captain; they must promote another captain or delete the team.
// @Tags Team Members
// @Produce json
// @Param team_id path uint true "Team ID"
// @Success 200 {object} responses.SuccessResponse "Successfully left the team"
// @Failure 400 {object} responses.ErrorResponse "Invalid team ID"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden - Cannot leave (e.g., creator is sole captain)"
// @Failure 404 {object} responses.ErrorResponse "Team or member not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /teams/{team_id}/leave [post]
func (tc *TeamController) LeaveTeam(c *gin.Context) {
	userID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	team, err := tc.repo.GetTeamByID(uint(teamID))
	if err != nil || team == nil || team.IsDeleted {
		responses.SendError(c, http.StatusNotFound, "Team not found")
		return
	}

	member, err := tc.repo.GetTeamMember(uint(teamID), userID)
	if err != nil || member == nil || !member.IsActive {
		responses.SendError(c, http.StatusNotFound, "You are not an active member of this team")
		return
	}

	// Creator specific logic for leaving
	isCreator, _ := tc.isTeamCreator(uint(teamID), userID)
	if isCreator {
		// Check if other captains exist
		captains, _, err := tc.repo.GetTeamMembersByRole(uint(teamID), RoleCaptain, 1, 2) // page 1, limit 2 (includes self if captain)
		if err != nil {
			responses.SendError(c, http.StatusInternalServerError, "Failed to check team captains: "+err.Error())
			return
		}

		soleCaptain := true
		if len(captains) > 1 { // If more than one captain record
			soleCaptain = false
		} else if len(captains) == 1 && captains[0].UserID != userID { // One captain record, but it's not the creator
			soleCaptain = false
		}

		if soleCaptain {
			responses.SendError(c, http.StatusForbidden, "Team creator cannot leave if they are the sole captain. Please promote another member to captain first or delete the team.")
			return
		}
		// If creator leaves and there are other captains, ownership doesn't automatically transfer here.
		// This logic might need to be expanded for explicit ownership transfer. For now, creator just leaves their member status.
		// The team.CreatedByID remains.
	}

	// Set member to inactive
	if err := tc.repo.RemoveTeamMember(uint(teamID), userID); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to leave team: "+err.Error())
		return
	}
	responses.SendSuccess(c, http.StatusOK, "Successfully left the team", nil)
}

// --- Join Request Handlers ---

// RequestToJoinTeam godoc
// @Summary Request to join a team
// @Description Sends a request to join a specific team.
// @Tags Join Requests
// @Accept json
// @Produce json
// @Param team_id path uint true "Team ID"
// @Param join_request body CreateJoinRequest true "Join Request Details"
// @Success 201 {object} responses.SuccessResponse{data=JoinRequest} "Join request sent successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid input or team ID"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden - Already a member, or pending request/invitation exists"
// @Failure 404 {object} responses.ErrorResponse "Team not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /teams/{team_id}/join-requests [post]
func (tc *TeamController) RequestToJoinTeam(c *gin.Context) {
	userID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	var req CreateJoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	team, err := tc.repo.GetTeamByID(uint(teamID))
	if err != nil || team == nil || team.IsDeleted {
		responses.SendError(c, http.StatusNotFound, "Team not found")
		return
	}

	// Check if already a member
	isMember, _ := tc.repo.IsUserTeamMember(uint(teamID), userID)
	if isMember {
		responses.SendError(c, http.StatusForbidden, "You are already a member of this team")
		return
	}

	// Check for existing pending join request
	existingRequest, _ := tc.repo.GetPendingJoinRequest(uint(teamID), userID)
	if existingRequest != nil {
		responses.SendError(c, http.StatusForbidden, "You already have a pending join request for this team")
		return
	}

	// Check for existing pending invitation
	existingInvitation, _ := tc.repo.GetPendingInvitation(uint(teamID), userID)
	if existingInvitation != nil {
		responses.SendError(c, http.StatusForbidden, "You have a pending invitation from this team. Please respond to it.")
		return
	}

	joinRequest := JoinRequest{
		TeamID:    uint(teamID),
		UserID:    userID,
		Message:   req.Message,
		Position:  req.Position,
		Skills:    req.Skills,
		Status:    StatusPending,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Example: 7 days expiry
	}

	if err := tc.repo.CreateJoinRequest(&joinRequest); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to send join request: "+err.Error())
		return
	}
	responses.SendSuccess(c, http.StatusCreated, "Join request sent successfully", joinRequest)
}

// GetJoinRequestsForTeam godoc
// @Summary Get join requests for a team
// @Description Retrieves join requests for a team. Only for team managers (creator, captain, vice-captain, moderator).
// @Tags Join Requests
// @Produce json
// @Param team_id path uint true "Team ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status (e.g., 'pending', 'approved', 'rejected')"
// @Success 200 {object} responses.PaginatedResponse{data=[]JoinRequest} "List of join requests"
// @Failure 400 {object} responses.ErrorResponse "Invalid team ID"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden - Insufficient permissions"
// @Failure 404 {object} responses.ErrorResponse "Team not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /teams/{team_id}/join-requests [get]
func (tc *TeamController) GetJoinRequestsForTeam(c *gin.Context) {
	currentUserID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	team, err := tc.repo.GetTeamByID(uint(teamID))
	if err != nil || team == nil || team.IsDeleted {
		responses.SendError(c, http.StatusNotFound, "Team not found")
		return
	}

	// Authorization: Check if current user is a manager of this team
	isManager, err := tc.isTeamManager(uint(teamID), currentUserID)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Error checking permissions: "+err.Error())
		return
	}
	if !isManager {
		responses.SendError(c, http.StatusForbidden, "Only team managers can view join requests")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	statusFilter := strings.ToLower(c.DefaultQuery("status", StatusPending)) // Default to pending

	requests, total, err := tc.repo.GetJoinRequestsByTeamID(uint(teamID), statusFilter, page, limit)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve join requests: "+err.Error())
		return
	}
	responses.SendPaginated(c, http.StatusOK, "Join requests retrieved successfully", requests, total, page, limit)
}

// GetMyJoinRequests godoc
// @Summary Get my join requests
// @Description Retrieves all join requests made by the authenticated user.
// @Tags Join Requests
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status (e.g., 'pending', 'approved', 'rejected')"
// @Success 200 {object} responses.PaginatedResponse{data=[]JoinRequest} "List of my join requests"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /users/me/join-requests [get]
func (tc *TeamController) GetMyJoinRequests(c *gin.Context) {
	userID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	statusFilter := strings.ToLower(c.Query("status"))

	requests, total, err := tc.repo.GetJoinRequestsByUserID(userID, statusFilter, page, limit)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve your join requests: "+err.Error())
		return
	}
	responses.SendPaginated(c, http.StatusOK, "Your join requests retrieved successfully", requests, total, page, limit)
}

// RespondToJoinRequest godoc
// @Summary Respond to a join request (Approve/Reject)
// @Description Allows a team manager (creator, captain, vice-captain, moderator) to approve or reject a join request.
// @Tags Join Requests
// @Produce json
// @Param team_id path uint true "Team ID"
// @Param request_id path uint true "Join Request ID"
// @Param action path string true "Action to perform: 'approve' or 'reject'"
// @Success 200 {object} responses.SuccessResponse "Join request processed"
// @Failure 400 {object} responses.ErrorResponse "Invalid input, ID(s), or action"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden - Insufficient permissions or request not pending"
// @Failure 404 {object} responses.ErrorResponse "Team or join request not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /teams/{team_id}/join-requests/{request_id}/{action} [put]
func (tc *TeamController) RespondToJoinRequest(c *gin.Context) {
	currentUserID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid team ID")
		return
	}
	requestID, err := strconv.ParseUint(c.Param("request_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request ID")
		return
	}
	action := strings.ToLower(c.Param("action"))
	if action != "approve" && action != "reject" {
		responses.SendError(c, http.StatusBadRequest, "Invalid action. Must be 'approve' or 'reject'.")
		return
	}

	team, err := tc.repo.GetTeamByID(uint(teamID))
	if err != nil || team == nil || team.IsDeleted {
		responses.SendError(c, http.StatusNotFound, "Team not found")
		return
	}

	isManager, err := tc.isTeamManager(uint(teamID), currentUserID)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Error checking permissions: "+err.Error())
		return
	}
	if !isManager {
		responses.SendError(c, http.StatusForbidden, "Only team managers can respond to join requests")
		return
	}

	joinRequest, err := tc.repo.GetJoinRequestByID(uint(requestID))
	if err != nil || joinRequest == nil {
		responses.SendError(c, http.StatusNotFound, "Join request not found")
		return
	}

	if joinRequest.TeamID != uint(teamID) {
		responses.SendError(c, http.StatusBadRequest, "Join request does not belong to this team")
		return
	}
	if joinRequest.Status != StatusPending {
		responses.SendError(c, http.StatusForbidden, "Join request is not pending, cannot be processed.")
		return
	}

	if action == "approve" {
		// Check team max player limit
		currentMembers, _, _ := tc.repo.GetTeamMembers(uint(teamID), 1, team.MaxPlayers+1) // get all members
		if len(currentMembers) >= team.MaxPlayers {
			responses.SendError(c, http.StatusForbidden, "Team has reached its maximum player capacity.")
			return
		}

		joinRequest.Status = StatusApproved
		newMember := TeamMember{
			TeamID:   joinRequest.TeamID,
			UserID:   joinRequest.UserID,
			Role:     RolePlayer, // Default role on approval, can be changed later
			Position: joinRequest.Position,
			JoinedAt: time.Now(),
			IsActive: true,
		}

		// Transaction to update request and add member
		txErr := tc.repo.WithTransaction(func(repo TeamRepository) error {
			if err := repo.UpdateJoinRequest(joinRequest); err != nil {
				return err
			}
			// Use the repository method for adding member
			newMember := TeamMember{
				TeamID:   joinRequest.TeamID,
				UserID:   joinRequest.UserID,
				Role:     RolePlayer,
				Position: joinRequest.Position,
				JoinedAt: time.Now(),
				IsActive: true,
			}
			return repo.Create(&newMember)
		})
		if txErr != nil {
			responses.SendError(c, http.StatusInternalServerError, "Failed to approve join request: "+txErr.Error())
			return
		}
		responses.SendSuccess(c, http.StatusOK, "Join request approved and member added", joinRequest)

	} else { // action == "reject"
		joinRequest.Status = StatusRejected
		if err := tc.repo.UpdateJoinRequest(joinRequest); err != nil {
			responses.SendError(c, http.StatusInternalServerError, "Failed to reject join request: "+err.Error())
			return
		}
		responses.SendSuccess(c, http.StatusOK, "Join request rejected", joinRequest)
	}
}

// CancelJoinRequest godoc
// @Summary Cancel a join request
// @Description Allows a user to cancel their own pending join request.
// @Tags Join Requests
// @Produce json
// @Param request_id path uint true "Join Request ID"
// @Success 200 {object} responses.SuccessResponse "Join request cancelled"
// @Failure 400 {object} responses.ErrorResponse "Invalid request ID"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden - Not owner or request not pending"
// @Failure 404 {object} responses.ErrorResponse "Join request not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /join-requests/{request_id} [delete]
func (tc *TeamController) CancelJoinRequest(c *gin.Context) {
	userID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	requestID, err := strconv.ParseUint(c.Param("request_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request ID")
		return
	}

	joinRequest, err := tc.repo.GetJoinRequestByID(uint(requestID))
	if err != nil || joinRequest == nil {
		responses.SendError(c, http.StatusNotFound, "Join request not found")
		return
	}

	if joinRequest.UserID != userID {
		responses.SendError(c, http.StatusForbidden, "You can only cancel your own join requests")
		return
	}
	if joinRequest.Status != StatusPending {
		responses.SendError(c, http.StatusForbidden, "Only pending join requests can be cancelled")
		return
	}

	// Instead of deleting, update status to 'cancelled' for history
	// Or, if requirement is hard delete:
	// if err := tc.repo.DeleteJoinRequest(uint(requestID)); err != nil { ... }

	joinRequest.Status = StatusCancelled
	if err := tc.repo.UpdateJoinRequest(joinRequest); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to cancel join request: "+err.Error())
		return
	}
	responses.SendSuccess(c, http.StatusOK, "Join request cancelled successfully", joinRequest)
}

// --- Team Invitation Handlers ---

// InviteUserToTeam godoc
// @Summary Invite a user to a team
// @Description Sends an invitation to a user to join a team. Only team managers can invite.
// @Tags Team Invitations
// @Accept json
// @Produce json
// @Param team_id path uint true "Team ID"
// @Param invite_request body InviteUserRequest true "Invitation Details"
// @Success 201 {object} responses.SuccessResponse{data=TeamInvitation} "Invitation sent successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid input or team ID"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden - Insufficient permissions, user already member, or pending request/invitation exists"
// @Failure 404 {object} responses.ErrorResponse "Team or user to invite not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /teams/{team_id}/invitations [post]
func (tc *TeamController) InviteUserToTeam(c *gin.Context) {
	currentUserID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	var req InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	if req.Role == "" {
		req.Role = RolePlayer // Default role if not specified
	}

	team, err := tc.repo.GetTeamByID(uint(teamID))
	if err != nil || team == nil || team.IsDeleted {
		responses.SendError(c, http.StatusNotFound, "Team not found")
		return
	}

	isManager, err := tc.isTeamManager(uint(teamID), currentUserID)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Error checking permissions: "+err.Error())
		return
	}
	if !isManager {
		responses.SendError(c, http.StatusForbidden, "Only team managers (creator, captain, vice-captain, moderator) can send invitations")
		return
	}

	// Check if invited user exists (optional, depends on user service availability)
	// _, err = tc.userRepo.GetUserByID(req.UserID)
	// if err != nil || invitedUser == nil {
	// 	responses.SendError(c, http.StatusNotFound, "User to invite not found")
	// 	return
	// }

	// Check if user is already a member
	isMember, _ := tc.repo.IsUserTeamMember(uint(teamID), req.UserID)
	if isMember {
		responses.SendError(c, http.StatusForbidden, "User is already a member of this team")
		return
	}

	// Check for existing pending invitation
	existingInvite, _ := tc.repo.GetPendingInvitation(uint(teamID), req.UserID)
	if existingInvite != nil {
		responses.SendError(c, http.StatusForbidden, "User already has a pending invitation for this team")
		return
	}

	// Check for existing pending join request from this user to this team
	existingJoinRequest, _ := tc.repo.GetPendingJoinRequest(uint(teamID), req.UserID)
	if existingJoinRequest != nil {
		responses.SendError(c, http.StatusForbidden, "This user has a pending join request for your team. Please process it instead.")
		return
	}

	// Check team max player limit
	currentMembers, _, _ := tc.repo.GetTeamMembers(uint(teamID), 1, team.MaxPlayers+1)
	if len(currentMembers) >= team.MaxPlayers {
		responses.SendError(c, http.StatusForbidden, "Team has reached its maximum player capacity. Cannot invite more players.")
		return
	}

	invitation := TeamInvitation{
		TeamID:    uint(teamID),
		UserID:    req.UserID,
		Role:      req.Role,
		Position:  req.Position,
		Message:   req.Message,
		Status:    StatusPending,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Example: 7 days expiry
	}

	if err := tc.repo.CreateTeamInvitation(&invitation); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to send invitation: "+err.Error())
		return
	}
	responses.SendSuccess(c, http.StatusCreated, "Invitation sent successfully", invitation)
}

// GetInvitationsForTeam godoc
// @Summary Get invitations sent by a team
// @Description Retrieves invitations sent by a specific team. Only for team managers.
// @Tags Team Invitations
// @Produce json
// @Param team_id path uint true "Team ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status (e.g., 'pending', 'accepted', 'rejected')"
// @Success 200 {object} responses.PaginatedResponse{data=[]TeamInvitation} "List of team invitations"
// @Failure 400 {object} responses.ErrorResponse "Invalid team ID"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden - Insufficient permissions"
// @Failure 404 {object} responses.ErrorResponse "Team not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /teams/{team_id}/invitations [get]
func (tc *TeamController) GetInvitationsForTeam(c *gin.Context) {
	currentUserID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	team, err := tc.repo.GetTeamByID(uint(teamID))
	if err != nil || team == nil || team.IsDeleted {
		responses.SendError(c, http.StatusNotFound, "Team not found")
		return
	}

	isManager, err := tc.isTeamManager(uint(teamID), currentUserID)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Error checking permissions: "+err.Error())
		return
	}
	if !isManager {
		responses.SendError(c, http.StatusForbidden, "Only team managers can view team invitations")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	statusFilter := strings.ToLower(c.Query("status"))

	invitations, total, err := tc.repo.GetTeamInvitationsByTeamID(uint(teamID), statusFilter, page, limit)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve team invitations: "+err.Error())
		return
	}
	responses.SendPaginated(c, http.StatusOK, "Team invitations retrieved successfully", invitations, total, page, limit)
}

// GetMyTeamInvitations godoc
// @Summary Get my team invitations
// @Description Retrieves all team invitations received by the authenticated user.
// @Tags Team Invitations
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status (e.g., 'pending', 'accepted', 'rejected')"
// @Success 200 {object} responses.PaginatedResponse{data=[]TeamInvitation} "List of my team invitations"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /users/me/invitations [get]
func (tc *TeamController) GetMyTeamInvitations(c *gin.Context) {
	userID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	statusFilter := strings.ToLower(c.DefaultQuery("status", StatusPending)) // Default to pending

	invitations, total, err := tc.repo.GetTeamInvitationsByUserID(userID, statusFilter, page, limit)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve your team invitations: "+err.Error())
		return
	}
	responses.SendPaginated(c, http.StatusOK, "Your team invitations retrieved successfully", invitations, total, page, limit)
}

// RespondToTeamInvitation godoc
// @Summary Respond to a team invitation (Accept/Reject)
// @Description Allows an authenticated user to accept or reject a team invitation.
// @Tags Team Invitations
// @Produce json
// @Param invitation_id path uint true "Invitation ID"
// @Param action path string true "Action to perform: 'accept' or 'reject'"
// @Success 200 {object} responses.SuccessResponse "Invitation processed"
// @Failure 400 {object} responses.ErrorResponse "Invalid input, ID, or action"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden - Not recipient, invitation not pending, or team full"
// @Failure 404 {object} responses.ErrorResponse "Invitation not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /invitations/{invitation_id}/{action} [put]
func (tc *TeamController) RespondToTeamInvitation(c *gin.Context) {
	userID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	invitationID, err := strconv.ParseUint(c.Param("invitation_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid invitation ID")
		return
	}
	action := strings.ToLower(c.Param("action"))
	if action != "accept" && action != "reject" {
		responses.SendError(c, http.StatusBadRequest, "Invalid action. Must be 'accept' or 'reject'.")
		return
	}

	invitation, err := tc.repo.GetTeamInvitationByID(uint(invitationID))
	if err != nil || invitation == nil {
		responses.SendError(c, http.StatusNotFound, "Invitation not found")
		return
	}

	if invitation.UserID != userID {
		responses.SendError(c, http.StatusForbidden, "This invitation is not for you")
		return
	}
	if invitation.Status != StatusPending {
		responses.SendError(c, http.StatusForbidden, "Invitation is not pending, cannot be processed.")
		return
	}
	if time.Now().After(invitation.ExpiresAt) {
		invitation.Status = StatusRejected // Or a new "expired" status
		tc.repo.UpdateTeamInvitation(invitation)
		responses.SendError(c, http.StatusForbidden, "Invitation has expired.")
		return
	}

	if action == "accept" {
		team, err := tc.repo.GetTeamByID(invitation.TeamID)
		if err != nil || team == nil || team.IsDeleted {
			responses.SendError(c, http.StatusNotFound, "Associated team not found or has been deleted")
			return
		}
		// Check team max player limit
		currentMembers, _, _ := tc.repo.GetTeamMembers(invitation.TeamID, 1, team.MaxPlayers+1)
		if len(currentMembers) >= team.MaxPlayers {
			invitation.Status = StatusRejected // Or a new "team_full" status
			tc.repo.UpdateTeamInvitation(invitation)
			responses.SendError(c, http.StatusForbidden, "Team has reached its maximum player capacity. Cannot join.")
			return
		}

		invitation.Status = StatusAccepted
		newMember := TeamMember{
			TeamID:    invitation.TeamID,
			UserID:    invitation.UserID,
			Role:      invitation.Role,
			Position:  invitation.Position,
			JoinedAt:  time.Now(),
			IsActive:  true,
			IsCaptain: (invitation.Role == RoleCaptain), // Set IsCaptain if role is captain
		}

		txErr := tc.repo.WithTransaction(func(repo TeamRepository) error {
			if err := tx.Save(invitation).Error; err != nil {
				return err
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "team_id"}, {Name: "user_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"role", "position", "is_active", "is_captain", "joined_at", "updated_at"}),
			}).Create(&newMember).Error; err != nil {
				return err
			}
			return nil
		})

		if txErr != nil {
			responses.SendError(c, http.StatusInternalServerError, "Failed to accept invitation: "+txErr.Error())
			return
		}
		responses.SendSuccess(c, http.StatusOK, "Invitation accepted and you have joined the team", invitation)

	} else { // action == "reject"
		invitation.Status = StatusRejected
		if err := tc.repo.UpdateTeamInvitation(invitation); err != nil {
			responses.SendError(c, http.StatusInternalServerError, "Failed to reject invitation: "+err.Error())
			return
		}
		responses.SendSuccess(c, http.StatusOK, "Invitation rejected", invitation)
	}
}

// CancelTeamInvitation godoc
// @Summary Cancel a team invitation
// @Description Allows a team manager who sent an invitation to cancel it if it's still pending.
// @Tags Team Invitations
// @Produce json
// @Param invitation_id path uint true "Invitation ID"
// @Success 200 {object} responses.SuccessResponse "Invitation cancelled"
// @Failure 400 {object} responses.ErrorResponse "Invalid invitation ID"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden - Not manager of the team or invitation not pending"
// @Failure 404 {object} responses.ErrorResponse "Invitation not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /invitations/{invitation_id} [delete]
func (tc *TeamController) CancelTeamInvitation(c *gin.Context) {
	currentUserID, authenticated := getCurrentUserID(c)
	if !authenticated {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	invitationID, err := strconv.ParseUint(c.Param("invitation_id"), 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid invitation ID")
		return
	}

	invitation, err := tc.repo.GetTeamInvitationByID(uint(invitationID))
	if err != nil || invitation == nil {
		responses.SendError(c, http.StatusNotFound, "Invitation not found")
		return
	}

	// Authorization: Check if current user is a manager of the team that sent the invitation
	isManager, err := tc.isTeamManager(invitation.TeamID, currentUserID)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Error checking permissions: "+err.Error())
		return
	}
	if !isManager {
		responses.SendError(c, http.StatusForbidden, "Only managers of the inviting team can cancel the invitation")
		return
	}

	if invitation.Status != StatusPending {
		responses.SendError(c, http.StatusForbidden, "Only pending invitations can be cancelled")
		return
	}

	// Instead of deleting, update status to 'cancelled' for history
	// Or, if requirement is hard delete:
	// if err := tc.repo.DeleteTeamInvitation(uint(invitationID)); err != nil { ... }
	invitation.Status = StatusCancelled
	if err := tc.repo.UpdateTeamInvitation(invitation); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to cancel invitation: "+err.Error())
		return
	}
	responses.SendSuccess(c, http.StatusOK, "Team invitation cancelled successfully", invitation)
}

// --- Admin Specific Endpoints (Example) ---

// AdminGetAllTeams godoc
// @Summary (Admin) Get all teams
// @Description (Admin) Retrieves a list of all teams, including soft-deleted ones if specified.
// @Tags Admin-Teams
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param include_deleted query bool false "Include soft-deleted teams" default(false)
// @Success 200 {object} responses.PaginatedResponse{data=[]Team} "List of all teams"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden - Admin access required"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /admin/teams [get]
func (tc *TeamController) AdminGetAllTeams(c *gin.Context) {
	if !isAdminUser(c) {
		responses.SendError(c, http.StatusForbidden, "Admin access required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	includeDeleted, _ := strconv.ParseBool(c.DefaultQuery("include_deleted", "false"))

	teams, total, err := tc.repo.GetAllTeamsAdmin(page, limit, includeDeleted)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve teams: "+err.Error())
		return
	}
	responses.SendPaginated(c, http.StatusOK, "All teams retrieved successfully", teams, total, page, limit)
}
