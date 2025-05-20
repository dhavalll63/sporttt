package team

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TeamRepository defines the interface for team data operations
type TeamRepository interface {
	// Team operations
	CreateTeam(team *Team) error
	GetTeamByID(id uint) (*Team, error)
	GetTeamByName(name string) (*Team, error)
	GetAllTeams(page, limit int, filters map[string]interface{}) ([]Team, int64, error)
	UpdateTeam(team *Team) error
	DeleteTeam(id uint, hardDelete bool) error
	GetTeamsByUserID(userID uint, page, limit int) ([]Team, int64, error) // Teams user is a member of
	GetTeamsCreatedByUserID(userID uint, page, limit int) ([]Team, int64, error)

	// TeamMember operations
	AddTeamMember(member *TeamMember) error
	GetTeamMember(teamID, userID uint) (*TeamMember, error)
	GetTeamMembers(teamID uint, page, limit int) ([]TeamMember, int64, error)
	GetTeamMembersByRole(teamID uint, role string, page, limit int) ([]TeamMember, int64, error)
	UpdateTeamMember(member *TeamMember) error
	RemoveTeamMember(teamID, userID uint) error
	IsUserTeamMember(teamID, userID uint) (bool, error)
	IsUserTeamCreator(teamID, userID uint) (bool, error)
	GetUserTeamRole(teamID, userID uint) (string, error)
	GetTeamCaptainsAndModerators(teamID uint) ([]TeamMember, error) // Includes creator, captains, vice-captains, moderators

	// TeamInvitation operations
	CreateTeamInvitation(invitation *TeamInvitation) error
	GetTeamInvitationByID(id uint) (*TeamInvitation, error)
	GetTeamInvitationsByTeamID(teamID uint, status string, page, limit int) ([]TeamInvitation, int64, error)
	GetTeamInvitationsByUserID(userID uint, status string, page, limit int) ([]TeamInvitation, int64, error)
	UpdateTeamInvitation(invitation *TeamInvitation) error
	DeleteTeamInvitation(id uint) error
	GetPendingInvitation(teamID, userID uint) (*TeamInvitation, error)

	// JoinRequest operations
	CreateJoinRequest(request *JoinRequest) error
	GetJoinRequestByID(id uint) (*JoinRequest, error)
	GetJoinRequestsByTeamID(teamID uint, status string, page, limit int) ([]JoinRequest, int64, error)
	GetJoinRequestsByUserID(userID uint, status string, page, limit int) ([]JoinRequest, int64, error)
	UpdateJoinRequest(request *JoinRequest) error
	DeleteJoinRequest(id uint) error
	GetPendingJoinRequest(teamID, userID uint) (*JoinRequest, error)
	WithTransaction(txFunc func(TeamRepository) error) error
	GetAllTeamsAdmin(page, limit int, includeDeleted bool) ([]Team, int64, error)
}

type teamRepository struct {
	db *gorm.DB
}

// NewTeamRepository creates a new instance of TeamRepository
func NewTeamRepository(db *gorm.DB) TeamRepository {
	return &teamRepository{db: db}
}

// --- Team Operations ---

func (r *teamRepository) CreateTeam(team *Team) error {
	return r.db.Create(team).Error
}

func (r *teamRepository) GetTeamByID(id uint) (*Team, error) {
	var team Team
	if err := r.db.Preload("Sport").First(&team, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) GetTeamByName(name string) (*Team, error) {
	var team Team
	if err := r.db.Preload("Sport").Where("name = ?", name).First(&team).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) GetAllTeams(page, limit int, filters map[string]interface{}) ([]Team, int64, error) {
	var teams []Team
	var total int64

	query := r.db.Model(&Team{}).Preload("Sport").Where("is_deleted = ?", false)

	if sportID, ok := filters["sport_id"]; ok {
		query = query.Where("sport_id = ?", sportID)
	}
	if level, ok := filters["level"]; ok {
		query = query.Where("level = ?", level)
	}
	if name, ok := filters["name"]; ok {
		query = query.Where("name ILIKE ?", "%"+name.(string)+"%")
	}

	query.Count(&total)
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&teams).Error; err != nil {
		return nil, 0, err
	}
	return teams, total, nil
}

