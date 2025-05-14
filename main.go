package main

import (
	"log"
	// "os" // No longer needed for PORT here if using config.GetConfig().App.Port

	"github.com/DhavalSuthar-24/miow/config"
	_ "github.com/DhavalSuthar-24/miow/docs" // For Swagger docs
	"github.com/DhavalSuthar-24/miow/internal/auth"
	"github.com/DhavalSuthar-24/miow/internal/sport"
	"github.com/DhavalSuthar-24/miow/internal/user"
	"github.com/DhavalSuthar-24/miow/internal/venue"
	"github.com/DhavalSuthar-24/miow/routes"
)

// @title MiowNation REST API(-_-)
// @version 1.0
// @description This is a sample server for demonstrating Swagger with Gin.
// @host localhost:8088  // Consider making this dynamic or updating based on config
// @BasePath /api
func main() {
	// Initialize configuration and database connection
	if err := config.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Now you can access the global config.DB or config.GetConfig()
	cfg := config.GetConfig() // Get the loaded configuration

	// AutoMigrate using the global config.DB
	err := config.DB.AutoMigrate(
		&user.User{}, &user.Role{}, &auth.OTP{},
		&sport.Sport{}, &sport.UserSport{},
		&venue.Venue{}, &venue.Ground{}, &venue.Booking{},
		&user.RefreshToken{},
	)
	if err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}
	log.Println("AutoMigrate successful")

	r := routes.SetupRoutes() // SetupRoutes will now internally call config.GetConfig()

	// Use port from loaded configuration
	log.Printf("Starting server on port %s in %s mode\n", cfg.App.Port, cfg.App.Env)
	if err := r.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
