package sport

import (
	"github.com/DhavalSuthar-24/miow/config"
	mw "github.com/DhavalSuthar-24/miow/internal/middleware" // Your middleware package
	"github.com/DhavalSuthar-24/miow/pkg/rmiddleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterSportRoutes(router *gin.RouterGroup, db *gorm.DB, appConfig *config.Config, jwtSecret string) {

	sportRepo := NewSportRepository(db)
	sportController := NewSportController(sportRepo, appConfig)

	publicSports := router.Group("/sports")
	{
		publicSports.GET("", sportController.GetAllSports)                       // Get all active sports
		publicSports.GET("/:sport_id", sportController.GetSportByID)             // Get a specific sport
		publicSports.GET("/:sport_id/skills", sportController.GetSkillsForSport) // Get skills for a sport
	}

	// Authenticated routes (requires a valid token)
	// AuthMiddleware defined by user, assuming it's globally applied or applied to parent group
	// Here, we will apply it explicitly if needed, or assume parent group has it.
	// Let's assume the main router group passed to RegisterSportRoutes might already have some base middleware.
	// For clarity, we can create a subgroup for routes requiring auth.

	authenticated := router.Group("/")
	authenticated.Use(mw.AuthMiddleware(jwtSecret, db)) // Apply general authentication
	{
		// Sport management - Admin only
		adminSports := authenticated.Group("/sports")
		adminSports.Use(rmiddleware.AdminMiddleware()) // Requires "admin" role
		{
			adminSports.POST("", sportController.CreateSport)
			adminSports.PUT("/:sport_id", sportController.UpdateSport)
			adminSports.DELETE("/:sport_id", sportController.DeleteSport)
			// Admin can also view all sports including inactive ones if GetAllSports handles a special query param for admins
		}

		// Skill management - Admin only
		adminSkills := authenticated.Group("/skills")
		adminSkills.Use(rmiddleware.AdminMiddleware()) // Requires "admin" role
		{
			// Note: AddSkillToSport is nested under /sports/:sport_id/skills for better RESTful design
			// This route group is for managing skills directly by skill_id if ever needed,
			// or if you want to group all skill admin operations.
			adminSkills.PUT("/:skill_id", sportController.UpdateSkill)
			adminSkills.DELETE("/:skill_id", sportController.DeleteSkill)
		}
		// Add skill to sport (Admin only) - nested under sports
		adminSportSkills := authenticated.Group("/sports/:sport_id/skills")
		adminSportSkills.Use(rmiddleware.AdminMiddleware())
		{
			adminSportSkills.POST("", sportController.AddSkillToSport)
		}

		// User sport preferences - Authenticated users (Player, Coach, Admin)
		userSports := authenticated.Group("/users/me/sports")
		// No specific role middleware here if AuthMiddleware is enough and any authenticated user can manage their own.
		// If you need PlayerOrCoachOrAdmin, you could add:
		userSports.Use(rmiddleware.RoleMiddleware("player", "coach", "admin"))
		{
			userSports.POST("", sportController.AddUserSportPreference)
			userSports.GET("", sportController.GetUserSportPreferences)
			userSports.DELETE("/:sport_id", sportController.RemoveUserSportPreference)
		}
	}
}
