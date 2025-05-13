// venue/model.go
package venue

import (
	"time"

	"github.com/DhavalSuthar-24/miow/internal/user"
	"gorm.io/gorm"
)

// Venue represents a sports facility
type Venue struct {
	gorm.Model
	Name        string    `json:"name" gorm:"unique;not null"`
	Location    string    `json:"location" gorm:"not null"`
	Coordinates string    `json:"coordinates" gorm:"type:json"`
	Facilities  string    `json:"facilities" gorm:"type:json"`
	Available   bool      `json:"available" gorm:"default:true"`
	ContactInfo string    `json:"contact_info"`
	Description string    `json:"description"`
	Images      string    `json:"images" gorm:"type:json"`
	Capacity    int       `json:"capacity"`
	HourlyRate  float64   `json:"hourly_rate"`
	CourtCount  int       `json:"court_count" gorm:"default:1"`
	SocialHours string    `json:"social_hours" gorm:"type:json"`
	ManagerID   uint      `json:"manager_id"`
	Manager     user.User `json:"-" gorm:"foreignKey:ManagerID"`
}

// Ground represents a specific playing area within a venue
type Ground struct {
	gorm.Model
	VenueID     uint   `json:"venue_id" gorm:"index"`
	Name        string `json:"name" gorm:"not null"`
	Type        string `json:"type" gorm:"not null"`
	Description string `json:"description"`
}

// VenueSchedule defines the regular availability of a venue
type VenueSchedule struct {
	gorm.Model
	VenueID     uint      `json:"venue_id" gorm:"index"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Interval    int       `json:"interval"` // in minutes
	CourtNumber int       `json:"court_number"`
	DaysOfWeek  string    `json:"days_of_week" gorm:"type:json"`
	Price       float64   `json:"price"`
	Equipment   string    `json:"equipment" gorm:"type:json"`
	TimeZone    string    `json:"time_zone" gorm:"default:'UTC'"`
}

// Booking represents a reservation for a venue
type Booking struct {
	gorm.Model
	GroundID  uint      `json:"ground_id" gorm:"index"`
	UserID    uint      `json:"user_id" gorm:"index"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Purpose   string    `json:"purpose"`
}

// TimeSlot represents available booking slots for venues
type TimeSlot struct {
	gorm.Model
	VenueID     uint      `json:"venue_id" gorm:"index"`
	CourtNumber int       `json:"court_number" gorm:"default:1"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	IsBooked    bool      `json:"is_booked" gorm:"default:false"`
	BookedBy    uint      `json:"booked_by"`
	Price       float64   `json:"price"`
	BookingType string    `json:"booking_type" gorm:"default:'match'"`
	Equipment   string    `json:"equipment" gorm:"type:json"`
}
