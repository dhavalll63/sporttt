package main

import (
	"log"

	"github.com/DhavalSuthar-24/miow/config"
	_ "github.com/DhavalSuthar-24/miow/docs"
	"github.com/DhavalSuthar-24/miow/internal/auth"
	"github.com/DhavalSuthar-24/miow/internal/sport"
	"github.com/DhavalSuthar-24/miow/internal/user"
	"github.com/DhavalSuthar-24/miow/internal/venue"
	"github.com/DhavalSuthar-24/miow/routes"
)

// @title MiowNation REST API(-_-)
// @version 1.0
// @description This is a  server for Sport_go🏏.
// @host localhost:8088
// @BasePath /api
func main() {
	if err := config.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	cfg := config.GetConfig()

	err := config.DB.AutoMigrate(
		&user.User{}, &user.Role{}, &auth.OTP{}, &user.UserRole{},
		&sport.Sport{}, &sport.UserSport{}, &sport.Skill{},
		&venue.Venue{}, &venue.Ground{}, &venue.Booking{},
		&user.RefreshToken{},
	)
	if err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}
	log.Println("AutoMigrate successful")

	r := routes.SetupRoutes()

	// Use port from loaded configuration
	log.Printf("Starting server on port %s in %s mode\n", cfg.App.Port, cfg.App.Env)
	if err := r.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
