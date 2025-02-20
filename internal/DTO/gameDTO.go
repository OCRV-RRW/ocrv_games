package DTO

import (
	"Games/internal/models"
	"time"
)

type CreateGameInput struct {
	Name          string             `json:"name" validate:"required"`
	FriendlyName  string             `json:"friendly_name" validate:"required"`
	ReleaseSource string             `json:"release_source"`
	DebugSource   string             `json:"debug_source"`
	Config        string             `json:"config"`
	Description   string             `json:"description" validate:"required"`
	Skills        []CreateSkillInput `json:"skills"`
}

type UpdateGameInput struct {
	FriendlyName  string   `json:"friendly_name"`
	Description   string   `json:"description"`
	Skills        []string `json:"skills"`
	ReleaseSource string   `json:"release_source"`
	DebugSource   string   `json:"debug_source"`
	Config        string   `json:"config"`
}

type GameResponse struct {
	Name          string              `json:"name"`
	FriendlyName  string              `json:"friendly_name"`
	ReleaseSource string              `json:"release_source"`
	DebugSource   string              `json:"debug_source"`
	Description   string              `json:"description"`
	PreviewUrl    string              `json:"preview_url"`
	Skills        []SkillResponseOnly `json:"skills"`
	Config        string              `json:"config"`
	CreatedAt     time.Time           `json:"created_at"`
}

type GameResponseOnly struct {
	Name          string `json:"name"`
	FriendlyName  string `json:"friendly_name"`
	ReleaseSource string `json:"release_source"`
	DebugSource   string `json:"debug_source"`
	Description   string `json:"description"`
	PreviewUrl    string `json:"preview_url"`
	CreatedAt     string `json:"created_at"`
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
		Name:          game.Name,
		FriendlyName:  game.FriendlyName,
		Description:   game.Description,
		PreviewUrl:    game.PreviewUrl,
		ReleaseSource: game.ReleaseSource,
		DebugSource:   game.DebugSource,
		Skills:        responseTag,
		Config:        game.Config,
		CreatedAt:     *game.CreatedAt,
	}
}

func FilterGameToGameResponseOnly(game *models.Game) GameResponseOnly {
	return GameResponseOnly{
		Name:          game.Name,
		FriendlyName:  game.FriendlyName,
		ReleaseSource: game.ReleaseSource,
		DebugSource:   game.DebugSource,
		Description:   game.Description,
		PreviewUrl:    game.PreviewUrl,
	}
}
