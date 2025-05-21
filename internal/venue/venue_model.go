package venue

import (
	"time"

	"github.com/DhavalSuthar-24/miow/internal/user"
)

type BaseModel struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

type Venue struct {
	BaseModel
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

type Ground struct {
	BaseModel
	VenueID     uint   `json:"venue_id" gorm:"index"`
	Venue       Venue  `json:"venue" gorm:"foreignKey:VenueID"`
	Name        string `json:"name" gorm:"not null"`
	Type        string `json:"type" gorm:"not null"`
	Description string `json:"description"`
}

type VenueSchedule struct {
	BaseModel
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
	BaseModel
	GroundID  uint      `json:"ground_id" gorm:"index"`
	Ground    Ground    `json:"ground" gorm:"foreignKey:GroundID"`
	UserID    uint      `json:"user_id" gorm:"index"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Purpose   string    `json:"purpose"`
}

// TimeSlot represents available booking slots for venues
type TimeSlot struct {
	BaseModel
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

// VenueInput represents the input for venue creation and update
type VenueInput struct {
	Name        string  `json:"name" binding:"required"`
	Location    string  `json:"location" binding:"required"`
	Coordinates string  `json:"coordinates"`
	Facilities  string  `json:"facilities"`
	Available   bool    `json:"available"`
	ContactInfo string  `json:"contact_info"`
	Description string  `json:"description"`
	Images      string  `json:"images"`
	Capacity    int     `json:"capacity"`
	HourlyRate  float64 `json:"hourly_rate" binding:"required,min=0"`
	CourtCount  int     `json:"court_count" binding:"required,min=1"`
	SocialHours string  `json:"social_hours"`
}

// CourtInput represents the input for court creation and update
type CourtInput struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Description string `json:"description"`
}

// TimeSlotInput represents the input for time slot creation
type TimeSlotInput struct {
	CourtNumber int       `json:"court_number" binding:"required,min=1"`
	StartTime   time.Time `json:"start_time" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime     time.Time `json:"end_time" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	Price       float64   `json:"price" binding:"required,min=0"`
	BookingType string    `json:"booking_type"`
	Equipment   string    `json:"equipment"`
}

// AutoTimeSlotInput represents the input for generating time slots automatically
type AutoTimeSlotInput struct {
	CourtNumbers []int    `json:"court_numbers" binding:"required"`
	StartDate    string   `json:"start_date" binding:"required"`
	EndDate      string   `json:"end_date" binding:"required"`
	StartTime    string   `json:"start_time" binding:"required"`
	EndTime      string   `json:"end_time" binding:"required"`
	Duration     int      `json:"duration" binding:"required,min=15"`
	Price        float64  `json:"price" binding:"required,min=0"`
	DaysOfWeek   []string `json:"days_of_week" binding:"required"`
	BookingType  string   `json:"booking_type"`
	Equipment    string   `json:"equipment"`
}

type BookingInput struct {
	GroundID  uint      `json:"ground_id" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime   time.Time `json:"end_time" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	Purpose   string    `json:"purpose"`
}

// BookingStatusInput represents the input for updating booking status
type BookingStatusInput struct {
	Status string `json:"status" binding:"required,oneof=confirmed pending cancelled rejected completed"`
}

// PaginationInput represents the input for pagination
type PaginationInput struct {
	Page  int `form:"page,default=1" binding:"min=1"`
	Limit int `form:"limit,default=10" binding:"min=1,max=100"`
}
