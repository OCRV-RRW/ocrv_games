package gameDTO

import (
	"Games/internal/DTO/skillDTO"
	"Games/internal/models"
	"time"
)

type CreateGameInput struct {
	Name        string                      `json:"name" validate:"required"`
	Description string                      `json:"description" validate:"required"`
	Skills      []skillDTO.CreateSkillInput `json:"skills"`
}

type GameResponse struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Skills      []skillDTO.SkillOnlyResponse `json:"skills"`
	CreatedAt   time.Time                    `json:"created_at"`
}

type GamesResponse struct {
	Games []GameResponse `json:"games"`
}

func FilterGameRecord(game *models.Game) GameResponse {
	responseTag := []skillDTO.SkillOnlyResponse{}
	for _, tag := range game.Skills {
		responseTag = append(responseTag, skillDTO.FilterTagRecordToResponseOnly(tag))
	}

	return GameResponse{
		Name:        game.Name,
		Description: game.Description,
		Skills:      responseTag,
		CreatedAt:   *game.CreatedAt,
	}
}
