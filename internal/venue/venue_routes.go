// // venue/routes.go
package venue

// import (
// 	"github.com/gin-gonic/gin"
// 	"yourapp/pkg/middleware"
// )

// // RegisterRoutes sets up venue routes
// func RegisterRoutes(r *gin.RouterGroup) {
// 	venueRoutes := r.Group("/venues")
// 	{
// 		// Public routes
// 		venueRoutes.GET("", ListVenues)
// 		venueRoutes.GET("/:id", GetVenue)
// 		venueRoutes.GET("/:id/availability", GetVenueAvailability)

// 		// Protected routes - require authentication
// 		authorized := venueRoutes.Group("/")
// 		authorized.Use(middleware.Authenticate())
// 		{
// 			authorized.POST("", middleware.RequireRole("admin", "venue_manager"), CreateVenue)
// 			authorized.PUT("/:id", middleware.RequireRole("admin", "venue_manager"), UpdateVenue)
// 			authorized.DELETE("/:id", middleware.RequireRole("admin", "venue_manager"), DeleteVenue)

// 			// Booking routes
// 			authorized.POST("/book", CreateBooking)
// 			authorized.GET("/bookings", GetUserBookings)
// 			authorized.GET("/bookings/:id", GetBooking)
// 			authorized.PUT("/bookings/:id", UpdateBooking)
// 			authorized.DELETE("/bookings/:id", CancelBooking)

// 			// Manager-specific routes
// 			managerRoutes := authorized.Group("/")
// 			managerRoutes.Use(middleware.RequireRole("admin", "venue_manager"))
// 			{
// 				managerRoutes.GET("/:id/bookings", GetVenueBookings)
// 				managerRoutes.POST("/:id/schedule", CreateVenueSchedule)
// 				managerRoutes.PUT("/:id/schedule/:scheduleId", UpdateVenueSchedule)
// 				managerRoutes.DELETE("/:id/schedule/:scheduleId", DeleteVenueSchedule)
// 			}
// 		}
// 	}
// }
