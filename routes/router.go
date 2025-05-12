package routes

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/DhavalSuthar-24/miow/internal/auth"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()
	r.Use(cors.Default()) // allows all origins, GET/POST/PUT

	r.Static("/public", "./public")

	// Welcome page
	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
			<html>
				<head><title>Welcome</title></head>
				<body style="text-align:center; margin-top: 40px;">
					
				<h1>Welcome to MiowNation 🐱</h1>
				<div>
				<img src="/public/miow.jpeg" alt="Cat" width="300" />
					// <a src="http://localhost:8088/swagger/index.html">swagger</a>
				</div>
					</body>
			</html>
		`))
	})

	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := r.Group("/api")
	authGroup := api.Group("/auth")
	auth.RegisterAuthRoutes(authGroup)

	return r
}
