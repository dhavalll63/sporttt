package main

import (
	"log"
	"os"

	"github.com/DhavalSuthar-24/miow/config"
	_ "github.com/DhavalSuthar-24/miow/docs"
	"github.com/DhavalSuthar-24/miow/internal/sport"
	"github.com/DhavalSuthar-24/miow/internal/user"
	"github.com/DhavalSuthar-24/miow/internal/venue"
	"github.com/DhavalSuthar-24/miow/routes"
)

// @title MiowNation REST API(-_-)
// @version 1.0
// @description This is a sample server for demonstrating Swagger with Gin.
// @host localhost:8088
// @BasePath /api
func main() {
	config.LoadEnv()
	config.ConnectDB()
	err := config.DB.AutoMigrate(&user.User{}, &user.Role{}, &user.OTP{}, &sport.Sport{}, &sport.UserSport{}, &venue.Venue{}, &venue.Ground{}, &venue.Booking{}, &user.RefreshToken{})
	if err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}
	log.Println("AutoMigrate successful")

	r := routes.SetupRoutes()
	r.Run(":" + os.Getenv("PORT"))
}
