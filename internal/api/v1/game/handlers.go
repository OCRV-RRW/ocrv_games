package game

import (
	"Games/internal/DTO"
	"Games/internal/api"
	"Games/internal/database"
	"Games/internal/models"
	"Games/internal/repository"
	"Games/internal/validation"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// CreateGame godoc
//
// @Description	 create game
// @Tags         Game
// @Accept		 json
// @Produce		 json
// @Param        CreateGameInput		body		DTO.CreateGameInput		true   "CreateGameInput"
// @Success		 200
// @Failure      422 {object} api.ErrorResponse
// @Failure      400 {object} api.ErrorResponse
// @Failure      409 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router		 /api/v1/games/ [post]
func CreateGame(c *fiber.Ctx) error {
	var payload *DTO.CreateGameInput

	log.Info(c.FormValue("name"))

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
		skills = append(skills, DTO.FilterCreateSkillInputToSkill(&skill))
	}

	newGame := models.Game{
		Name:          payload.Name,
		FriendlyName:  payload.FriendlyName,
		Description:   payload.Description,
		ReleaseSource: payload.ReleaseSource,
		DebugSource:   payload.DebugSource,
		Skills:        skills,
	}

	r := repository.NewGameRepository()
	err := r.Create(&newGame)

	if err != nil {
		if errors.Is(err, repository.ErrDuplicatedKey) {
			return c.Status(fiber.StatusConflict).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.IncorrectParameter, Parameter: "name", Message: "game with this name already exists"},
			}))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.ServerError, Message: err.Error()},
			}))
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success"})
}

func UploadGamePreview(c *fiber.Ctx) error {
	gameName := c.Params("name")

	gr := repository.NewGameRepository()
	game, err := gr.GetByName(gameName)

	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.NotFound, Message: "couldn't find game"},
			}))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: "error"},
		}))
	}

	fileHeader, _ := c.FormFile("preview")

	if fileHeader == nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.IncorrectParameter, Parameter: "preview", Message: "file is nil"},
		}))
	}

	objectName := fileHeader.Filename
	fileReader, _ := fileHeader.Open()
	defer fileReader.Close()
	info, err := database.PutGamePreviewer(game, objectName, fileReader)
	if err != nil {
		var incorrectParameterMessage string = "Incorrect parameter"
		if errors.Is(err, database.S3ErrorIncorrectFormat) {
			incorrectParameterMessage = "Incorrect format. Allowed format is png, jpg."
		}
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.IncorrectParameter, Parameter: "preview", Message: incorrectParameterMessage},
		}))
	}
	game.PreviewUrl = info.Location
	err = gr.Update(game)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: "Something went wrong"},
		}))
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

// GetGames godoc
//
// @Description  get games
// @Tags         Game
// @Produce		 json
// @Param        name  query     string     false  "string name"
// @Success		 200    {object} api.SuccessResponse[DTO.GamesResponse]
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
			DTO.GamesResponse{Games: []DTO.GameResponse{DTO.FilterGameRecord(game)}}, ""))
	}

	games, err := r.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "error": err})
	}
	gamesResponse := []DTO.GameResponse{}
	for _, game := range games {
		gamesResponse = append(gamesResponse, DTO.FilterGameRecord(&game))
	}

	return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(DTO.GamesResponse{Games: gamesResponse}, ""))
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

// UpdateGame godoc
//
// @Description	 update game
// @Tags         Game
// @Produce		 json
// @Param        name   path string true "Game name"
// @Param        UpdateGameInput		body		DTO.UpdateGameInput		true   "UpdateGameInput"
// @Success		 200
// @Failure      500    {object} api.ErrorResponse
// @Router		 /api/v1/games/ [patch]
func UpdateGame(c *fiber.Ctx) error {
	var payload *DTO.UpdateGameInput

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.UnprocessableEntity, Message: err.Error()},
		}))
	}

	gameErrors := validation.ValidateStruct(payload)
	if gameErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse(gameErrors))
	}

	repo := repository.NewGameRepository()

	game, err := repo.GetByName(c.Params("name"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.NotFound, Message: "game not found"},
		}))
	}

	if payload.FriendlyName != "" {
		game.FriendlyName = payload.FriendlyName
	}
	if payload.Description != "" {
		game.Description = payload.Description
	}
	if payload.ReleaseSource != "" {
		game.ReleaseSource = payload.ReleaseSource
	}
	if payload.DebugSource != "" {
		game.DebugSource = payload.DebugSource
	}
	if payload.Config != "" {
		game.Config = payload.Config
	}
	if payload.Skills != nil {
		skillRepo := repository.NewSkillRepository()
		game.Skills = []*models.Skill{}
		for _, skill := range payload.Skills {
			skill, err := skillRepo.GetByName(skill)
			if err != nil {
				continue
			}
			game.Skills = append(game.Skills, skill)
		}
	}

	err = repo.Update(game)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: "couldn't update game"},
		}))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

func TestHandler(c *fiber.Ctx) error {
	fileHeader, _ := c.FormFile("preview")
	objectName := fileHeader.Filename
	fileReader, _ := fileHeader.Open()
	defer fileReader.Close()
	database.PutObject(objectName, fileReader)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}
