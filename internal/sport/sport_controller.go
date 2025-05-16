package sport

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/DhavalSuthar-24/miow/config"

	"github.com/DhavalSuthar-24/miow/internal/middleware" // Your middleware package
	"github.com/DhavalSuthar-24/miow/pkg/responses"       // A common responses package (you might need to create this)
	"github.com/DhavalSuthar-24/miow/pkg/validator"       // A common validator package (you might need to create this)
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SportController handles API requests related to sports.
type SportController struct {
	repo   SportRepository
	config *config.Config // If needed for specific configurations
}

// NewSportController creates a new SportController.
func NewSportController(repo SportRepository, cfg *config.Config) *SportController {
	return &SportController{
		repo:   repo,
		config: cfg,
	}
}

// --- DTOs (Data Transfer Objects) for requests/responses ---

type CreateSportRequest struct {
	Name        string    `json:"name" binding:"required,min=3,max=100"`
	Description string    `json:"description" binding:"omitempty,max=5000"`
	Icon        string    `json:"icon" binding:"omitempty,url|uri,max=255"`
	IsActive    *bool     `json:"is_active" binding:"omitempty"` // Pointer to distinguish between not provided and false
	Rules       Rules     `json:"rules,omitempty"`
	Positions   Positions `json:"positions,omitempty"`
	Equipment   Equipment `json:"equipment,omitempty"`
}

type UpdateSportRequest struct {
	Name        string     `json:"name" binding:"omitempty,min=3,max=100"`
	Description string     `json:"description" binding:"omitempty,max=5000"`
	Icon        string     `json:"icon" binding:"omitempty,url|uri,max=255"`
	IsActive    *bool      `json:"is_active" binding:"omitempty"`
	Rules       *Rules     `json:"rules,omitempty"` // Pointer to allow partial update of complex fields
	Positions   *Positions `json:"positions,omitempty"`
	Equipment   *Equipment `json:"equipment,omitempty"`
}

type CreateSkillRequest struct {
	Name        string  `json:"name" binding:"required,min=3,max=100"`
	Description string  `json:"description" binding:"omitempty,max=1000"`
	Weight      float64 `json:"weight" binding:"omitempty,min=0,max=10"`
}

type UpdateSkillRequest struct {
	Name        string  `json:"name" binding:"omitempty,min=3,max=100"`
	Description string  `json:"description" binding:"omitempty,max=1000"`
	Weight      float64 `json:"weight" binding:"omitempty,min=0,max=10"`
}

type UserSportRequest struct {
	SportID  uint   `json:"sport_id" binding:"required"`
	Position string `json:"position" binding:"omitempty,max=100"`
	Level    string `json:"level" binding:"omitempty,max=50"` // e.g., "Beginner", "Intermediate"
}

// --- Sport Handlers ---

// CreateSport godoc
// @Summary Create a new sport
// @Description Admin can create a new sport
// @Tags Sports
// @Accept json
// @Produce json
// @Param sport body CreateSportRequest true "Sport creation request"
// @Success 201 {object} responses.SuccessResponse{data=Sport}
// @Failure 400 {object} responses.ErrorResponse "Validation error or bad request"
// @Failure 409 {object} responses.ErrorResponse "Sport with this name already exists"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /sports [post]
// @Security BearerAuth
func (sc *SportController) CreateSport(c *gin.Context) {
	var req CreateSportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors := validator.ParseError(err)
		responses.SendError(c, http.StatusBadRequest, "Validation failed", errors)
		return
	}

	existingSport, _ := sc.repo.FindSportByName(req.Name)
	if existingSport != nil {
		responses.SendError(c, http.StatusConflict, "Sport with this name already exists", nil)
		return
	}

	sport := Sport{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		Rules:       req.Rules,
		Positions:   req.Positions,
		// Equipment:   req.Equipment, // Typo in model, should be plural Equipments
	}
	if req.IsActive != nil {
		sport.IsActive = *req.IsActive
	} else {
		sport.IsActive = true // Default to active
	}

	if err := sc.repo.CreateSport(&sport); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to create sport", err.Error())
		return
	}

	responses.SendSuccess(c, http.StatusCreated, "Sport created successfully", sport)
}