func (r *teamRepository) UpdateTeam(team *Team) error {
	return r.db.Save(team).Error
}

func (r *teamRepository) DeleteTeam(id uint, hardDelete bool) error {
	if hardDelete {
		// Hard delete related records first if necessary, or rely on GORM's cascade if setup
		// For example, delete members, invitations, join requests
		if err := r.db.Where("team_id = ?", id).Delete(&TeamMember{}).Error; err != nil {
			// Log or handle error, but proceed to delete team
		}
		if err := r.db.Where("team_id = ?", id).Delete(&TeamInvitation{}).Error; err != nil {
			// Log or handle error
		}
		if err := r.db.Where("team_id = ?", id).Delete(&JoinRequest{}).Error; err != nil {
			// Log or handle error
		}
		return r.db.Unscoped().Delete(&Team{}, id).Error
	}
	return r.db.Model(&Team{}).Where("id = ?", id).Update("is_deleted", true).Error
}

func (r *teamRepository) GetTeamsByUserID(userID uint, page, limit int) ([]Team, int64, error) {
	var teams []Team
	var total int64

	query := r.db.Joins("JOIN team_members on team_members.team_id = teams.id").
		Where("team_members.user_id = ? AND team_members.is_active = ? AND teams.is_deleted = ?", userID, true, false).
		Preload("Sport")

	query.Model(&Team{}).Count(&total) // Use Model(&Team{}) for correct count on joined query

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("teams.created_at DESC").Find(&teams).Error; err != nil {
		return nil, 0, err
	}
	return teams, total, nil
}

func (r *teamRepository) GetTeamsCreatedByUserID(userID uint, page, limit int) ([]Team, int64, error) {
	var teams []Team
	var total int64
	query := r.db.Model(&Team{}).Where("created_by_id = ? AND is_deleted = ?", userID, false).Preload("Sport")
	query.Count(&total)
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&teams).Error; err != nil {
		return nil, 0, err
	}
	return teams, total, nil
}

// --- TeamMember Operations ---

func (r *teamRepository) AddTeamMember(member *TeamMember) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "team_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"role", "position", "is_active", "is_captain", "jersey_number", "stats", "updated_at"}),
	}).Create(member).Error
}

func (r *teamRepository) GetTeamMember(teamID, userID uint) (*TeamMember, error) {
	var member TeamMember
	if err := r.db.Preload("Team").Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

func (r *teamRepository) GetTeamMembers(teamID uint, page, limit int) ([]TeamMember, int64, error) {
	var members []TeamMember
	var total int64
	query := r.db.Model(&TeamMember{}).Where("team_id = ? AND is_active = ?", teamID, true) // Add Preload for User if needed
	query.Count(&total)
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at asc").Find(&members).Error; err != nil {
		return nil, 0, err
	}
	return members, total, nil
}

func (r *teamRepository) GetTeamMembersByRole(teamID uint, role string, page, limit int) ([]TeamMember, int64, error) {
	var members []TeamMember
	var total int64
	query := r.db.Model(&TeamMember{}).Where("team_id = ? AND role = ? AND is_active = ?", teamID, role, true)
	query.Count(&total)
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at asc").Find(&members).Error; err != nil {
		return nil, 0, err
	}
	return members, total, nil
}

func (r *teamRepository) UpdateTeamMember(member *TeamMember) error {
	return r.db.Save(member).Error
}

func (r *teamRepository) RemoveTeamMember(teamID, userID uint) error {
	// This could be a soft delete (setting is_active = false) or hard delete
	// return r.db.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&TeamMember{}).Error
	return r.db.Model(&TeamMember{}).Where("team_id = ? AND user_id = ?", teamID, userID).Update("is_active", false).Error
}

func (r *teamRepository) IsUserTeamMember(teamID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&TeamMember{}).Where("team_id = ? AND user_id = ? AND is_active = ?", teamID, userID, true).Count(&count).Error
	return count > 0, err
}

func (r *teamRepository) IsUserTeamCreator(teamID, userID uint) (bool, error) {
	var team Team
	if err := r.db.Select("created_by_id").First(&team, teamID).Error; err != nil {
		return false, err
	}
	return team.CreatedByID == userID, nil
}

