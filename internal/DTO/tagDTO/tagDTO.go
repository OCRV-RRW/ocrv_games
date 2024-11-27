package tagDTO

import "Games/internal/models"

type CreateTagInput struct {
	Name string `json:"name" validate:"required"`
}

type TagOnlyResponse struct {
	Name string `json:"name" validate:"required"`
}

func FilterTagRecordToResponseOnly(tag *models.Tag) TagOnlyResponse {
	return TagOnlyResponse{
		Name: tag.Name,
	}
}

func CreateTagInputToTag(createTagInput *CreateTagInput) *models.Tag {
	return &models.Tag{
		Name: createTagInput.Name,
	}
}
