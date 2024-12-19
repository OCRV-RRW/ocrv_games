package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID                 *uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name               string     `gorm:"type:varchar(100);not null"`
	Email              string     `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password           []byte     `gorm:"type:varchar(100);not null"`
	VerificationCode   string     `gorm:"type:varchar(100);not null"`
	Verified           bool       `gorm:"not null"`
	Age                int        `gorm:"type:integer"`
	Gender             string     `gorm:"type:varchar(100);default A"`
	Grade              int        `gorm:"type:integer"`
	ContinuousProgress int        `gorm:"type:integer"`
	CreatedAt          *time.Time `gorm:"not null;default:now()"`
}

type Game struct {
	Name        string     `gorm:"type:varchar(100);primary_key; not null;"`
	Description string     `gorm:"type:varchar(1000);"`
	Skills      []*Skill   `gorm:"many2many:skill_game;"`
	CreatedAt   *time.Time `gorm:"not null;default:now()"`
}

type Skill struct {
	Name      string     `gorm:"type:text;primary_key"`
	Games     []*Game    `gorm:"many2many:skill_game;"`
	CreatedAt *time.Time `gorm:"not null;default:now()"`
}
