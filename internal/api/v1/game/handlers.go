package game

import (
	"Games/internal/DTO/gameDTO"
	"Games/internal/DTO/skillDTO"
	"Games/internal/api"
	"Games/internal/models"
	"Games/internal/repository"
	"Games/internal/validation"
	"errors"
	"github.com/gofiber/fiber/v2"
)

// CreateGame godoc
//
// @Description	 create game
// @Tags         Game
// @Accept		 json
// @Produce		 json
// @Success		 200
// @Failure      422 {object} api.ErrorResponse
// @Failure      400 {object} api.ErrorResponse
// @Failure      409 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router		 /api/v1/games/ [post]
func CreateGame(c *fiber.Ctx) error {
	var payload *gameDTO.CreateGameInput

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.UnprocessableEntity, Message: err.Error()},
		}))
	}

	gameErrors := validation.ValidateStruct(payload)
	if gameErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse(gameErrors))
	}

	skills := []*models.Skill{}
	for _, skill := range payload.Skills {
		skills = append(skills, skillDTO.FilterSkillByCreateSkillInput(&skill))
	}
	newGame := models.Game{
		Name:        payload.Name,
		Description: payload.Description,
		Skills:      skills,
	}

	r := repository.NewGameRepository()
	err := r.Create(&newGame)

	if errors.Is(err, repository.ErrDuplicatedKey) {
		return c.Status(fiber.StatusConflict).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.IncorrectParameter, Parameter: "name", Message: "game with this name already exists"},
		}))
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: err.Error()},
		}))
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success"})
}

// GetGames godoc
//
// @Description  get games
// @Tags         Game
// @Produce		 json
// @Param        name  query     string     false  "string name"
// @Success		 200    {object} api.SuccessResponse[gameDTO.GamesResponse]
// @Failure      404    {object} api.ErrorResponse
// @Failure      500    {object} api.ErrorResponse
// @Router		 /api/v1/games/ [get]
func GetGames(c *fiber.Ctx) error {
	r := repository.NewGameRepository()

	name := c.Query("name")
	if name != "" {
		game, err := r.GetByName(name)

		if err != nil {
			if errors.Is(err, repository.ErrRecordNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(api.NewErrorResponse([]*api.Error{
					{Code: api.NotFound, Message: "game not found"},
				}))
			} else {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "error": err})
			}
		}

		return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(
			gameDTO.GamesResponse{Games: []gameDTO.GameResponse{gameDTO.FilterGameRecord(game)}}, ""))
	}

	games, err := r.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "error": err})
	}
	var gamesResponse []gameDTO.GameResponse
	for _, game := range games {
		gamesResponse = append(gamesResponse, gameDTO.FilterGameRecord(&game))
	}

	return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(gameDTO.GamesResponse{Games: gamesResponse}, ""))
}

// DeleteGame godoc
//
// @Description	 delete game by id
// @Tags         Game
// @Produce		 json
// @Param        name   path string true "Game name"
// @Success		 200
// @Failure      500    {object} api.ErrorResponse
// @Router		 /api/v1/games/ [delete]
func DeleteGame(c *fiber.Ctx) error {
	r := repository.NewGameRepository()
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
