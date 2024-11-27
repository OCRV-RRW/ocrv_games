package tag

import (
	"Games/internal/DTO/tagDTO"
	"Games/internal/database"
	"Games/internal/models"
	"Games/internal/validation"
	"Games/internal/validation/error_code"
	"github.com/gofiber/fiber/v2"
	"strings"
)

func CreateTag(c *fiber.Ctx) error {
	var payload *tagDTO.CreateTagInput

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ApiError{
			Errors: []*validation.ErrorResponse{
				{
					Code:    error_code.PARSE_ERROR,
					Message: err.Error(),
				}}})
	}

	errors := validation.ValidateStruct(payload)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ApiError{Errors: errors})
	}

	newTag := models.Tag{
		Name: payload.Name,
	}

	result := database.DB.Create(&newTag)
	if result.Error != nil && strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
		return c.Status(fiber.StatusConflict).JSON(validation.ApiError{
			Errors: []*validation.ErrorResponse{
				{
					Code:    error_code.ALREADY_EXIST,
					Message: "Tag with that name already exists",
				}}})
	} else if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(validation.ApiError{
			Errors: []*validation.ErrorResponse{
				{
					Code:    error_code.ERROR,
					Message: "Something bad happened",
				}}})
	}

	c.Status(fiber.StatusCreated)
	return nil
}
