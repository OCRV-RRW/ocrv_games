package skillDTO

import "Games/internal/models"

type CreateSkillInput struct {
	Name string `json:"name" validate:"required"`
}

type SkillOnlyResponse struct {
	Name string `json:"name" validate:"required"`
}

func FilterTagRecordToResponseOnly(tag *models.Skill) SkillOnlyResponse {
	return SkillOnlyResponse{
		Name: tag.Name,
	}
}

func FilterSkillByCreateSkillInput(createTagInput *CreateSkillInput) *models.Skill {
	return &models.Skill{
		Name: createTagInput.Name,
	}
}