func (r *teamRepository) GetUserTeamRole(teamID, userID uint) (string, error) {
	var member TeamMember
	if err := r.db.Select("role", "is_captain").Where("team_id = ? AND user_id = ? AND is_active = ?", teamID, userID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil // Not a member or not active
		}
		return "", err
	}
	if member.IsCaptain { // Specific IsCaptain flag takes precedence or enhances role
		return "captain", nil
	}
	return member.Role, nil
}

func (r *teamRepository) GetTeamCaptainsAndModerators(teamID uint) ([]TeamMember, error) {
	var members []TeamMember
	// Roles that can manage requests: captain, vice_captain, moderator. Creator implicitly is a captain.
	roles := []string{"captain", "vice_captain", "moderator"}
	err := r.db.Where("team_id = ? AND is_active = ? AND (role IN ? OR is_captain = ?)", teamID, true, roles, true).Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

// --- TeamInvitation Operations ---

func (r *teamRepository) CreateTeamInvitation(invitation *TeamInvitation) error {
	return r.db.Create(invitation).Error
}

func (r *teamRepository) GetTeamInvitationByID(id uint) (*TeamInvitation, error) {
	var invitation TeamInvitation
	if err := r.db.First(&invitation, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &invitation, nil
}

func (r *teamRepository) GetTeamInvitationsByTeamID(teamID uint, status string, page, limit int) ([]TeamInvitation, int64, error) {
	var invitations []TeamInvitation
	var total int64
	query := r.db.Model(&TeamInvitation{}).Where("team_id = ?", teamID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&invitations).Error; err != nil {
		return nil, 0, err
	}
	return invitations, total, nil
}

func (r *teamRepository) GetTeamInvitationsByUserID(userID uint, status string, page, limit int) ([]TeamInvitation, int64, error) {
	var invitations []TeamInvitation
	var total int64
	query := r.db.Model(&TeamInvitation{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&invitations).Error; err != nil {
		return nil, 0, err
	}
	return invitations, total, nil
}

func (r *teamRepository) UpdateTeamInvitation(invitation *TeamInvitation) error {
	return r.db.Save(invitation).Error
}

func (r *teamRepository) DeleteTeamInvitation(id uint) error {
	return r.db.Delete(&TeamInvitation{}, id).Error
}

func (r *teamRepository) GetPendingInvitation(teamID, userID uint) (*TeamInvitation, error) {
	var invitation TeamInvitation
	err := r.db.Where("team_id = ? AND user_id = ? AND status = 'pending'", teamID, userID).First(&invitation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &invitation, nil
}

// --- JoinRequest Operations ---

func (r *teamRepository) CreateJoinRequest(request *JoinRequest) error {
	return r.db.Create(request).Error
}

func (r *teamRepository) GetJoinRequestByID(id uint) (*JoinRequest, error) {
	var request JoinRequest
	if err := r.db.First(&request, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &request, nil
}

func (r *teamRepository) GetJoinRequestsByTeamID(teamID uint, status string, page, limit int) ([]JoinRequest, int64, error) {
	var requests []JoinRequest
	var total int64
	query := r.db.Model(&JoinRequest{}).Where("team_id = ?", teamID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&requests).Error; err != nil {
		return nil, 0, err
	}
	return requests, total, nil
}

func (r *teamRepository) GetJoinRequestsByUserID(userID uint, status string, page, limit int) ([]JoinRequest, int64, error) {
	var requests []JoinRequest
	var total int64
	query := r.db.Model(&JoinRequest{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&requests).Error; err != nil {
		return nil, 0, err
	}
	return requests, total, nil
}

func (r *teamRepository) UpdateJoinRequest(request *JoinRequest) error {
	return r.db.Save(request).Error
}

func (r *teamRepository) DeleteJoinRequest(id uint) error {
	return r.db.Delete(&JoinRequest{}, id).Error
}

func (r *teamRepository) GetPendingJoinRequest(teamID, userID uint) (*JoinRequest, error) {
	var request JoinRequest
	err := r.db.Where("team_id = ? AND user_id = ? AND status = 'pending'", teamID, userID).First(&request).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &request, nil
}

func (r *teamRepository) WithTransaction(txFunc func(TeamRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		txRepo := &teamRepository{db: tx}
		// Execute the function with the transactional repository
		return txFunc(txRepo)
	})
}
