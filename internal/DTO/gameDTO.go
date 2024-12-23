package DTO

import (
	"Games/internal/models"
	"time"
)

type CreateGameInput struct {
	Name         string             `json:"name" validate:"required"`
	FriendlyName string             `json:"friendly_name" validate:"required"`
	Source       string             `json:"source" validate:"required"`
	Description  string             `json:"description" validate:"required"`
	Skills       []CreateSkillInput `json:"skills"`
}

type GameResponse struct {
	Name         string              `json:"name"`
	FriendlyName string              `json:"friendly_name"`
	Source       string              `json:"source"`
	Description  string              `json:"description"`
	Skills       []SkillResponseOnly `json:"skills"`
	CreatedAt    time.Time           `json:"created_at"`
}

type GameResponseOnly struct {
	Name         string `json:"name"`
	FriendlyName string `json:"friendly_name"`
	Source       string `json:"source"`
	Description  string `json:"description"`
	CreatedAt    string `json:"created_at"`
}

type GamesResponse struct {
	Games []GameResponse `json:"games"`
}

func FilterGameRecord(game *models.Game) GameResponse {
	responseTag := []SkillResponseOnly{}
	for _, tag := range game.Skills {
		responseTag = append(responseTag, FilterSkillToResponseOnly(tag))
	}

	return GameResponse{
		Name:         game.Name,
		FriendlyName: game.FriendlyName,
		Description:  game.Description,
		Source:       game.Source,
		Skills:       responseTag,
		CreatedAt:    *game.CreatedAt,
	}
}

func FilterGameToGameResponseOnly(game *models.Game) GameResponseOnly {
	return GameResponseOnly{
		Name:         game.Name,
		FriendlyName: game.FriendlyName,
		Source:       game.Source,
		Description:  game.Description,
	}
}
