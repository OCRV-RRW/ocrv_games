package repository

import (
	"Games/internal/database"
	"Games/internal/models"
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
	result := database.DB.First(&user, "email = ?", email)
	err := GetRepositoryErrorByGormError(result.Error)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserById(id string) (user *models.User, err error) {
	result := database.DB.First(&user, "id = ?", id)
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
