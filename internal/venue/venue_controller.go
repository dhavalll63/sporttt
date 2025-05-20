package venue

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DhavalSuthar-24/miow/config"
	"github.com/DhavalSuthar-24/miow/pkg/utils"
	"github.com/gin-gonic/gin"
)

// VenueController handles venue-related HTTP requests
type VenueController struct {
	repo      VenueRepository
	appConfig *config.Config
}

// NewVenueController creates a new venue controller
func NewVenueController(repo VenueRepository, appConfig *config.Config) *VenueController {
	return &VenueController{
		repo:      repo,
		appConfig: appConfig,
	}
}

// CreateVenue godoc
// @Summary Create a new venue
// @Description Create a new venue with the provided details
// @Tags venues
// @Accept json
// @Produce json
// @Param venue body VenueInput true "Venue information"
// @Success 201 {object} Venue "Venue created successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid input"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /manager/venues [post]
// @Security Bearer
func (c *VenueController) CreateVenue(ctx *gin.Context) {
	var input VenueInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "unauthorized"})
		return
	}

	// Create venue object
	venue := &Venue{
		Name:        input.Name,
		Location:    input.Location,
		Coordinates: input.Coordinates,
		Facilities:  input.Facilities,
		Available:   input.Available,
		ContactInfo: input.ContactInfo,
		Description: input.Description,
		Images:      input.Images,
		Capacity:    input.Capacity,
		HourlyRate:  input.HourlyRate,
		CourtCount:  input.CourtCount,
		SocialHours: input.SocialHours,
		ManagerID:   userID.(uint),
	}

	// Save venue to database
	if err := c.repo.CreateVenue(venue); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to create venue: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, venue)
}

// GetVenueByID godoc
// @Summary Get venue by ID
// @Description Get detailed information about a venue by its ID
// @Tags venues
// @Produce json
// @Param venue_id path int true "Venue ID"
// @Success 200 {object} Venue "Venue details"
// @Failure 400 {object} utils.ErrorResponse "Invalid venue ID"
// @Failure 404 {object} utils.ErrorResponse "Venue not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /venues/{venue_id} [get]
func (c *VenueController) GetVenueByID(ctx *gin.Context) {
	venueID, err := strconv.ParseUint(ctx.Param("venue_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid venue ID"})
		return
	}

	venue, err := c.repo.GetVenueByID(uint(venueID))
	if err != nil {
		if err.Error() == "venue not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "venue not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get venue: " + err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, venue)
}

// GetAllVenues godoc
// @Summary Get all venues
// @Description Get a paginated list of all venues with optional filters
// @Tags venues
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 10, max: 100)"
// @Param available query boolean false "Filter by availability"
// @Param location query string false "Filter by location (partial match)"
// @Param min_courts query int false "Filter by minimum number of courts"
// @Param max_price query number false "Filter by maximum hourly rate"
// @Success 200 {object} utils.PaginatedResponse{data=[]Venue} "List of venues"
// @Failure 400 {object} utils.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /venues [get]
func (c *VenueController) GetAllVenues(ctx *gin.Context) {
	var pagination PaginationInput
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
		return
	}

	// Build filters
	filters := make(map[string]interface{})

	// Check if available filter is provided
	if availableStr := ctx.Query("available"); availableStr != "" {
		available, err := strconv.ParseBool(availableStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid available parameter"})
			return
		}
		filters["available"] = available
	}

	// Check if location filter is provided
	if location := ctx.Query("location"); location != "" {
		filters["location"] = location
	}

	// Check if min_courts filter is provided
	if minCourtsStr := ctx.Query("min_courts"); minCourtsStr != "" {
		minCourts, err := strconv.Atoi(minCourtsStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid min_courts parameter"})
			return
		}
		filters["min_courts"] = minCourts
	}

	// Check if max_price filter is provided
	if maxPriceStr := ctx.Query("max_price"); maxPriceStr != "" {
		maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid max_price parameter"})
			return
		}
		filters["max_price"] = maxPrice
	}

	venues, totalCount, err := c.repo.GetAllVenues(pagination.Page, pagination.Limit, filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get venues: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, utils.PaginatedResponse{
		Data: venues,
		Pagination: utils.PaginationData{
			Total:      totalCount,
			Page:       pagination.Page,
			Limit:      pagination.Limit,
			TotalPages: int64((int(totalCount) + pagination.Limit - 1) / pagination.Limit),
		},
	})

}

