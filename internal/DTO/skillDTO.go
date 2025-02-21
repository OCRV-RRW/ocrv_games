package DTO

import (
	"Games/internal/models"
)

type UpdateSkillInput struct {
	FriendlyName string `json:"friendly_name" validate:"required"`
	Description  string `json:"description" validate:"required"`
}

type CreateSkillInput struct {
	Name         string `json:"name" validate:"required"`
	FriendlyName string `json:"friendly_name" validate:"required"`
	Description  string `json:"description" validate:"required"`
}

type SkillResponseOnly struct {
	Name         string `json:"name"`
	FriendlyName string `json:"friendly_name"`
	Description  string `json:"description"`
}

type SkillResponse struct {
	Name         string             `json:"name"`
	FriendlyName string             `json:"friendly_name"`
	Description  string             `json:"description"`
	Games        []GameResponseOnly `json:"games"`
}

type SkillsResponse struct {
	Skills []SkillResponse `json:"skills"`
}

func FilterSkillToResponseOnly(skill *models.Skill) SkillResponseOnly {
	return SkillResponseOnly{
		Name:         skill.Name,
		FriendlyName: skill.FriendlyName,
		Description:  skill.Description,
	}
}

func FilterSkillToSkillResponse(skill *models.Skill) SkillResponse {
	responseGames := []GameResponseOnly{}
	for _, game := range skill.Games {
		responseGames = append(responseGames, FilterGameToGameResponseOnly(game))
	}

	return SkillResponse{
		Name:         skill.Name,
		FriendlyName: skill.FriendlyName,
		Description:  skill.Description,
		Games:        responseGames,
	}
}

func FilterCreateSkillInputToSkill(createTagInput *CreateSkillInput) *models.Skill {
	return &models.Skill{
		Name:         createTagInput.Name,
		FriendlyName: createTagInput.FriendlyName,
		Description:  createTagInput.Description,
	}
}
