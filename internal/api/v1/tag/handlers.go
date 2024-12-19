package tag

import (
	"Games/internal/DTO/skillDTO"
	"Games/internal/database"
	"Games/internal/models"
	"Games/internal/validation"
	"github.com/gofiber/fiber/v2"
	"strings"
)

func CreateTag(c *fiber.Ctx) error {
	var payload *skillDTO.CreateSkillInput

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "error": err})
	}

	errors := validation.ValidateStruct(payload)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "error": errors})
	}

	newTag := models.Skill{
		Name: payload.Name,
	}

	result := database.DB.Create(&newTag)
	if result.Error != nil && strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"status": "fail", "error": "Skill with that name already exists"})
	} else if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "error": result.Error.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success"})
}

func GetTags(c *fiber.Ctx) error {
	var tags []models.Skill
	err := database.DB.Model(&models.Skill{}).Preload("Games").Find(&tags).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "error": err})
	}
	var tagResponse []skillDTO.SkillOnlyResponse
	for _, tag := range tags {
		tagResponse = append(tagResponse, skillDTO.FilterTagRecordToResponseOnly(&tag))
	}

	return c.Status(fiber.StatusOK).JSON(tagResponse)
}