// UpdateVenue godoc
// @Summary Update venue
// @Description Update an existing venue's details
// @Tags venues
// @Accept json
// @Produce json
// @Param venue_id path int true "Venue ID"
// @Param venue body VenueInput true "Updated venue information"
// @Success 200 {object} Venue "Venue updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid input"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden - not the venue manager"
// @Failure 404 {object} utils.ErrorResponse "Venue not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /manager/venues/{venue_id} [put]
// @Security Bearer
func (c *VenueController) UpdateVenue(ctx *gin.Context) {
	venueID, err := strconv.ParseUint(ctx.Param("venue_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid venue ID"})
		return
	}

	var input VenueInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	_, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get existing venue
	venue, err := c.repo.GetVenueByID(uint(venueID))
	if err != nil {
		if err.Error() == "venue not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "venue not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get venue: " + err.Error()})
		}
		return
	}

	// Update venue fields
	venue.Name = input.Name
	venue.Location = input.Location
	venue.Coordinates = input.Coordinates
	venue.Facilities = input.Facilities
	venue.Available = input.Available
	venue.ContactInfo = input.ContactInfo
	venue.Description = input.Description
	venue.Images = input.Images
	venue.Capacity = input.Capacity
	venue.HourlyRate = input.HourlyRate
	venue.CourtCount = input.CourtCount
	venue.SocialHours = input.SocialHours

	// Save updated venue
	if err := c.repo.UpdateVenue(venue); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to update venue: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, venue)
}

// DeleteVenue godoc
// @Summary Delete venue
// @Description Delete an existing venue and all its associated data
// @Tags venues
// @Produce json
// @Param venue_id path int true "Venue ID"
// @Success 200 {object} utils.SuccessResponse "Venue deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid venue ID"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden - not the venue manager"
// @Failure 404 {object} utils.ErrorResponse "Venue not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /manager/venues/{venue_id} [delete]
// @Security Bearer
func (c *VenueController) DeleteVenue(ctx *gin.Context) {
	venueID, err := strconv.ParseUint(ctx.Param("venue_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid venue ID"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get existing venue
	venue, err := c.repo.GetVenueByID(uint(venueID))
	if err != nil {
		if err.Error() == "venue not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "venue not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get venue: " + err.Error()})
		}
		return
	}

	// Check if the user is the venue manager
	if venue.ManagerID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "you are not authorized to delete this venue"})
		return
	}

	// Delete venue and related data
	if err := c.repo.DeleteVenue(uint(venueID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to delete venue: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessResponse{Message: "venue deleted successfully"})
}

// AddCourt godoc
// @Summary Add court to venue
// @Description Add a new court to an existing venue
// @Tags venues
// @Accept json
// @Produce json
// @Param venue_id path int true "Venue ID"
// @Param court body CourtInput true "Court information"
// @Success 201 {object} Ground "Court added successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid input"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden - not the venue manager"
// @Failure 404 {object} utils.ErrorResponse "Venue not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /manager/venues/{venue_id}/courts [post]
// @Security Bearer
func (c *VenueController) AddCourt(ctx *gin.Context) {
	venueID, err := strconv.ParseUint(ctx.Param("venue_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid venue ID"})
		return
	}

	var input CourtInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get existing venue
	venue, err := c.repo.GetVenueByID(uint(venueID))
	if err != nil {
		if err.Error() == "venue not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "venue not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get venue: " + err.Error()})
		}
		return
	}

	// Check if the user is the venue manager
	if venue.ManagerID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "you are not authorized to add courts to this venue"})
		return
	}

	// Create court object
	court := &Ground{
		VenueID:     uint(venueID),
		Name:        input.Name,
		Type:        input.Type,
		Description: input.Description,
	}

	// Save court to database
	if err := c.repo.AddCourt(court); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to add court: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, court)
}

// GetVenueCourts godoc
// @Summary Get venue courts
// @Description Get all courts for a specific venue
// @Tags venues
// @Produce json
// @Param venue_id path int true "Venue ID"
// @Success 200 {array} Ground "List of courts"
// @Failure 400 {object} utils.ErrorResponse "Invalid venue ID"
// @Failure 404 {object} utils.ErrorResponse "Venue not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /venues/{venue_id}/courts [get]
func (c *VenueController) GetVenueCourts(ctx *gin.Context) {
	venueID, err := strconv.ParseUint(ctx.Param("venue_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid venue ID"})
		return
	}

	// Verify venue exists
	_, err = c.repo.GetVenueByID(uint(venueID))
	if err != nil {
		if err.Error() == "venue not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "venue not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get venue: " + err.Error()})
		}
		return
	}

	courts, err := c.repo.GetCourtsByVenueID(uint(venueID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get courts: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, courts)
}

