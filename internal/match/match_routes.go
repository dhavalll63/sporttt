package match

import (
	"github.com/DhavalSuthar-24/miow/config"
	mw "github.com/DhavalSuthar-24/miow/internal/middleware"
	"github.com/DhavalSuthar-24/miow/internal/team"
	"github.com/DhavalSuthar-24/miow/pkg/rmiddleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MatchRoutes sets up all match-related routes.
func MatchRoutes(router *gin.RouterGroup, db *gorm.DB, appConfig *config.Config, teamRepo team.TeamRepository, jwtSecret string) {
	matchRepo := NewGormMatchRepository(db)
	matchController := NewMatchController(matchRepo, teamRepo, appConfig)

	// Authenticated routes
	authRoutes := router.Group("/matches")
	authRoutes.Use(mw.AuthMiddleware(jwtSecret, db)) // Require authentication
	{
		// Challenge routes
		authRoutes.POST("/challenges", matchController.CreateChallenge)
		authRoutes.GET("/challenges", matchController.GetChallenges)
		authRoutes.GET("/challenges/:id", matchController.GetChallengeByID)
		authRoutes.PUT("/challenges/:id", matchController.UpdateChallenge)
		authRoutes.DELETE("/challenges/:id", matchController.DeleteChallenge)
		authRoutes.GET("/challenges/user", matchController.GetUserChallenges)
		authRoutes.GET("/challenges/team/:teamId", matchController.GetTeamChallenges)
		authRoutes.POST("/challenges/:id/accept", matchController.AcceptChallenge)
		authRoutes.POST("/challenges/:id/reject", matchController.RejectChallenge)
		authRoutes.POST("/challenges/:id/cancel", matchController.CancelChallenge)

		// Match routes
		authRoutes.POST("", matchController.CreateDirectMatch)
		authRoutes.GET("", matchController.GetMatches)
		authRoutes.GET("/:id", matchController.GetMatchByID)
		authRoutes.PUT("/:id", matchController.UpdateMatch)
		authRoutes.DELETE("/:id", matchController.DeleteMatch)
		authRoutes.GET("/user", matchController.GetUserMatches)
		authRoutes.GET("/team/:teamId", matchController.GetTeamMatches)

		// Match status updates
		authRoutes.POST("/:id/start", matchController.StartMatch)
		authRoutes.POST("/:id/end", matchController.EndMatch)
		authRoutes.POST("/:id/cancel", matchController.CancelMatch)
		authRoutes.POST("/:id/postpone", matchController.PostponeMatch)

		// Match score updates
		authRoutes.POST("/:id/score", matchController.UpdateMatchScore)
	}

	// Tournament routes
	tournamentRoutes := router.Group("/tournaments")
	tournamentRoutes.Use(mw.AuthMiddleware(jwtSecret, db)) // Require authentication
	{
		tournamentRoutes.POST("", matchController.CreateTournament)
		tournamentRoutes.GET("", matchController.GetTournaments)
		tournamentRoutes.GET("/:id", matchController.GetTournamentByID)
		tournamentRoutes.PUT("/:id", matchController.UpdateTournament)
		tournamentRoutes.DELETE("/:id", matchController.DeleteTournament)
		tournamentRoutes.POST("/:id/register", matchController.RegisterTeamForTournament)
		tournamentRoutes.POST("/:id/unregister", matchController.UnregisterTeamFromTournament)
		tournamentRoutes.GET("/:id/matches", matchController.GetTournamentMatches)
	}

	// Admin match routes
	adminRoutes := router.Group("/admin/matches")
	adminRoutes.Use(mw.AuthMiddleware(jwtSecret, db))
	adminRoutes.Use(rmiddleware.AdminMiddleware())
	{
		adminRoutes.POST("/expire-challenges", matchController.ExpireChallenges)
		adminRoutes.POST("/:id/override-status", matchController.AdminOverrideMatchStatus)
		adminRoutes.POST("/:id/override-score", matchController.AdminOverrideMatchScore)
	}
}
