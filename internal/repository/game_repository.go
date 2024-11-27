package game_repository

import (
	"Games/internal/models"
	"gorm.io/gorm"
)

type GameRepository struct {
	db *gorm.DB
}

func (r GameRepository) Add(game *models.Game) (tx *gorm.DB) {
	return r.db.Create(game)
}

func (r GameRepository) GetAll() {

}