// UpdateCourt godoc
// @Summary Update court
// @Description Update an existing court's details
// @Tags venues
// @Accept json
// @Produce json
// @Param venue_id path int true "Venue ID"
// @Param court_id path int true "Court ID"
// @Param court body CourtInput true "Updated court information"
// @Success 200 {object} Ground "Court updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid input"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden - not the venue manager"
// @Failure 404 {object} utils.ErrorResponse "Court or venue not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /manager/venues/{venue_id}/courts/{court_id} [put]
// @Security Bearer
func (c *VenueController) UpdateCourt(ctx *gin.Context) {
	venueID, err := strconv.ParseUint(ctx.Param("venue_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid venue ID"})
		return
	}

	courtID, err := strconv.ParseUint(ctx.Param("court_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid court ID"})
		return
	}

	var input CourtInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get existing venue
	venue, err := c.repo.GetVenueByID(uint(venueID))
	if err != nil {
		if err.Error() == "venue not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "venue not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get venue: " + err.Error()})
		}
		return
	}

	// Check if the user is the venue manager
	if venue.ManagerID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "you are not authorized to update courts in this venue"})
		return
	}

	// Get existing court
	court, err := c.repo.GetCourtByID(uint(courtID))
	if err != nil {
		if err.Error() == "court not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "court not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get court: " + err.Error()})
		}
		return
	}

	// Verify court belongs to the venue
	if court.VenueID != uint(venueID) {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "court does not belong to the specified venue"})
		return
	}

	// Update court fields
	court.Name = input.Name
	court.Type = input.Type
	court.Description = input.Description

	// Save updated court
	if err := c.repo.UpdateCourt(court); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to update court: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, court)
}

// DeleteCourt godoc
// @Summary Delete court
// @Description Delete an existing court from a venue
// @Tags venues
// @Produce json
// @Param venue_id path int true "Venue ID"
// @Param court_id path int true "Court ID"
// @Success 200 {object} utils.SuccessResponse "Court deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid input or court doesn't belong to venue"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden - not the venue manager"
// @Failure 404 {object} utils.ErrorResponse "Court or venue not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /manager/venues/{venue_id}/courts/{court_id} [delete]
// @Security Bearer
func (c *VenueController) DeleteCourt(ctx *gin.Context) {
	venueID, err := strconv.ParseUint(ctx.Param("venue_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid venue ID"})
		return
	}

	courtID, err := strconv.ParseUint(ctx.Param("court_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid court ID"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get existing venue
	venue, err := c.repo.GetVenueByID(uint(venueID))
	if err != nil {
		if err.Error() == "venue not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "venue not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get venue: " + err.Error()})
		}
		return
	}

	// Check if the user is the venue manager
	if venue.ManagerID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "you are not authorized to delete courts from this venue"})
		return
	}

	// Get existing court
	court, err := c.repo.GetCourtByID(uint(courtID))
	if err != nil {
		if err.Error() == "court not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "court not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get court: " + err.Error()})
		}
		return
	}

	// Verify court belongs to the venue
	if court.VenueID != uint(venueID) {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "court does not belong to the specified venue"})
		return
	}

	// Delete court
	if err := c.repo.DeleteCourt(uint(courtID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to delete court: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessResponse{Message: "court deleted successfully"})
}