// GetAllSports godoc
// @Summary Get all sports
// @Description Get a list of all available sports with optional filters
// @Tags Sports
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Number of items per page" default(10)
// @Param search query string false "Search term for name or description"
// @Param is_active query boolean false "Filter by active status (admin only)"
// @Success 200 {object} responses.PaginatedResponse{data=[]Sport}
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /sports [get]
func (sc *SportController) GetAllSports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	searchTerm := c.Query("search")

	var isActiveFilter *bool
	isActiveQuery := c.Query("is_active")
	if isActiveQuery != "" {
		val, err := strconv.ParseBool(isActiveQuery)
		// Only allow admin to filter by is_active=false.
		// This logic should ideally be in a service layer or here if controller handles authorization nuances.
		// For simplicity, if 'is_active' is provided, we use it. Secure this if non-admins shouldn't see inactive.
		// We assume RoleMiddleware has already passed.
		if err == nil {
			isActiveFilter = &val
		}
	}

	sports, total, err := sc.repo.GetAllSports(page, pageSize, searchTerm, isActiveFilter)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve sports", err.Error())
		return
	}

	responses.SendPaginated(c, http.StatusOK, "Sports retrieved successfully", sports, total, page, pageSize)
}

// GetSportByID godoc
// @Summary Get a sport by ID
// @Description Get details of a specific sport by its ID
// @Tags Sports
// @Produce json
// @Param sport_id path int true "Sport ID"
// @Success 200 {object} responses.SuccessResponse{data=Sport}
// @Failure 404 {object} responses.ErrorResponse "Sport not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /sports/{sport_id} [get]
func (sc *SportController) GetSportByID(c *gin.Context) {
	sportIDStr := c.Param("sport_id")
	sportID, err := strconv.ParseUint(sportIDStr, 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid sport ID format", nil)
		return
	}

	sport, err := sc.repo.GetSportByID(uint(sportID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || sport == nil { // Check both error and nil sport
			responses.SendError(c, http.StatusNotFound, "Sport not found", nil)
			return
		}
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve sport", err.Error())
		return
	}
	if sport == nil { // Double check after error handling
		responses.SendError(c, http.StatusNotFound, "Sport not found", nil)
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Sport retrieved successfully", sport)
}

// UpdateSport godoc
// @Summary Update a sport
// @Description Admin can update an existing sport's details
// @Tags Sports
// @Accept json
// @Produce json
// @Param sport_id path int true "Sport ID"
// @Param sport body UpdateSportRequest true "Sport update request"
// @Success 200 {object} responses.SuccessResponse{data=Sport}
// @Failure 400 {object} responses.ErrorResponse "Validation error or bad request"
// @Failure 404 {object} responses.ErrorResponse "Sport not found"
// @Failure 409 {object} responses.ErrorResponse "Sport with this name already exists (if name changed)"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /sports/{sport_id} [put]
// @Security BearerAuth
func (sc *SportController) UpdateSport(c *gin.Context) {
	sportIDStr := c.Param("sport_id")
	sportID, err := strconv.ParseUint(sportIDStr, 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid sport ID format", nil)
		return
	}

	var req UpdateSportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors := validator.ParseError(err)
		responses.SendError(c, http.StatusBadRequest, "Validation failed", errors)
		return
	}

	sport, err := sc.repo.GetSportByID(uint(sportID))
	if err != nil || sport == nil {
		responses.SendError(c, http.StatusNotFound, "Sport not found", nil)
		return
	}

	if req.Name != "" && req.Name != sport.Name {
		existingSport, _ := sc.repo.FindSportByName(req.Name)
		if existingSport != nil && existingSport.ID != sport.ID {
			responses.SendError(c, http.StatusConflict, "Another sport with this name already exists", nil)
			return
		}
		sport.Name = req.Name
	}
	if req.Description != "" {
		sport.Description = req.Description
	}
	if req.Icon != "" {
		sport.Icon = req.Icon
	}
	if req.IsActive != nil {
		sport.IsActive = *req.IsActive
	}
	if req.Rules != nil {
		sport.Rules = *req.Rules
	}
	if req.Positions != nil {
		sport.Positions = *req.Positions
	}
	// if req.Equipment != nil {
	//     sport.Equipment = *req.Equipment // Typo in model
	// }

	if err := sc.repo.UpdateSport(sport); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to update sport", err.Error())
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Sport updated successfully", sport)
}

// DeleteSport godoc
// @Summary Delete a sport
// @Description Admin can delete a sport (and its associated skills due to DB constraints)
// @Tags Sports
// @Produce json
// @Param sport_id path int true "Sport ID"
// @Success 200 {object} responses.SuccessResponse "Sport deleted successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid sport ID"
// @Failure 404 {object} responses.ErrorResponse "Sport not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /sports/{sport_id} [delete]
// @Security BearerAuth
func (sc *SportController) DeleteSport(c *gin.Context) {
	sportIDStr := c.Param("sport_id")
	sportID, err := strconv.ParseUint(sportIDStr, 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid sport ID format", nil)
		return
	}

	sport, err := sc.repo.GetSportByID(uint(sportID))
	if err != nil || sport == nil { // Check both error and nil sport
		responses.SendError(c, http.StatusNotFound, "Sport not found to delete", nil)
		return
	}

	if err := sc.repo.DeleteSport(uint(sportID)); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to delete sport", err.Error())
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Sport deleted successfully", nil)
}

// --- Skill Handlers ---

// AddSkillToSport godoc
// @Summary Add a skill to a sport
// @Description Admin can add a new skill to a specific sport
// @Tags Skills
// @Accept json
// @Produce json
// @Param sport_id path int true "Sport ID"
// @Param skill body CreateSkillRequest true "Skill creation request"
// @Success 201 {object} responses.SuccessResponse{data=Skill}
// @Failure 400 {object} responses.ErrorResponse "Validation error or bad request"
// @Failure 404 {object} responses.ErrorResponse "Sport not found"
// @Failure 409 {object} responses.ErrorResponse "Skill with this name already exists for this sport"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /sports/{sport_id}/skills [post]
// @Security BearerAuth
func (sc *SportController) AddSkillToSport(c *gin.Context) {
	sportIDStr := c.Param("sport_id")
	sportID, err := strconv.ParseUint(sportIDStr, 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid sport ID format", nil)
		return
	}

	var req CreateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors := validator.ParseError(err)
		responses.SendError(c, http.StatusBadRequest, "Validation failed", errors)
		return
	}

	sport, err := sc.repo.GetSportByID(uint(sportID))
	if err != nil || sport == nil {
		responses.SendError(c, http.StatusNotFound, "Sport not found", nil)
		return
	}

	existingSkill, _ := sc.repo.FindSkillByNameAndSportID(req.Name, uint(sportID))
	if existingSkill != nil {
		responses.SendError(c, http.StatusConflict, "Skill with this name already exists for this sport", nil)
		return
	}

	skill := Skill{
		Name:        req.Name,
		Description: req.Description,
		SportID:     uint(sportID),
		Weight:      req.Weight,
	}
	if skill.Weight == 0 { // Default weight if not provided or zero
		skill.Weight = 1.0
	}

	if err := sc.repo.CreateSkill(&skill); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to add skill", err.Error())
		return
	}

	responses.SendSuccess(c, http.StatusCreated, "Skill added successfully", skill)
}

