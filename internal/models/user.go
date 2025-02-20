package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID                 *uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name               string      `gorm:"type:varchar(100);not null"`
	Email              string      `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password           []byte      `gorm:"type:varchar(100);not null"`
	IsAdmin            bool        `gorm:"type:boolean;default:false"`
	Skills             []UserSkill `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"` //`gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	VerificationCode   string      `gorm:"type:varchar(100);not null"`
	Verified           bool        `gorm:"not null"`
	Age                int         `gorm:"type:integer"`
	Gender             string      `gorm:"type:varchar(100);default A"`
	Grade              int         `gorm:"type:integer"`
	ContinuousProgress int         `gorm:"type:integer"`
	CreatedAt          *time.Time  `gorm:"not null;default:now()"`
}

func (User) TableName() string {
	return "person"
}

type Game struct {
	Name          string     `gorm:"type:varchar(100);primary_key; not null;"`
	FriendlyName  string     `gorm:"type:varchar(100);not null;"`
	ReleaseSource string     `gorm:"type:varchar(100);not null;"`
	DebugSource   string     `gorm:"type:varchar(100);not null;"`
	PreviewUrl    string     `gorm:"type:varchar(300);"`
	Description   string     `gorm:"type:varchar(1000);"`
	Skills        []*Skill   `gorm:"many2many:skill_game;constraint:OnDelete:CASCADE;"`
	Config        string     `gorm:"type:text"`
	CreatedAt     *time.Time `gorm:"not null;default:now()"`
}

func (Game) TableName() string {
	return "game"
}

type Skill struct {
	Name         string      `gorm:"type:varchar(100);primary_key"`
	FriendlyName string      `gorm:"type:varchar(100);not null;"`
	Description  string      `gorm:"type:varchar(1000);"`
	Games        []*Game     `gorm:"many2many:skill_game;constraint:OnDelete:CASCADE;"`
	Users        []UserSkill `gorm:"foreignKey:SkillName;constraint:OnDelete:CASCADE;"`
	CreatedAt    *time.Time  `gorm:"not null;default:now()"`
}

func (Skill) TableName() string {
	return "skill"
}

type UserSkill struct {
	UserID    uuid.UUID  `gorm:"type:uuid;primary_key;foreignKey:ID;constraint:OnDelete:CASCADE"`
	SkillName string     `gorm:"type:varchar(100);primary_key;foreignKey:name;constraint:OnDelete:CASCADE"`
	Score     int        `gorm:"type:numeric;"`
	CreatedAt *time.Time `gorm:"not null;default:now()"`
}

func (UserSkill) TableName() string {
	return "person_skill"
}