// CreateTimeSlots godoc
// @Summary Create time slots
// @Description Create one or more time slots for a venue
// @Tags venues
// @Accept json
// @Produce json
// @Param venue_id path int true "Venue ID"
// @Param timeSlots body []TimeSlotInput true "Time slot information"
// @Success 201 {array} TimeSlot "Time slots created successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid input"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden - not the venue manager"
// @Failure 404 {object} utils.ErrorResponse "Venue not found"
// @Failure 409 {object} utils.ErrorResponse "Time slot conflict"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /manager/venues/{venue_id}/timeslots [post]
// @Security Bearer
func (c *VenueController) CreateTimeSlots(ctx *gin.Context) {
	venueID, err := strconv.ParseUint(ctx.Param("venue_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid venue ID"})
		return
	}

	var inputs []TimeSlotInput
	if err := ctx.ShouldBindJSON(&inputs); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
		return
	}

	if len(inputs) == 0 {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "at least one time slot is required"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get existing venue
	venue, err := c.repo.GetVenueByID(uint(venueID))
	if err != nil {
		if err.Error() == "venue not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "venue not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get venue: " + err.Error()})
		}
		return
	}

	// Check if the user is the venue manager
	if venue.ManagerID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "you are not authorized to create time slots for this venue"})
		return
	}

	// Validate time slots
	for _, input := range inputs {
		// Check if start time is before end time
		if !input.StartTime.Before(input.EndTime) {
			ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "start time must be before end time"})
			return
		}

		// Check if court number is within venue's court count
		if input.CourtNumber > venue.CourtCount {
			ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("court number %d exceeds venue's court count of %d", input.CourtNumber, venue.CourtCount)})
			return
		}
	}

	// Create time slots
	var timeSlots []TimeSlot
	for _, input := range inputs {
		timeSlot := TimeSlot{
			VenueID:     uint(venueID),
			CourtNumber: input.CourtNumber,
			StartTime:   input.StartTime,
			EndTime:     input.EndTime,
			Price:       input.Price,
			BookingType: input.BookingType,
			Equipment:   input.Equipment,
			IsBooked:    false,
		}
		timeSlots = append(timeSlots, timeSlot)
	}

	// Save time slots to database
	err = c.repo.CreateTimeSlots(timeSlots)
	if err != nil {
		if err.Error() == "overlapping time slot exists" {
			ctx.JSON(http.StatusConflict, utils.ErrorResponse{Error: "one or more time slots overlap with existing time slots"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to create time slots: " + err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusCreated, timeSlots)
}

// GenerateAutoTimeSlots godoc
// @Summary Generate time slots automatically
// @Description Generate time slots automatically for a venue based on specified parameters
// @Tags venues
// @Accept json
// @Produce json
// @Param venue_id path int true "Venue ID"
// @Param autoSlots body AutoTimeSlotInput true "Auto time slot generation parameters"
// @Success 201 {array} TimeSlot "Time slots created successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid input"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden - not the venue manager"
// @Failure 404 {object} utils.ErrorResponse "Venue not found"
// @Failure 409 {object} utils.ErrorResponse "Time slot conflict"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /manager/venues/{venue_id}/timeslots/auto [post]
// @Security Bearer
func (c *VenueController) GenerateAutoTimeSlots(ctx *gin.Context) {
	venueID, err := strconv.ParseUint(ctx.Param("venue_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid venue ID"})
		return
	}

	var input AutoTimeSlotInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get existing venue
	venue, err := c.repo.GetVenueByID(uint(venueID))
	if err != nil {
		if err.Error() == "venue not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "venue not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get venue: " + err.Error()})
		}
		return
	}

	// Check if the user is the venue manager
	if venue.ManagerID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "you are not authorized to create time slots for this venue"})
		return
	}

	// Validate input
	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid start date format (use YYYY-MM-DD)"})
		return
	}

	endDate, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid end date format (use YYYY-MM-DD)"})
		return
	}

	if startDate.After(endDate) {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "start date must be before or equal to end date"})
		return
	}

	// Validate court numbers
	for _, courtNum := range input.CourtNumbers {
		if courtNum <= 0 || courtNum > venue.CourtCount {
			ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("court number %d is invalid (must be between 1 and %d)", courtNum, venue.CourtCount)})
			return
		}
	}

	// Parse time strings
	startTimeStr := input.StartDate + "T" + input.StartTime + ":00Z"
	dailyStartTime, err := time.Parse("2006-01-02T15:04:05Z", startTimeStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid start time format (use HH:MM)"})
		return
	}

	endTimeStr := input.StartDate + "T" + input.EndTime + ":00Z"
	dailyEndTime, err := time.Parse("2006-01-02T15:04:05Z", endTimeStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid end time format (use HH:MM)"})
		return
	}

	if !dailyStartTime.Before(dailyEndTime) {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "daily start time must be before daily end time"})
		return
	}

	// Validate days of week
	validDays := map[string]bool{
		"monday": true, "tuesday": true, "wednesday": true, "thursday": true,
		"friday": true, "saturday": true, "sunday": true,
	}
	for _, day := range input.DaysOfWeek {
		day = strings.ToLower(day)
		if _, valid := validDays[day]; !valid {
			ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid day of week: " + day})
			return
		}
	}

	// Generate time slots
	var timeSlots []TimeSlot

	// Loop through each day in the date range
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		// Check if this day of the week should be included
		dayOfWeek := strings.ToLower(d.Weekday().String())
		include := false
		for _, day := range input.DaysOfWeek {
			if strings.ToLower(day) == dayOfWeek {
				include = true
				break
			}
		}

		if !include {
			continue
		}

		// For each court
		for _, courtNum := range input.CourtNumbers {
			// Set times for this day
			currentStart := time.Date(
				d.Year(), d.Month(), d.Day(),
				dailyStartTime.Hour(), dailyStartTime.Minute(), 0, 0,
				time.UTC,
			)

			dailyEnd := time.Date(
				d.Year(), d.Month(), d.Day(),
				dailyEndTime.Hour(), dailyEndTime.Minute(), 0, 0,
				time.UTC,
			)

			// Generate slots until we reach the end time
			for currentStart.Before(dailyEnd) {
				slotEnd := currentStart.Add(time.Duration(input.Duration) * time.Minute)

				// Ensure we don't go past the daily end time
				if slotEnd.After(dailyEnd) {
					slotEnd = dailyEnd
				}

				if slotEnd.After(currentStart) {
					timeSlot := TimeSlot{
						VenueID:     uint(venueID),
						CourtNumber: courtNum,
						StartTime:   currentStart,
						EndTime:     slotEnd,
						Price:       input.Price,
						BookingType: input.BookingType,
						Equipment:   input.Equipment,
						IsBooked:    false,
					}
					timeSlots = append(timeSlots, timeSlot)
				}

				// Move to next slot
				currentStart = slotEnd
			}
		}
	}

	// Save generated time slots
	if len(timeSlots) == 0 {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "no valid time slots could be generated with the provided parameters"})
		return
	}

	err = c.repo.CreateTimeSlots(timeSlots)
	if err != nil {
		if err.Error() == "overlapping time slot exists" {
			ctx.JSON(http.StatusConflict, utils.ErrorResponse{Error: "one or more generated time slots overlap with existing time slots"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to create time slots: " + err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusCreated, timeSlots)
}

// GetVenueTimeSlots godoc
// @Summary Get venue time slots
// @Description Get time slots for a specific venue, optionally filtered by date and court number
// @Tags venues
// @Produce json
// @Param venue_id path int true "Venue ID"
// @Param date query string false "Filter by date (YYYY-MM-DD format)"
// @Param court_number query int false "Filter by court number"
// @Success 200 {array} TimeSlot "List of time slots"
// @Failure 400 {object} utils.ErrorResponse "Invalid input parameters"
// @Failure 404 {object} utils.ErrorResponse "Venue not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /venues/{venue_id}/timeslots [get]
func (c *VenueController) GetVenueTimeSlots(ctx *gin.Context) {
	venueID, err := strconv.ParseUint(ctx.Param("venue_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid venue ID"})
		return
	}

	// Verify venue exists
	_, err = c.repo.GetVenueByID(uint(venueID))
	if err != nil {
		if err.Error() == "venue not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "venue not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get venue: " + err.Error()})
		}
		return
	}

	// Parse date filter if provided
	var dateFilter time.Time
	if dateStr := ctx.Query("date"); dateStr != "" {
		dateFilter, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid date format (use YYYY-MM-DD)"})
			return
		}
	}

	// Parse court number filter if provided
	courtNumber := 0
	if courtNumberStr := ctx.Query("court_number"); courtNumberStr != "" {
		courtNumber, err = strconv.Atoi(courtNumberStr)
		if err != nil || courtNumber < 0 {
			ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid court number"})
			return
		}
	}

	// Get time slots
	timeSlots, err := c.repo.GetTimeSlotsByVenueID(uint(venueID), dateFilter, courtNumber)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get time slots: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, timeSlots)
}

