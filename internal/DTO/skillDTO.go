package DTO

import (
	"Games/internal/models"
)

type CreateSkillInput struct {
	Name string `json:"name" validate:"required"`
}

type SkillResponseOnly struct {
	Name string `json:"name" validate:"required"`
}

type SkillResponse struct {
	Name  string             `json:"name"`
	Games []GameResponseOnly `json:"games"`
}

type SkillsResponse struct {
	Skills []SkillResponse `json:"skills"`
}

func FilterSkillToResponseOnly(skill *models.Skill) SkillResponseOnly {
	return SkillResponseOnly{
		Name: skill.Name,
	}
}

func FilterSkillToSkillResponse(skill *models.Skill) SkillResponse {
	responseGames := []GameResponseOnly{}
	for _, game := range skill.Games {
		responseGames = append(responseGames, FilterGameToGameResponseOnly(game))
	}

	return SkillResponse{
		Name:  skill.Name,
		Games: responseGames,
	}
}

func FilterCreateSkillInputToSkill(createTagInput *CreateSkillInput) *models.Skill {
	return &models.Skill{
		Name: createTagInput.Name,
	}
}
