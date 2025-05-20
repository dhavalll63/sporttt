// venue/repository.go
package venue

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// VenueRepository interface defines all database operations for venue management
type VenueRepository interface {
	// Venue operations
	CreateVenue(venue *Venue) error
	GetVenueByID(id uint) (*Venue, error)
	GetVenuesByManagerID(managerID uint) ([]Venue, error)
	GetAllVenues(page, limit int, filters map[string]interface{}) ([]Venue, int64, error)
	UpdateVenue(venue *Venue) error
	DeleteVenue(id uint) error

	// Court operations
	AddCourt(court *Ground) error
	GetCourtsByVenueID(venueID uint) ([]Ground, error)
	GetCourtByID(id uint) (*Ground, error)
	UpdateCourt(court *Ground) error
	DeleteCourt(id uint) error

	// TimeSlot operations
	CreateTimeSlot(timeSlot *TimeSlot) error
	CreateTimeSlots(timeSlots []TimeSlot) error
	GetTimeSlotsByVenueID(venueID uint, date time.Time, courtNumber int) ([]TimeSlot, error)
	GetTimeSlotByID(id uint) (*TimeSlot, error)
	UpdateTimeSlot(timeSlot *TimeSlot) error
	DeleteTimeSlot(id uint) error

	// Booking operations
	CreateBooking(booking *Booking) error
	GetBookingByID(id uint) (*Booking, error)
	GetBookingsByUserID(userID uint, page, limit int) ([]Booking, int64, error)
	GetBookingsByVenueID(venueID uint, page, limit int, filters map[string]interface{}) ([]Booking, int64, error)
	UpdateBookingStatus(id uint, status string) error
	CancelBooking(id uint) error

	// Schedule operations
	CreateVenueSchedule(schedule *VenueSchedule) error
	GetVenueSchedules(venueID uint) ([]VenueSchedule, error)
	UpdateVenueSchedule(schedule *VenueSchedule) error
	DeleteVenueSchedule(id uint) error
}

// venueRepository implements VenueRepository interface
type venueRepository struct {
	db *gorm.DB
}

// NewVenueRepository creates a new venue repository
func NewVenueRepository(db *gorm.DB) VenueRepository {
	return &venueRepository{db: db}
}

// CreateVenue adds a new venue to the database
func (r *venueRepository) CreateVenue(venue *Venue) error {
	return r.db.Create(venue).Error
}

// GetVenueByID retrieves a venue by its ID
func (r *venueRepository) GetVenueByID(id uint) (*Venue, error) {
	var venue Venue
	if err := r.db.First(&venue, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("venue not found")
		}
		return nil, err
	}
	return &venue, nil
}

// GetVenuesByManagerID retrieves all venues managed by a specific user
func (r *venueRepository) GetVenuesByManagerID(managerID uint) ([]Venue, error) {
	var venues []Venue
	if err := r.db.Where("manager_id = ?", managerID).Find(&venues).Error; err != nil {
		return nil, err
	}
	return venues, nil
}

// GetAllVenues retrieves all venues with pagination and filters
func (r *venueRepository) GetAllVenues(page, limit int, filters map[string]interface{}) ([]Venue, int64, error) {
	var venues []Venue
	var totalCount int64

	// Calculate offset
	offset := (page - 1) * limit

	// Start with the base query
	query := r.db.Model(&Venue{})

	// Apply filters if any
	for key, value := range filters {
		switch key {
		case "available":
			query = query.Where("available = ?", value)
		case "location":
			query = query.Where("location LIKE ?", "%"+value.(string)+"%")
		case "min_courts":
			query = query.Where("court_count >= ?", value)
		case "max_price":
			query = query.Where("hourly_rate <= ?", value)
		}
	}

	// Get total count for pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Find(&venues).Error; err != nil {
		return nil, 0, err
	}

	return venues, totalCount, nil
}

// UpdateVenue updates venue information
func (r *venueRepository) UpdateVenue(venue *Venue) error {
	return r.db.Save(venue).Error
}

