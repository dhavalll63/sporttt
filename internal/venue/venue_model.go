package venue

import (
	"os/user"
	"time"

	"gorm.io/gorm"
)

type Venue struct {
	gorm.Model
	Name      string `gorm:"not null;unique"`
	Location  string `gorm:"not null"`
	ManagerID uint
	Manager   user.User
	Grounds   []Ground
}

type Ground struct {
	gorm.Model
	VenueID uint
	Name    string `gorm:"not null"`
	Type    string `gorm:"not null"`

	Description string
}

type Booking struct {
	gorm.Model
	GroundID  uint
	UserID    uint
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `gorm:"type:VARCHAR(20);check:status IN ('confirmed','cancelled','pending');default:'pending'" json:"status"`
	Purpose   string    `json:"purpose"`
}
