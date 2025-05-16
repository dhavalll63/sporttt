package sport

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SportRepository interface {
	CreateSport(sport *Sport) error       // Changed to pointer for consistency if Create modifies ID
	GetSportByID(id uint) (*Sport, error) // Changed to pointer
	GetAllSports(page, pageSize int, searchTerm string, isActive *bool) ([]Sport, int64, error)
	UpdateSport(sport *Sport) error // Changed to pointer
	DeleteSport(id uint) error
	FindSportByName(name string) (*Sport, error) // Changed to pointer

	// Skill methods
	CreateSkill(skill *Skill) error       // Changed to pointer
	GetSkillByID(id uint) (*Skill, error) // Changed to pointer
	GetSkillsBySportID(sportID uint, page, pageSize int) ([]Skill, int64, error)
	UpdateSkill(skill *Skill) error // Changed to pointer
	DeleteSkill(id uint) error
	FindSkillByNameAndSportID(name string, sportID uint) (*Skill, error) // Changed to pointer

	// UserSport methods
	AddUserSport(userSport *UserSport) error // Changed to pointer
	GetUserSports(userID uint) ([]UserSport, error)
	GetUserSportBySportID(userID, sportID uint) (*UserSport, error) // Changed to pointer
	UpdateUserSport(userSport *UserSport) error                     // Changed to pointer
	RemoveUserSport(userID, sportID uint) error
}

type sportRepository struct {
	db *gorm.DB
}

// NewSportRepository creates a new instance of SportRepository.
func NewSportRepository(db *gorm.DB) SportRepository {
	return &sportRepository{db: db}
}

// --- Sport Methods ---

func (r *sportRepository) CreateSport(sport *Sport) error {
	return r.db.Create(sport).Error
}

func (r *sportRepository) GetSportByID(id uint) (*Sport, error) {
	var sport Sport
	err := r.db.First(&sport, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Convention: return (nil, nil) if not found is not an "error" for the caller
			// Alternatively, to propagate the specific error: return nil, gorm.ErrRecordNotFound
		}
		return nil, err // Other database error
	}
	return &sport, nil
}

func (r *sportRepository) GetAllSports(page, pageSize int, searchTerm string, isActive *bool) ([]Sport, int64, error) {
	var sports []Sport
	var total int64

	query := r.db.Model(&Sport{})

	if searchTerm != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+searchTerm+"%", "%"+searchTerm+"%")
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	} else {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("name ASC").Offset(offset).Limit(pageSize).Find(&sports).Error; err != nil {
		return nil, 0, err
	}
	return sports, total, nil
}

func (r *sportRepository) UpdateSport(sport *Sport) error {
	return r.db.Session(&gorm.Session{FullSaveAssociations: false}).Updates(sport).Error
}

func (r *sportRepository) DeleteSport(id uint) error {
	return r.db.Select(clause.Associations).Delete(&Sport{}, id).Error
}

func (r *sportRepository) FindSportByName(name string) (*Sport, error) {
	var sport Sport
	err := r.db.Where("name = ?", name).First(&sport).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sport, nil
}

// --- Skill Methods ---

func (r *sportRepository) CreateSkill(skill *Skill) error {
	return r.db.Create(skill).Error
}

func (r *sportRepository) GetSkillByID(id uint) (*Skill, error) {
	var skill Skill
	err := r.db.First(&skill, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &skill, nil
}

func (r *sportRepository) GetSkillsBySportID(sportID uint, page, pageSize int) ([]Skill, int64, error) {
	var skills []Skill
	var total int64

	query := r.db.Model(&Skill{}).Where("sport_id = ?", sportID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("name ASC").Offset(offset).Limit(pageSize).Find(&skills).Error; err != nil {
		return nil, 0, err
	}
	return skills, total, nil
}

func (r *sportRepository) UpdateSkill(skill *Skill) error {
	return r.db.Save(skill).Error
}

func (r *sportRepository) DeleteSkill(id uint) error {
	return r.db.Delete(&Skill{}, id).Error
}

func (r *sportRepository) FindSkillByNameAndSportID(name string, sportID uint) (*Skill, error) {
	var skill Skill
	err := r.db.Where("name = ? AND sport_id = ?", name, sportID).First(&skill).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Corrected: return (nil,nil) for not found as per convention
		}
		return nil, err // Other DB error
	}
	return &skill, nil
}

// --- UserSport Methods ---

func (r *sportRepository) AddUserSport(userSport *UserSport) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "sport_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"position", "level", "updated_at"}),
	}).Create(userSport).Error
}

func (r *sportRepository) GetUserSports(userID uint) ([]UserSport, error) {
	var userSports []UserSport
	err := r.db.Preload("Sport").Where("user_id = ?", userID).Find(&userSports).Error
	return userSports, err
}

func (r *sportRepository) GetUserSportBySportID(userID, sportID uint) (*UserSport, error) {
	var userSport UserSport
	err := r.db.Preload("Sport").Where("user_id = ? AND sport_id = ?", userID, sportID).First(&userSport).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &userSport, nil
}

func (r *sportRepository) UpdateUserSport(userSport *UserSport) error {
	return r.db.Model(&UserSport{}).
		Where("user_id = ? AND sport_id = ?", userSport.UserID, userSport.SportID).
		Updates(map[string]interface{}{
			"position": userSport.Position,
			"level":    userSport.Level,
		}).Error
}

func (r *sportRepository) RemoveUserSport(userID, sportID uint) error {
	return r.db.Where("user_id = ? AND sport_id = ?", userID, sportID).Delete(&UserSport{}).Error
}