// GetSkillsForSport godoc
// @Summary Get skills for a sport
// @Description Get all skills associated with a specific sport
// @Tags Skills
// @Produce json
// @Param sport_id path int true "Sport ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Number of items per page" default(10)
// @Success 200 {object} responses.PaginatedResponse{data=[]Skill}
// @Failure 400 {object} responses.ErrorResponse "Invalid sport ID"
// @Failure 404 {object} responses.ErrorResponse "Sport not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /sports/{sport_id}/skills [get]
func (sc *SportController) GetSkillsForSport(c *gin.Context) {
	sportIDStr := c.Param("sport_id")
	sportID, err := strconv.ParseUint(sportIDStr, 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid sport ID format", nil)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	sport, err := sc.repo.GetSportByID(uint(sportID))
	if err != nil || sport == nil {
		responses.SendError(c, http.StatusNotFound, "Sport not found", nil)
		return
	}

	skills, total, err := sc.repo.GetSkillsBySportID(uint(sportID), page, pageSize)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve skills", err.Error())
		return
	}

	responses.SendPaginated(c, http.StatusOK, "Skills retrieved successfully", skills, total, page, pageSize)
}

// UpdateSkill godoc
// @Summary Update a skill
// @Description Admin can update an existing skill's details
// @Tags Skills
// @Accept json
// @Produce json
// @Param skill_id path int true "Skill ID"
// @Param skill body UpdateSkillRequest true "Skill update request"
// @Success 200 {object} responses.SuccessResponse{data=Skill}
// @Failure 400 {object} responses.ErrorResponse "Validation error or bad request"
// @Failure 404 {object} responses.ErrorResponse "Skill not found"
// @Failure 409 {object} responses.ErrorResponse "Skill with this name already exists for its sport"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /skills/{skill_id} [put]
// @Security BearerAuth
func (sc *SportController) UpdateSkill(c *gin.Context) {
	skillIDStr := c.Param("skill_id")
	skillID, err := strconv.ParseUint(skillIDStr, 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid skill ID format", nil)
		return
	}

	var req UpdateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors := validator.ParseError(err)
		responses.SendError(c, http.StatusBadRequest, "Validation failed", errors)
		return
	}

	skill, err := sc.repo.GetSkillByID(uint(skillID))
	if err != nil || skill == nil {
		responses.SendError(c, http.StatusNotFound, "Skill not found", nil)
		return
	}

	if req.Name != "" && req.Name != skill.Name {
		existingSkill, _ := sc.repo.FindSkillByNameAndSportID(req.Name, skill.SportID)
		if existingSkill != nil && existingSkill.ID != skill.ID {
			responses.SendError(c, http.StatusConflict, "Another skill with this name already exists for this sport", nil)
			return
		}
		skill.Name = req.Name
	}
	if req.Description != "" { // Allow clearing description
		skill.Description = req.Description
	}
	if req.Weight != 0 { // Check if weight is provided explicitly
		skill.Weight = req.Weight
	}

	if err := sc.repo.UpdateSkill(skill); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to update skill", err.Error())
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Skill updated successfully", skill)
}

