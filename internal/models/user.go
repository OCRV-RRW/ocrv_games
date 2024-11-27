package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	ID        *uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name      string     `gorm:"type:varchar(100);not null"`
	Email     string     `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password  []byte     `gorm:"type:varchar(100);not null"`
	CreatedAt *time.Time `gorm:"not null;default:now()"`
}

type Game struct {
	Name        string     `gorm:"type:varchar(100);primary_key; not null;"`
	Description string     `gorm:"type:varchar(1000);"`
	Tags        []*Tag     `gorm:"many2many:tag_game;"`
	CreatedAt   *time.Time `gorm:"not null;default:now()"`
}

type Tag struct {
	Name      string     `gorm:"type:text;primary_key"`
	Games     []*Game    `gorm:"many2many:tag_game;"`
	CreatedAt *time.Time `gorm:"not null;default:now()"`
}
