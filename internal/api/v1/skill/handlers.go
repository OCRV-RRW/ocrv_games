package skill

import (
	"Games/internal/DTO"
	"Games/internal/api"
	"Games/internal/models"
	"Games/internal/repository"
	"Games/internal/validation"
	"errors"
	"github.com/gofiber/fiber/v2"
)

// CreateSkill godoc
//
// @Description	 create skill
// @Tags         Skill
// @Accept		 json
// @Produce		 json
// @Param        CreateSkillInput		body		DTO.CreateSkillInput		true   "CreateSkillInput"
// @Success		 200
// @Failure      422 {object} api.ErrorResponse
// @Failure      400 {object} api.ErrorResponse
// @Failure      409 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router		 /api/v1/skills/ [post]
func CreateSkill(c *fiber.Ctx) error {
	var payload *DTO.CreateSkillInput

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.UnprocessableEntity, Message: err.Error()},
		}))
	}

	skillErrors := validation.ValidateStruct(payload)
	if skillErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse(skillErrors))
	}

	newSkill := models.Skill{
		Name: payload.Name,
	}

	r := repository.NewSkillRepository()

	err := r.Create(&newSkill)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicatedKey) {
			return c.Status(fiber.StatusConflict).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.IncorrectParameter, Parameter: "name", Message: "skill with this name already exists"},
			}))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.ServerError, Message: err.Error()},
			}))
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success"})
}

// GetSkills godoc
//
// @Description  get skills
// @Tags         Skill
// @Produce		 json
// @Param        name  query     string     false  "string name"
// @Success		 200    {object} api.SuccessResponse[DTO.SkillsResponse]
// @Failure      404    {object} api.ErrorResponse
// @Failure      500    {object} api.ErrorResponse
// @Router		 /api/v1/skills/ [get]
func GetSkills(c *fiber.Ctx) error {
	r := repository.NewSkillRepository()

	name := c.Query("name")
	if name != "" {
		skill, err := r.GetByName(name)

		if err != nil {
			if errors.Is(err, repository.ErrRecordNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(api.NewErrorResponse([]*api.Error{
					{Code: api.NotFound, Message: "skill not found"},
				}))
			} else {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "error": err})
			}
		}

		return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(
			DTO.SkillsResponse{Skills: []DTO.SkillResponse{DTO.FilterSkillToSkillResponse(skill)}}, ""))
	}

	skills, err := r.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "error": err})
	}
	var skillResponses []DTO.SkillResponse
	for _, skill := range skills {
		skillResponses = append(skillResponses, DTO.FilterSkillToSkillResponse(&skill))
	}

	return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(
		DTO.SkillsResponse{Skills: skillResponses}, ""))
}

// DeleteSkill godoc
//
// @Description	 delete skill by id
// @Tags         Skill
// @Produce		 json
// @Param        name   path string true "Skill name"
// @Success		 200
// @Failure      500    {object} api.ErrorResponse
// @Router		 /api/v1/skills/ [delete]
func DeleteSkill(c *fiber.Ctx) error {
	r := repository.NewSkillRepository()
	err := r.Delete(c.Params("name"))
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.NotFound, Message: "not found game"},
			}))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.ServerError, Message: err.Error()},
			}))
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}