// DeleteVenue removes a venue from the database
func (r *venueRepository) DeleteVenue(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete related time slots first
		if err := tx.Where("venue_id = ?", id).Delete(&TimeSlot{}).Error; err != nil {
			return err
		}

		// Delete related courts
		if err := tx.Where("venue_id = ?", id).Delete(&Ground{}).Error; err != nil {
			return err
		}

		// Delete related schedules
		if err := tx.Where("venue_id = ?", id).Delete(&VenueSchedule{}).Error; err != nil {
			return err
		}

		// Finally delete the venue
		if err := tx.Delete(&Venue{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}

// AddCourt adds a new court to a venue
func (r *venueRepository) AddCourt(court *Ground) error {
	return r.db.Create(court).Error
}

// GetCourtsByVenueID retrieves all courts for a specific venue
func (r *venueRepository) GetCourtsByVenueID(venueID uint) ([]Ground, error) {
	var courts []Ground
	if err := r.db.Where("venue_id = ?", venueID).Find(&courts).Error; err != nil {
		return nil, err
	}
	return courts, nil
}

// GetCourtByID retrieves a court by its ID
func (r *venueRepository) GetCourtByID(id uint) (*Ground, error) {
	var court Ground
	if err := r.db.First(&court, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("court not found")
		}
		return nil, err
	}
	return &court, nil
}

// UpdateCourt updates court information
func (r *venueRepository) UpdateCourt(court *Ground) error {
	return r.db.Save(court).Error
}

// DeleteCourt removes a court from the database
func (r *venueRepository) DeleteCourt(id uint) error {
	return r.db.Delete(&Ground{}, id).Error
}

// CreateTimeSlot adds a new time slot
func (r *venueRepository) CreateTimeSlot(timeSlot *TimeSlot) error {
	// Check if there's an overlapping time slot for the same court
	var count int64
	err := r.db.Model(&TimeSlot{}).
		Where("venue_id = ? AND court_number = ? AND ((start_time <= ? AND end_time > ?) OR (start_time < ? AND end_time >= ?) OR (start_time >= ? AND end_time <= ?))",
			timeSlot.VenueID, timeSlot.CourtNumber,
			timeSlot.StartTime, timeSlot.StartTime,
			timeSlot.EndTime, timeSlot.EndTime,
			timeSlot.StartTime, timeSlot.EndTime).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("overlapping time slot exists")
	}

	return r.db.Create(timeSlot).Error
}

// CreateTimeSlots adds multiple time slots at once
func (r *venueRepository) CreateTimeSlots(timeSlots []TimeSlot) error {
	return r.db.Create(&timeSlots).Error
}

// GetTimeSlotsByVenueID retrieves all time slots for a specific venue, optionally filtered by date and court number
func (r *venueRepository) GetTimeSlotsByVenueID(venueID uint, date time.Time, courtNumber int) ([]TimeSlot, error) {
	var timeSlots []TimeSlot
	query := r.db.Where("venue_id = ?", venueID)

	// Add date filter if provided
	if !date.IsZero() {
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)
		query = query.Where("start_time >= ? AND start_time < ?", startOfDay, endOfDay)
	}

	// Add court number filter if provided
	if courtNumber > 0 {
		query = query.Where("court_number = ?", courtNumber)
	}

	// Order by court number and start time
	query = query.Order("court_number asc, start_time asc")

	if err := query.Find(&timeSlots).Error; err != nil {
		return nil, err
	}

	return timeSlots, nil
}

// GetTimeSlotByID retrieves a time slot by its ID
func (r *venueRepository) GetTimeSlotByID(id uint) (*TimeSlot, error) {
	var timeSlot TimeSlot
	if err := r.db.First(&timeSlot, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("time slot not found")
		}
		return nil, err
	}
	return &timeSlot, nil
}

// UpdateTimeSlot updates time slot information
func (r *venueRepository) UpdateTimeSlot(timeSlot *TimeSlot) error {
	return r.db.Save(timeSlot).Error
}

// DeleteTimeSlot removes a time slot from the database
func (r *venueRepository) DeleteTimeSlot(id uint) error {
	return r.db.Delete(&TimeSlot{}, id).Error
}