// UpdateTimeSlot godoc
// @Summary Update time slot
// @Description Update an existing time slot's details
// @Tags venues
// @Accept json
// @Produce json
// @Param venue_id path int true "Venue ID"
// @Param timeslot_id path int true "Time Slot ID"
// @Param timeSlot body TimeSlotInput true "Updated time slot information"
// @Success 200 {object} TimeSlot "Time slot updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid input"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden - not the venue manager"
// @Failure 404 {object} utils.ErrorResponse "Time slot or venue not found"
// @Failure 409 {object} utils.ErrorResponse "Time slot conflict"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /manager/venues/{venue_id}/timeslots/{timeslot_id} [put]
// @Security Bearer
func (c *VenueController) UpdateTimeSlot(ctx *gin.Context) {
	venueID, err := strconv.ParseUint(ctx.Param("venue_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid venue ID"})
		return
	}

	timeSlotID, err := strconv.ParseUint(ctx.Param("timeslot_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid time slot ID"})
		return
	}

	var input TimeSlotInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get existing venue
	venue, err := c.repo.GetVenueByID(uint(venueID))
	if err != nil {
		if err.Error() == "venue not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "venue not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get venue: " + err.Error()})
		}
		return
	}

	// Check if the user is the venue manager
	if venue.ManagerID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "you are not authorized to update time slots for this venue"})
		return
	}

	// Get existing time slot
	timeSlot, err := c.repo.GetTimeSlotByID(uint(timeSlotID))
	if err != nil {
		if err.Error() == "time slot not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "time slot not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get time slot: " + err.Error()})
		}
		return
	}

	// Verify time slot belongs to the venue
	if timeSlot.VenueID != uint(venueID) {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "time slot does not belong to the specified venue"})
		return
	}

	// Check if the time slot is already booked
	if timeSlot.IsBooked {
		ctx.JSON(http.StatusConflict, utils.ErrorResponse{Error: "cannot update a time slot that is already booked"})
		return
	}

	// Check if court number is within venue's court count
	if input.CourtNumber > venue.CourtCount {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("court number %d exceeds venue's court count of %d", input.CourtNumber, venue.CourtCount)})
		return
	}

	// Check if start time is before end time
	if !input.StartTime.Before(input.EndTime) {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "start time must be before end time"})
		return
	}

	// Update time slot fields
	timeSlot.CourtNumber = input.CourtNumber
	timeSlot.StartTime = input.StartTime
	timeSlot.EndTime = input.EndTime
	timeSlot.Price = input.Price
	timeSlot.BookingType = input.BookingType
	timeSlot.Equipment = input.Equipment

	// Save updated time slot
	if err := c.repo.UpdateTimeSlot(timeSlot); err != nil {
		if err.Error() == "overlapping time slot exists" {
			ctx.JSON(http.StatusConflict, utils.ErrorResponse{Error: "the updated time slot overlaps with an existing time slot"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to update time slot: " + err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, timeSlot)
}

// DeleteTimeSlot godoc
// @Summary Delete time slot
// @Description Delete an existing time slot
// @Tags venues
// @Produce json
// @Param venue_id path int true "Venue ID"
// @Param timeslot_id path int true "Time Slot ID"
// @Success 200 {object} utils.SuccessResponse "Time slot deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid input or time slot doesn't belong to venue"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden - not the venue manager"
// @Failure 404 {object} utils.ErrorResponse "Time slot or venue not found"
// @Failure 409 {object} utils.ErrorResponse "Cannot delete a booked time slot"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /manager/venues/{venue_id}/timeslots/{timeslot_id} [delete]
// @Security Bearer
func (c *VenueController) DeleteTimeSlot(ctx *gin.Context) {
	venueID, err := strconv.ParseUint(ctx.Param("venue_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid venue ID"})
		return
	}

	timeSlotID, err := strconv.ParseUint(ctx.Param("timeslot_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid time slot ID"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get existing venue
	venue, err := c.repo.GetVenueByID(uint(venueID))
	if err != nil {
		if err.Error() == "venue not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "venue not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get venue: " + err.Error()})
		}
		return
	}

	// Check if the user is the venue manager
	if venue.ManagerID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "you are not authorized to delete time slots for this venue"})
		return
	}

	// Get existing time slot
	timeSlot, err := c.repo.GetTimeSlotByID(uint(timeSlotID))
	if err != nil {
		if err.Error() == "time slot not found" {
			ctx.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "time slot not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to get time slot: " + err.Error()})
		}
		return
	}

	// Verify time slot belongs to the venue
	if timeSlot.VenueID != uint(venueID) {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "time slot does not belong to the specified venue"})
		return
	}

	// Check if the time slot is already booked
	if timeSlot.IsBooked {
		ctx.JSON(http.StatusConflict, utils.ErrorResponse{Error: "cannot delete a time slot that is already booked"})
		return
	}

	// Delete time slot
	if err := c.repo.DeleteTimeSlot(uint(timeSlotID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to delete time slot: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessResponse{Message: "time slot deleted successfully"})
}

