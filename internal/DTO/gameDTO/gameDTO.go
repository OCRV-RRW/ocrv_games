package gameDTO

import (
	"Games/internal/DTO/tagDTO"
	"Games/internal/models"
	"time"
)

type CreateGameInput struct {
	Name        string                  `json:"name" validate:"required"`
	Description string                  `json:"description" validate:"required"`
	Tags        []tagDTO.CreateTagInput `json:"tags"`
}

type GameResponse struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Tags        []tagDTO.TagOnlyResponse `json:"tags"`
	CreatedAt   time.Time                `json:"created_at"`
}

func FilterGameRecord(game *models.Game) GameResponse {
	responseTag := []tagDTO.TagOnlyResponse{}
	for _, tag := range game.Tags {
		responseTag = append(responseTag, tagDTO.FilterTagRecordToResponseOnly(tag))
	}

	return GameResponse{
		Name:        game.Name,
		Description: game.Description,
		Tags:        responseTag,
		CreatedAt:   *game.CreatedAt,
	}
}