// CreateBooking adds a new booking
func (r *venueRepository) CreateBooking(booking *Booking) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create the booking
		if err := tx.Create(booking).Error; err != nil {
			return err
		}

		// Update the time slot to show it's booked
		if err := tx.Model(&TimeSlot{}).
			Where("venue_id = ? AND court_number = ? AND start_time = ? AND end_time = ?",
				booking.Ground.VenueID, booking.Ground.ID, booking.StartTime, booking.EndTime).
			Updates(map[string]interface{}{
				"is_booked": true,
				"booked_by": booking.UserID,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetBookingByID retrieves a booking by its ID
func (r *venueRepository) GetBookingByID(id uint) (*Booking, error) {
	var booking Booking
	if err := r.db.Preload("Ground").First(&booking, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("booking not found")
		}
		return nil, err
	}
	return &booking, nil
}

// GetBookingsByUserID retrieves all bookings for a specific user with pagination
func (r *venueRepository) GetBookingsByUserID(userID uint, page, limit int) ([]Booking, int64, error) {
	var bookings []Booking
	var totalCount int64

	offset := (page - 1) * limit

	if err := r.db.Model(&Booking{}).Where("user_id = ?", userID).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Preload("Ground").Where("user_id = ?", userID).
		Order("start_time desc").
		Offset(offset).Limit(limit).
		Find(&bookings).Error; err != nil {
		return nil, 0, err
	}

	return bookings, totalCount, nil
}

// GetBookingsByVenueID retrieves all bookings for a specific venue with pagination and filters
func (r *venueRepository) GetBookingsByVenueID(venueID uint, page, limit int, filters map[string]interface{}) ([]Booking, int64, error) {
	var bookings []Booking
	var totalCount int64

	offset := (page - 1) * limit

	// Join with Ground to filter by venueID
	query := r.db.Model(&Booking{}).
		Joins("JOIN grounds ON bookings.ground_id = grounds.id").
		Where("grounds.venue_id = ?", venueID)

	// Apply filters
	for key, value := range filters {
		switch key {
		case "status":
			query = query.Where("bookings.status = ?", value)
		case "date":
			date, ok := value.(time.Time)
			if ok {
				startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
				endOfDay := startOfDay.Add(24 * time.Hour)
				query = query.Where("bookings.start_time >= ? AND bookings.start_time < ?", startOfDay, endOfDay)
			}
		case "court_id":
			query = query.Where("bookings.ground_id = ?", value)
		}
	}

	// Get total count
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Preload("Ground").
		Order("bookings.start_time desc").
		Offset(offset).Limit(limit).
		Find(&bookings).Error; err != nil {
		return nil, 0, err
	}

	return bookings, totalCount, nil
}

// UpdateBookingStatus updates the status of a booking
func (r *venueRepository) UpdateBookingStatus(id uint, status string) error {
	return r.db.Model(&Booking{}).Where("id = ?", id).Update("status", status).Error
}

// CancelBooking cancels a booking and releases the time slot
func (r *venueRepository) CancelBooking(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var booking Booking
		if err := tx.Preload("Ground").First(&booking, id).Error; err != nil {
			return err
		}

		// Update booking status
		if err := tx.Model(&booking).Update("status", "cancelled").Error; err != nil {
			return err
		}

		// Find the ground to get venue ID
		var ground Ground
		if err := tx.First(&ground, booking.GroundID).Error; err != nil {
			return err
		}

		// Release the time slot
		if err := tx.Model(&TimeSlot{}).
			Where("venue_id = ? AND court_number = ? AND start_time = ? AND end_time = ?",
				ground.VenueID, ground.ID, booking.StartTime, booking.EndTime).
			Updates(map[string]interface{}{
				"is_booked": false,
				"booked_by": 0,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

// CreateVenueSchedule adds a new venue schedule
func (r *venueRepository) CreateVenueSchedule(schedule *VenueSchedule) error {
	return r.db.Create(schedule).Error
}

// GetVenueSchedules retrieves all schedules for a venue
func (r *venueRepository) GetVenueSchedules(venueID uint) ([]VenueSchedule, error) {
	var schedules []VenueSchedule
	if err := r.db.Where("venue_id = ?", venueID).Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

// UpdateVenueSchedule updates a venue schedule
func (r *venueRepository) UpdateVenueSchedule(schedule *VenueSchedule) error {
	return r.db.Save(schedule).Error
}

// DeleteVenueSchedule removes a venue schedule
func (r *venueRepository) DeleteVenueSchedule(id uint) error {
	return r.db.Delete(&VenueSchedule{}, id).Error
}
