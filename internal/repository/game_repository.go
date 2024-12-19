package repository

import (
	"Games/internal/database"
	"Games/internal/models"
	"gorm.io/gorm"
)

type GameRepository struct {
	db *gorm.DB
}

func NewGameRepository() *GameRepository {
	return &GameRepository{
		db: database.DB,
	}
}

func (r *GameRepository) Create(game *models.Game) error {
	return GetRepositoryErrorByGormError(r.db.Create(game).Error)
}

func (r *GameRepository) Update(game *models.Game) error {
	return GetRepositoryErrorByGormError(r.db.Save(game).Error)
}

func (r *GameRepository) GetByName(name string) (*models.Game, error) {
	var game models.Game
	result := database.DB.First(&game, "name = ?", name)
	err := GetRepositoryErrorByGormError(result.Error)
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *GameRepository) GetAll() ([]models.Game, error) {
	var games []models.Game
	result := r.db.Model(&models.Game{}).Preload("Skills").Find(&games)
	err := GetRepositoryErrorByGormError(result.Error)
	if err != nil {
		return nil, err
	}
	return games, nil
}

func (r *GameRepository) Delete(name string) error {
	return GetRepositoryErrorByGormError(r.db.Where("name = ?", name).Delete(&models.Game{}).Error)
}
