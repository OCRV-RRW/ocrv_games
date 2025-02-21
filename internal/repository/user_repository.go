package repository

import (
	"Games/internal/database"
	"Games/internal/models"
	"github.com/google/uuid"
	"slices"
)

type UserRepository struct {
}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(user *models.User) error {
	return GetRepositoryErrorByGormError(database.DB.Create(user).Error)
}

func (r *UserRepository) Update(user *models.User) error {
	return GetRepositoryErrorByGormError(database.DB.Save(user).Error)
}

func (r *UserRepository) CreateUserWithTransaction(user *models.User, callback func(*models.User) error) error {
	tx := database.DB.Begin()
	err := tx.Create(user).Error

	if err != nil {
		return GetRepositoryErrorByGormError(err)
	}
	err = callback(user)
	if err == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}
	return err
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	result := database.DB.Preload("Skills").First(&user, "email = ?", email)
	err := GetRepositoryErrorByGormError(result.Error)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserById(id string) (user *models.User, err error) {
	result := database.DB.Preload("Skills").First(&user, "id = ?", id)
	err = GetRepositoryErrorByGormError(result.Error)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetAll() ([]models.User, error) {
	var users []models.User
	result := database.DB.Find(&users)
	err := GetRepositoryErrorByGormError(result.Error)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) DeleteUser(id string) error {
	return GetRepositoryErrorByGormError(database.DB.Where("id = ?", id).Delete(&models.User{}).Error)
}

func (r *UserRepository) UpdateSkillScore(user *models.User, skillName string, score int) error {

	skill := models.Skill{}
	err := database.DB.Model(&skill).First(&skill, "name = ?", skillName).Error
	if err != nil {
		return GetRepositoryErrorByGormError(err)
	}

	skill_index := slices.IndexFunc(user.Skills, func(skill *models.UserSkill) bool {
		if skill.SkillName == skillName {
			return true
		}
		return false
	})

	var userSkill *models.UserSkill

	if skill_index == -1 {
		userSkill = &models.UserSkill{
			UserID:    *user.ID,
			SkillName: skillName,
			Score:     0,
		}
		err = database.DB.Create(&userSkill).Error
		if err != nil {
			return GetRepositoryErrorByGormError(err)
		}
	} else {
		userSkill = user.Skills[skill_index]
	}

	userSkill.Score = userSkill.Score + score

	return GetRepositoryErrorByGormError(database.DB.Save(&userSkill).Error)
}

func (r *UserRepository) CreateUserSkill(skill models.UserSkill) error {
	return GetRepositoryErrorByGormError(database.DB.Create(&skill).Error)
}

func (r *UserRepository) GetUserSkills(userId uuid.UUID) ([]*models.UserSkill, error) {
	var userSkills []*models.UserSkill
	err := database.DB.Find(&userSkills, "user_id = ?", userId.String()).Error
	if err != nil {
		return nil, GetRepositoryErrorByGormError(err)
	}

	return userSkills, nil
}
