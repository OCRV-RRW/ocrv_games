package repository

import (
	"Games/internal/database"
	"Games/internal/models"
	"gorm.io/gorm"
)

type SkillRepository struct {
	db *gorm.DB
}

func NewSkillRepository() *SkillRepository {
	return &SkillRepository{
		db: database.DB,
	}
}

func (r *SkillRepository) Create(skill *models.Skill) error {
	return GetRepositoryErrorByGormError(r.db.Create(skill).Error)
}

func (r *SkillRepository) Update(skill *models.Skill) error {
	return GetRepositoryErrorByGormError(r.db.Save(skill).Error)
}

func (r *SkillRepository) GetByName(name string) (*models.Skill, error) {
	var skill models.Skill
	result := database.DB.First(&skill, "name = ?", name)
	err := GetRepositoryErrorByGormError(result.Error)
	if err != nil {
		return nil, err
	}
	return &skill, nil
}

func (r *SkillRepository) GetAll() ([]models.Skill, error) {
	var games []models.Skill
	result := r.db.Model(&models.Skill{}).Preload("Games").Find(&games)
	err := GetRepositoryErrorByGormError(result.Error)
	if err != nil {
		return nil, err
	}
	return games, nil
}

func (r *SkillRepository) Delete(name string) error {
	return GetRepositoryErrorByGormError(r.db.Where("name = ?", name).Delete(&models.Skill{}).Error)
}