// UpdateBookingStatusRequest represents the request body for status updates
type UpdateBookingStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=confirmed rejected cancelled completed pending"`
}

// PaginationQuery represents common pagination parameters
type PaginationQuery struct {
	Page  int `form:"page,default=1" binding:"min=1"`
	Limit int `form:"limit,default=10" binding:"min=1,max=100"`
}

// GetVenueBookings godoc
// @Summary Get bookings for a specific venue
// @Description Retrieves all bookings for a venue with pagination and optional filters
// @Tags venues
// @Accept json
// @Produce json
// @Param venue_id path int true "Venue ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10) maximum(100)
// @Param status query string false "Filter by status (pending, confirmed, cancelled, completed, rejected)"
// @Param date query string false "Filter by date (YYYY-MM-DD format)"
// @Param court_id query int false "Filter by court ID"
// @Success 200 {object} map[string]interface{} "List of bookings and pagination metadata"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Venue not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/venue-manager/{venue_id}/bookings [get]
func (c *VenueController) GetVenueBookings(ctx *gin.Context) {
	// Parse venue ID from URL
	venueIDStr := ctx.Param("venue_id")
	venueID, err := strconv.ParseUint(venueIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid venue ID format"})
		return
	}

	// Check if venue exists
	venue, err := c.repo.GetVenueByID(uint(venueID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Venue not found"})
		return
	}

	// Get manager ID from context (assuming it was set during authentication)
	managerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	// Ensure the requester is the manager of this venue
	if venue.ManagerID != managerID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view bookings for this venue"})
		return
	}

	// Parse pagination parameters
	var pagination PaginationQuery
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
		return
	}

	// Build filters
	filters := map[string]interface{}{}

	// Status filter
	if status := ctx.Query("status"); status != "" {
		validStatuses := map[string]bool{
			"pending":   true,
			"confirmed": true,
			"cancelled": true,
			"completed": true,
			"rejected":  true,
		}
		if _, valid := validStatuses[status]; valid {
			filters["status"] = status
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status filter"})
			return
		}
	}

	// Date filter
	if dateStr := ctx.Query("date"); dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
		filters["date"] = date
	}

	// Court ID filter
	if courtIDStr := ctx.Query("court_id"); courtIDStr != "" {
		courtID, err := strconv.ParseUint(courtIDStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid court ID format"})
			return
		}
		filters["court_id"] = uint(courtID)
	}

	// Get bookings from repository
	bookings, totalCount, err := c.repo.GetBookingsByVenueID(uint(venueID), pagination.Page, pagination.Limit, filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings: " + err.Error()})
		return
	}

	// Calculate pagination metadata
	totalPages := (totalCount + int64(pagination.Limit) - 1) / int64(pagination.Limit)
	hasNextPage := int64(pagination.Page) < totalPages
	hasPrevPage := pagination.Page > 1

	ctx.JSON(http.StatusOK, gin.H{
		"bookings": bookings,
		"pagination": gin.H{
			"total":       totalCount,
			"page":        pagination.Page,
			"limit":       pagination.Limit,
			"total_pages": totalPages,
			"has_next":    hasNextPage,
			"has_prev":    hasPrevPage,
		},
	})
}