// DeleteSkill godoc
// @Summary Delete a skill
// @Description Admin can delete a skill
// @Tags Skills
// @Produce json
// @Param skill_id path int true "Skill ID"
// @Success 200 {object} responses.SuccessResponse "Skill deleted successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid skill ID"
// @Failure 404 {object} responses.ErrorResponse "Skill not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /skills/{skill_id} [delete]
// @Security BearerAuth
func (sc *SportController) DeleteSkill(c *gin.Context) {
	skillIDStr := c.Param("skill_id")
	skillID, err := strconv.ParseUint(skillIDStr, 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid skill ID format", nil)
		return
	}

	skill, errRepo := sc.repo.GetSkillByID(uint(skillID))
	if errRepo != nil || skill == nil {
		responses.SendError(c, http.StatusNotFound, "Skill not found to delete", nil)
		return
	}

	if err := sc.repo.DeleteSkill(uint(skillID)); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to delete skill", err.Error())
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Skill deleted successfully", nil)
}

// --- UserSport Handlers ---

// AddUserSportPreference godoc
// @Summary Add or update a sport preference for the logged-in user
// @Description Authenticated user can add or update their preference for a sport, including position and level
// @Tags UserSports
// @Accept json
// @Produce json
// @Param preference body UserSportRequest true "User sport preference request"
// @Success 200 {object} responses.SuccessResponse{data=UserSport} "Preference updated"
// @Success 201 {object} responses.SuccessResponse{data=UserSport} "Preference added"
// @Failure 400 {object} responses.ErrorResponse "Validation error or bad request"
// @Failure 404 {object} responses.ErrorResponse "Sport not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /users/me/sports [post]
// @Security BearerAuth
func (sc *SportController) AddUserSportPreference(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		responses.SendError(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	var req UserSportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors := validator.ParseError(err)
		responses.SendError(c, http.StatusBadRequest, "Validation failed", errors)
		return
	}

	// Check if sport exists
	sport, err := sc.repo.GetSportByID(req.SportID)
	if err != nil || sport == nil {
		responses.SendError(c, http.StatusNotFound, "Sport not found", nil)
		return
	}

	userSport := UserSport{
		UserID:   userID,
		SportID:  req.SportID,
		Position: req.Position,
		Level:    req.Level,
	}

	// Use AddUserSport which handles upsert logic
	if err := sc.repo.AddUserSport(&userSport); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to add/update user sport preference", err.Error())
		return
	}

	// Fetch the newly created/updated record to include sport details if needed
	createdOrUpdatedUserSport, err := sc.repo.GetUserSportBySportID(userID, req.SportID)
	if err != nil || createdOrUpdatedUserSport == nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve user sport preference after update", err.Error())
		return
	}

	// Determine if it was a create or update for the status code (optional)
	// For simplicity, we can return 200 OK for upsert. GORM doesn't easily tell if it was an insert or update in this upsert.
	responses.SendSuccess(c, http.StatusOK, "User sport preference saved successfully", createdOrUpdatedUserSport)
}

