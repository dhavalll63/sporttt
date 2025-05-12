package main

import (
	"os"

	"github.com/DhavalSuthar-24/miow/config"
	_ "github.com/DhavalSuthar-24/miow/docs"
	"github.com/DhavalSuthar-24/miow/internal/user"
	"github.com/DhavalSuthar-24/miow/routes"
)

// @title MiowNation REST API
// @version 1.0
// @description This is a sample server for demonstrating Swagger with Gin.
// @host localhost:8088
// @BasePath /api

func main() {
	config.LoadEnv()
	config.ConnectDB()
	config.DB.AutoMigrate(&user.User{}, &user.Role{})

	r := routes.SetupRoutes()
	r.Run(":" + os.Getenv("PORT"))
}
