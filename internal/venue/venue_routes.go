package venue

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/DhavalSuthar-24/miow/config"
	mw "github.com/DhavalSuthar-24/miow/internal/middleware"

	"github.com/DhavalSuthar-24/miow/pkg/rmiddleware"
)

func RequireOwnership[T any](load func(uint) (*T, error), ownerField func(*T) uint, idParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDIfc, exists := c.Get("currentUserID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}
		userID := userIDIfc.(uint)

		param := c.Param(idParam)
		id64, err := strconv.ParseUint(param, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		resID := uint(id64)

		model, err := load(resID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "resource not found"})
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "load error"})
			}
			return
		}

		if ownerField(model) != userID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: not owner"})
			return
		}

		c.Set("loadedResource", model)
		c.Next()
	}
}

func VenueSetupRoutes(r *gin.Engine, db *gorm.DB, appConfig *config.Config, jwtSecret string) {
	public := r.Group("/")
	venueController := NewVenueController(NewVenueRepository(db), appConfig)
	public.GET("/venues", venueController.GetAllVenues)
	public.GET("/venues/:venue_id", venueController.GetVenueByID)
	public.GET("/venues/:venue_id/courts", venueController.GetVenueCourts)
	public.GET("/venues/:venue_id/timeslots", venueController.GetVenueTimeSlots)

	authenticated := r.Group("/")
	authenticated.Use(mw.AuthMiddleware(jwtSecret, db))
	{
		authenticated.POST("/bookings", venueController.CreateBooking)
		authenticated.GET("/bookings", venueController.GetUserBookings)
		authenticated.GET("/bookings/:booking_id", venueController.GetBookingByID)
		authenticated.DELETE("/bookings/:booking_id", venueController.CancelBooking)
	}

	venueManager := authenticated.Group("/manager/venues")
	venueManager.Use(rmiddleware.VenueManagerhOrAdminMiddleware())
	{
		venueManager.POST("", venueController.CreateVenue)

		venueManager.PUT("/:venue_id",
			RequireOwnership(
				func(id uint) (*Venue, error) { var v Venue; return &v, db.First(&v, id).Error },
				func(v *Venue) uint { return v.ManagerID },
				"venue_id",
			),
			venueController.UpdateVenue,
		)
		venueManager.DELETE("/:venue_id",
			RequireOwnership(
				func(id uint) (*Venue, error) { var v Venue; return &v, db.First(&v, id).Error },
				func(v *Venue) uint { return v.ManagerID },
				"venue_id",
			),
			venueController.DeleteVenue,
		)

		venueManager.POST("/:venue_id/courts",
			RequireOwnership(
				func(id uint) (*Venue, error) { var v Venue; return &v, db.First(&v, id).Error },
				func(v *Venue) uint { return v.ManagerID },
				"venue_id",
			),
			venueController.AddCourt,
		)
		venueManager.PUT("/:venue_id/courts/:court_id",
			RequireOwnership(
				func(cid uint) (*Ground, error) { var g Ground; return &g, db.Preload("Venue").First(&g, cid).Error },
				func(g *Ground) uint { return g.Venue.ManagerID },
				"court_id",
			),
			venueController.UpdateCourt,
		)
		venueManager.DELETE("/:venue_id/courts/:court_id",
			RequireOwnership(
				func(cid uint) (*Ground, error) { var g Ground; return &g, db.Preload("Venue").First(&g, cid).Error },
				func(g *Ground) uint { return g.Venue.ManagerID },
				"court_id",
			),
			venueController.DeleteCourt,
		)

		venueManager.POST("/:venue_id/timeslots",
			RequireOwnership(
				func(id uint) (*Venue, error) { var v Venue; return &v, db.First(&v, id).Error },
				func(v *Venue) uint { return v.ManagerID },
				"venue_id",
			),
			venueController.CreateTimeSlots,
		)
		venueManager.POST("/:venue_id/timeslots/auto",
			RequireOwnership(
				func(id uint) (*Venue, error) { var v Venue; return &v, db.First(&v, id).Error },
				func(v *Venue) uint { return v.ManagerID },
				"venue_id",
			),
			venueController.GenerateAutoTimeSlots,
		)
		venueManager.PUT("/:venue_id/timeslots/:timeslot_id",
			RequireOwnership(
				func(id uint) (*TimeSlot, error) { var ts TimeSlot; return &ts, db.First(&ts, id).Error },
				func(ts *TimeSlot) uint { return ts.VenueID },
				"timeslot_id",
			),
			venueController.UpdateTimeSlot,
		)
		venueManager.DELETE("/:venue_id/timeslots/:timeslot_id",
			RequireOwnership(
				func(id uint) (*TimeSlot, error) { var ts TimeSlot; return &ts, db.First(&ts, id).Error },
				func(ts *TimeSlot) uint { return ts.VenueID },
				"timeslot_id",
			),
			venueController.DeleteTimeSlot,
		)

		venueManager.GET("/:venue_id/bookings", venueController.GetVenueBookings)
		venueManager.PUT("/bookings/:booking_id/status",
			RequireOwnership(
				func(id uint) (*Booking, error) { var b Booking; return &b, db.First(&b, id).Error },
				func(b *Booking) uint { return b.UserID },
				"booking_id",
			),
			venueController.UpdateBookingStatus,
		)
	}
}