// UpdateBookingStatus godoc
// @Summary Update booking status
// @Description Updates the status of a specific booking (confirmed, rejected, cancelled, completed)
// @Tags venues
// @Accept json
// @Produce json
// @Param booking_id path int true "Booking ID"
// @Param status body UpdateBookingStatusRequest true "New status"
// @Success 200 {object} map[string]interface{} "Status updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Booking not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/venue-manager/bookings/{booking_id}/status [put]
func (c *VenueController) UpdateBookingStatus(ctx *gin.Context) {
	// Parse booking ID from URL
	bookingIDStr := ctx.Param("booking_id")
	bookingID, err := strconv.ParseUint(bookingIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID format"})
		return
	}

	// Parse request body
	var req UpdateBookingStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Get the booking to verify ownership
	booking, err := c.repo.GetBookingByID(uint(bookingID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// Get the court to get the venue
	court, err := c.repo.GetCourtByID(booking.GroundID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify venue ownership"})
		return
	}

	// Get the venue to check manager ID
	venue, err := c.repo.GetVenueByID(court.VenueID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify venue ownership"})
		return
	}

	// Get manager ID from context (assuming it was set during authentication)
	managerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	// Ensure the requester is the manager of this venue
	if venue.ManagerID != managerID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this booking"})
		return
	}

	// Check if current status allows the requested change
	if booking.Status == "cancelled" && req.Status != "cancelled" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot change status of a cancelled booking"})
		return
	}

	if booking.Status == "completed" && req.Status != "completed" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot change status of a completed booking"})
		return
	}

	// Special handling for cancellation - must use the cancel endpoint instead
	if req.Status == "cancelled" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "To cancel a booking, use the cancel booking endpoint"})
		return
	}

	// Update the booking status
	if err := c.repo.UpdateBookingStatus(uint(bookingID), req.Status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update booking status: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Booking status updated successfully",
		"status":  req.Status,
	})
}

