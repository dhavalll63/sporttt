package routes

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/DhavalSuthar-24/miow/config" // Import the config package
	"github.com/DhavalSuthar-24/miow/internal/auth"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()
	r.Use(cors.Default()) // allows all origins, GET/POST/PUT

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
	authGroup := api.Group("/auth") // This group will be /api/auth

	// Get the loaded configuration and database instance
	cfg := config.GetConfig()
	dbInstance := config.DB // Access the global DB instance

	// Pass dbInstance and cfg to RegisterAuthRoutes
	auth.RegisterAuthRoutes(authGroup, dbInstance, cfg)

	return r
}
