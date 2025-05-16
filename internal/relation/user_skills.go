// internal/relations/user_skill.go
package relations

import (
	"github.com/DhavalSuthar-24/miow/internal/sport"
	"github.com/DhavalSuthar-24/miow/internal/user"
	"gorm.io/gorm"
)

type UserSkill struct {
	gorm.Model
	UserID  uint        `json:"user_id" gorm:"index"`
	SkillID uint        `json:"skill_id" gorm:"index"`
	SportID uint        `json:"sport_id" gorm:"index"`
	Level   string      `json:"level"`
	User    user.User   `gorm:"foreignKey:UserID"`
	Skill   sport.Skill `gorm:"foreignKey:SkillID"`
	Sport   sport.Sport `gorm:"foreignKey:SportID"`
}