// GetUserSportPreferences godoc
// @Summary Get logged-in user's sport preferences
// @Description Authenticated user can retrieve their list of sport preferences
// @Tags UserSports
// @Produce json
// @Success 200 {object} responses.SuccessResponse{data=[]UserSport}
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /users/me/sports [get]
// @Security BearerAuth
func (sc *SportController) GetUserSportPreferences(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		responses.SendError(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	preferences, err := sc.repo.GetUserSports(userID)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve user sport preferences", err.Error())
		return
	}

	responses.SendSuccess(c, http.StatusOK, "User sport preferences retrieved successfully", preferences)
}

// RemoveUserSportPreference godoc
// @Summary Remove a sport preference for the logged-in user
// @Description Authenticated user can remove one of their sport preferences
// @Tags UserSports
// @Produce json
// @Param sport_id path int true "Sport ID to remove preference for"
// @Success 200 {object} responses.SuccessResponse "Preference removed successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid sport ID"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 404 {object} responses.ErrorResponse "Sport preference not found for this user"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /users/me/sports/{sport_id} [delete]
// @Security BearerAuth
func (sc *SportController) RemoveUserSportPreference(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		responses.SendError(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	sportIDStr := c.Param("sport_id")
	sportID, err := strconv.ParseUint(sportIDStr, 10, 32)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid sport ID format", nil)
		return
	}

	// Check if the preference exists before deleting
	existingPreference, err := sc.repo.GetUserSportBySportID(userID, uint(sportID))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		responses.SendError(c, http.StatusInternalServerError, "Error checking preference", err.Error())
		return
	}
	if existingPreference == nil {
		responses.SendError(c, http.StatusNotFound, "Sport preference not found for this user", nil)
		return
	}

	if err := sc.repo.RemoveUserSport(userID, uint(sportID)); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to remove user sport preference", err.Error())
		return
	}

	responses.SendSuccess(c, http.StatusOK, "User sport preference removed successfully", nil)
}