type CreateBookingRequest struct {
	GroundID  uint      `json:"ground_id" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
	Purpose   string    `json:"purpose"`
}

// CreateBooking godoc
// @Summary Create a new booking
// @Description Creates a new booking for a specific ground/court
// @Tags bookings
// @Accept json
// @Produce json
// @Param booking body CreateBookingRequest true "Booking details"
// @Success 201 {object} map[string]interface{} "Booking created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Ground not found"
// @Failure 409 {object} map[string]interface{} "Time slot not available"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/bookings [post]
func (c *VenueController) CreateBooking(ctx *gin.Context) {
	// Parse request body
	var req CreateBookingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Validate time range
	if req.EndTime.Before(req.StartTime) || req.EndTime.Equal(req.StartTime) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "End time must be after start time"})
		return
	}

	// Ensure booking is not in the past
	if req.StartTime.Before(time.Now()) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot create bookings in the past"})
		return
	}

	// Check if the ground exists
	ground, err := c.repo.GetCourtByID(req.GroundID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Ground not found"})
		return
	}

	// Get userID from the context (set during authentication)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	// Check if the time slot is available
	timeSlots, err := c.repo.GetTimeSlotsByVenueID(ground.VenueID, req.StartTime, int(req.GroundID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check availability: " + err.Error()})
		return
	}

	// Find an exact matching time slot
	var matchingSlot *TimeSlot
	for i := range timeSlots {
		if timeSlots[i].StartTime.Equal(req.StartTime) && timeSlots[i].EndTime.Equal(req.EndTime) {
			matchingSlot = &timeSlots[i]
			break
		}
	}

	if matchingSlot == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No matching time slot found for the requested time range"})
		return
	}

	if matchingSlot.IsBooked {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Time slot is already booked"})
		return
	}

	// Create the booking
	booking := &Booking{
		GroundID:  req.GroundID,
		UserID:    userID.(uint),
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    "pending", // Default status
		Purpose:   req.Purpose,
	}

	if err := c.repo.CreateBooking(booking); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Booking created successfully",
		"booking": booking,
	})
}

// GetUserBookings godoc
// @Summary Get user's bookings
// @Description Retrieves all bookings made by the current user
// @Tags bookings
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10) maximum(100)
// @Success 200 {object} map[string]interface{} "List of user's bookings"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/bookings [get]
func (c *VenueController) GetUserBookings(ctx *gin.Context) {
	// Get user ID from context (set during authentication)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	// Parse pagination parameters
	var pagination PaginationQuery
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
		return
	}

	// Get bookings from repository
	bookings, totalCount, err := c.repo.GetBookingsByUserID(userID.(uint), pagination.Page, pagination.Limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings: " + err.Error()})
		return
	}

	// Calculate pagination metadata
	totalPages := (totalCount + int64(pagination.Limit) - 1) / int64(pagination.Limit)
	hasNextPage := int64(pagination.Page) < totalPages
	hasPrevPage := pagination.Page > 1

	ctx.JSON(http.StatusOK, gin.H{
		"bookings": bookings,
		"pagination": gin.H{
			"total":       totalCount,
			"page":        pagination.Page,
			"limit":       pagination.Limit,
			"total_pages": totalPages,
			"has_next":    hasNextPage,
			"has_prev":    hasPrevPage,
		},
	})
}

// GetBookingByID godoc
// @Summary Get booking details
// @Description Retrieves details of a specific booking
// @Tags bookings
// @Accept json
// @Produce json
// @Param booking_id path int true "Booking ID"
// @Success 200 {object} Booking "Booking details"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Booking not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/bookings/{booking_id} [get]
func (c *VenueController) GetBookingByID(ctx *gin.Context) {
	// Parse booking ID from URL
	bookingIDStr := ctx.Param("booking_id")
	bookingID, err := strconv.ParseUint(bookingIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID format"})
		return
	}

	// Get the booking
	booking, err := c.repo.GetBookingByID(uint(bookingID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// Get user ID from context (set during authentication)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	// Check if the requester is the owner of the booking
	if booking.UserID != userID.(uint) {
		// Get the court to get the venue
		court, err := c.repo.GetCourtByID(booking.GroundID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify access permission"})
			return
		}

		// Get venue to check if requester is the venue manager
		venue, err := c.repo.GetVenueByID(court.VenueID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify access permission"})
			return
		}

		// If not the venue manager either, deny access
		if venue.ManagerID != userID.(uint) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this booking"})
			return
		}
	}

	// Return the booking details
	ctx.JSON(http.StatusOK, booking)
}

// CancelBooking godoc
// @Summary Cancel a booking
// @Description Cancels a specific booking and releases the time slot
// @Tags bookings
// @Accept json
// @Produce json
// @Param booking_id path int true "Booking ID"
// @Success 200 {object} map[string]interface{} "Booking cancelled successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Booking not found"
// @Failure 409 {object} map[string]interface{} "Cannot cancel booking"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/bookings/{booking_id} [delete]
func (c *VenueController) CancelBooking(ctx *gin.Context) {
	// Parse booking ID from URL
	bookingIDStr := ctx.Param("booking_id")
	bookingID, err := strconv.ParseUint(bookingIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID format"})
		return
	}

	// Get the booking
	booking, err := c.repo.GetBookingByID(uint(bookingID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// Get user ID from context (set during authentication)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	// Check cancellation permissions
	isVenueManager := false
	if booking.UserID != userID.(uint) {
		// Check if the requester is the venue manager
		court, err := c.repo.GetCourtByID(booking.GroundID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify access permission"})
			return
		}

		venue, err := c.repo.GetVenueByID(court.VenueID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify access permission"})
			return
		}

		if venue.ManagerID != userID.(uint) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to cancel this booking"})
			return
		}
		isVenueManager = true
	}

	// Check if booking can be cancelled
	if booking.Status == "cancelled" {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Booking is already cancelled"})
		return
	}

	if booking.Status == "completed" {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Cannot cancel a completed booking"})
		return
	}

	// Check cancellation time policy (e.g., must cancel at least 24 hours before)
	// Only apply to user cancellations, not manager cancellations
	if !isVenueManager && time.Until(booking.StartTime) < 24*time.Hour {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Bookings must be cancelled at least 24 hours in advance. Current time until booking: %.1f hours",
				time.Until(booking.StartTime).Hours()),
		})
		return
	}

	// Cancel the booking
	if err := c.repo.CancelBooking(uint(bookingID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel booking: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Booking cancelled successfully",
	})
}
