package repository

import (
	"Games/internal/database"
	"Games/internal/models"
	"errors"
	"gorm.io/gorm"
)

type UserRepository struct {
}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(user *models.User) error {
	return database.DB.Create(user).Error
}

func (r *UserRepository) Update(user *models.User) error {
	return database.DB.Save(user).Error
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	result := database.DB.First(&user, "email = ?", email)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}
	return &user, nil
}

func (r *UserRepository) GetUserById(id string) (user *models.User, err error) {
	result := database.DB.First(&user, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}
	return user, nil
}

func (r *UserRepository) GetAll() ([]models.User, error) {
	var users []models.User
	result := database.DB.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *UserRepository) DeleteUser(id string) error {
	return database.DB.Where("id = ?", id).Delete(&models.User{}).Error
}
