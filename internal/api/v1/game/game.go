package game

import (
	"Games/internal/DTO/gameDTO"
	"Games/internal/DTO/tagDTO"
	"Games/internal/database"
	"Games/internal/models"
	"Games/internal/validation"
	"github.com/gofiber/fiber/v2"
	"strings"
)

func CreateGame(c *fiber.Ctx) error {
	var payload *gameDTO.CreateGameInput

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "error": err})
	}

	errors := validation.ValidateStruct(payload)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "error": errors})
	}

	tags := []*models.Tag{}
	for _, tag := range payload.Tags {
		tags = append(tags, tagDTO.CreateTagInputToTag(&tag))
	}
	newGame := models.Game{
		Name:        payload.Name,
		Description: payload.Description,
		Tags:        tags,
	}

	result := database.DB.Create(&newGame)
	if result.Error != nil && strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"status":  "fail",
			"message": "Game with that name already exists",
		})
	} else if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "fail",
			"message": result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success"})
}

func GetGames(c *fiber.Ctx) error {
	var games []models.Game
	err := database.DB.Model(&models.Game{}).Preload("Tags").Find(&games).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "error": err})
	}
	var gamesResponse []gameDTO.GameResponse
	for _, game := range games {
		gamesResponse = append(gamesResponse, gameDTO.FilterGameRecord(&game))
	}

	return c.Status(fiber.StatusOK).JSON(gamesResponse)
}
