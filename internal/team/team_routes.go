package team

import (
	"github.com/DhavalSuthar-24/miow/config"                 // Assuming your config package
	mw "github.com/DhavalSuthar-24/miow/internal/middleware" // Assuming your middleware package

	// "github.com/DhavalSuthar-24/miow/internal/user" // If userRepo is needed by controller
	"github.com/DhavalSuthar-24/miow/pkg/rmiddleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TeamRoutes sets up all team-related routes
func TeamRoutes(router *gin.RouterGroup, db *gorm.DB, appConfig *config.Config, jwtSecret string,
) {
	// userRepo := user.NewUserRepository(db) // If needed
	teamRepo := NewTeamRepository(db)
	teamController := NewTeamController(teamRepo, appConfig /*, userRepo*/)

	// Public team routes
	router.GET("/teams", teamController.GetAllTeams)
	router.GET("/teams/:team_id", teamController.GetTeamByID)
	router.GET("/teams/:team_id/members", teamController.GetTeamMembers) // Publicly viewable members

	// Authenticated user routes
	authRoutes := router.Group("/")
	authRoutes.Use(mw.AuthMiddleware(jwtSecret, db)) // General authentication middleware
	{
		// Team CRUD by authenticated users
		authRoutes.POST("/teams", teamController.CreateTeam)
		authRoutes.PUT("/teams/:team_id", teamController.UpdateTeam)    // Authorization within handler
		authRoutes.DELETE("/teams/:team_id", teamController.DeleteTeam) // Authorization within handler

		// User's perspective on teams
		authRoutes.GET("/users/me/teams", teamController.GetMyTeams)
		authRoutes.GET("/users/me/teams/created", teamController.GetTeamsCreatedByMe)

		// Team Membership management by team managers (creator, captain)
		// Authorization for these actions is handled within the controller methods
		authRoutes.DELETE("/teams/:team_id/members/:user_id", teamController.RemoveTeamMember)
		authRoutes.PUT("/teams/:team_id/members/:user_id/role", teamController.UpdateTeamMemberRole)
		authRoutes.POST("/teams/:team_id/leave", teamController.LeaveTeam)

		// Join Requests
		authRoutes.POST("/teams/:team_id/join-requests", teamController.RequestToJoinTeam)
		authRoutes.GET("/teams/:team_id/join-requests", teamController.GetJoinRequestsForTeam)                   // Manager access
		authRoutes.PUT("/teams/:team_id/join-requests/:request_id/:action", teamController.RespondToJoinRequest) // Manager access (action: approve/reject)
		authRoutes.GET("/users/me/join-requests", teamController.GetMyJoinRequests)
		authRoutes.DELETE("/join-requests/:request_id", teamController.CancelJoinRequest) // User cancels their own request

		// Team Invitations
		authRoutes.POST("/teams/:team_id/invitations", teamController.InviteUserToTeam)     // Manager access
		authRoutes.GET("/teams/:team_id/invitations", teamController.GetInvitationsForTeam) // Manager access
		authRoutes.GET("/users/me/invitations", teamController.GetMyTeamInvitations)
		authRoutes.PUT("/invitations/:invitation_id/:action", teamController.RespondToTeamInvitation) // User responds (action: accept/reject)
		authRoutes.DELETE("/invitations/:invitation_id", teamController.CancelTeamInvitation)         // Manager cancels their invitation

	}

	// Admin routes (example, could be a separate group with admin-specific middleware)
	adminRoutes := router.Group("/admin")
	adminRoutes.Use(mw.AuthMiddleware(jwtSecret, db)) // General auth
	adminRoutes.Use(rmiddleware.AdminMiddleware())    // Admin-specific role check middleware
	{
		adminRoutes.GET("/teams", teamController.AdminGetAllTeams)
		// Add more admin-specific team management routes here:
		// adminRoutes.PUT("/teams/:team_id", teamController.AdminUpdateTeam)
		// adminRoutes.DELETE("/teams/:team_id", teamController.AdminDeleteTeam)
		// adminRoutes.GET("/teams/:team_id/members", teamController.AdminGetTeamMembers)
	}
}
