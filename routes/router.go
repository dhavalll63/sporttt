package routes

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"github.com/DhavalSuthar-24/miow/config" // Import the config package
	"github.com/DhavalSuthar-24/miow/internal/auth"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"}, // Where Swagger UI is hosted
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	var db *gorm.DB
	r.Use(func(c *gin.Context) {
		c.Set("auth_repo", auth.NewAuthRepository(db)) // Assuming NewAuthRepository() initializes an instance
		c.Next()
	})

	// Access config for static file serving if needed (e.g., dynamic public path)
	// cfg := config.GetConfig()
	// r.Static("/public", cfg.App.PublicDir) // Example if public dir was configurable

	r.Static("/public", "./public") // Current setup

	// Welcome page
	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
			<html>
				<head><title>Welcome</title></head>
				<body style="text-align:center; margin-top: 40px;">
					
				<h1>Welcome to MiowNation 🐱</h1>
				<div>
				<img src="/public/miow.jpeg" alt="Cat" width="300" />
					 <!-- Corrected link -->
					 <br/><a href="/swagger/index.html">View API Documentation (Swagger)</a>
				</div>
					</body>
			</html>
		`))
	})

	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := r.Group("/api")

	// Get the loaded configuration and database instance
	cfg := config.GetConfig()
	dbInstance := config.DB // Access the global DB instance

	// Pass dbInstance and cfg to RegisterAuthRoutes
	auth.RegisterAuthRoutes(api, dbInstance, cfg)

	return r
}
