package user

import (
	"Games/internal/DTO"
	"Games/internal/api"
	"Games/internal/database"
	"Games/internal/models"
	"Games/internal/repository"
	"Games/internal/token"
	"Games/internal/validation"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"slices"
	"time"
)

// GetMe godoc
//
// @Description	 get current user
// @Tags         User
// @Produce		 json
// @Success		 200 {object}  api.SuccessResponse[DTO.UserResponseDTO]
// @Failure      401
// @Router		 /api/v1/users/me [get]
func GetMe(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	userSkills, err := GetUserSkillResponse(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: err.Error()},
		}))
	}

	return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(
		DTO.UserResponseDTO{User: DTO.FilterUserRecord(user, userSkills)}, ""))
}

// DeleteUser godoc
//
// @Description	 delete user by id
// @Tags         User
// @Produce		 json
// @Param        id   path string true "User ID"
// @Success		 200
// @Failure      500
// @Router		 /api/v1/users/ [delete]
func DeleteUser(c *fiber.Ctx) error {
	r := repository.NewUserRepository()
	userId := c.Params("id")
	err := r.DeleteUser(userId)

	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.NotFound, Message: "not found user"},
			}))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.ServerError, Message: err.Error()},
			}))
		}
	}

	tokenRepo := token.NewAuthTokenRepository(database.RedisClient)
	err = tokenRepo.RemoveAllUserToken(userId)
	if err != nil {
		log.Warnf("Couldn't reset user token error: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

// GetUser godoc
//
// @Description	 get user by id
// @Tags         User
// @Produce		 json
// @Param        id  query     string     false  "user id"
// @Success		 200 {object} api.SuccessResponse[DTO.UsersResponse]
// @Failure      500
// @Router		 /api/v1/users/ [get]
func GetUser(c *fiber.Ctx) error {
	repo := repository.NewUserRepository()

	id := c.Query("id")
	if id != "" {
		userResponse, err := getUserById(c, id)
		if err != nil {
			return err
		}
		return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(fiber.Map{
			"users": []*DTO.UserResponse{userResponse}}, ""))
	}

	email := c.Query("email")
	if email != "" {
		userResponse, err := getUserByEmail(c, email)
		if err != nil {
			return err
		}
		return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(fiber.Map{
			"users": []*DTO.UserResponse{userResponse}}, ""))
	}

	users, err := repo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: "couldn't get user"},
		}))
	}

	var userRecords = make([]DTO.UserResponse, len(users))
	for i := 0; i < len(userRecords); i++ {
		userSkills, err := GetUserSkillResponse(&users[i])
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.ServerError, Message: err.Error()},
			}))
		}

		userRecords[i] = DTO.FilterUserRecord(&users[i], userSkills)
	}

	return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(fiber.Map{"users": userRecords}, ""))
}

// UpdateMe godoc
//
// @Description	 update user
// @Tags         User
// @Produce		 json
// @Param        UpdateUserInput		body		DTO.UpdateUserInput		true   "UpdateUserInput"
// @Success		 200
// @Failure      500
// @Failure      404
// @Router		 /api/v1/users/me [patch]
func UpdateMe(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	return updateUser(c, user)
}

// UpdateUser godoc
//
// @Description	 update another user
// @Tags         User
// @Produce		 json
// @Param        id   path string true "User ID"
// @Param        UpdateUserInput		body		DTO.UpdateUserInput		true   "UpdateUserInput"
// @Success		 200
// @Failure      500
// @Router		 /api/v1/users/ [patch]
func UpdateUser(c *fiber.Ctx) error {
	r := repository.NewUserRepository()
	user, err := r.GetUserById(c.Params("id"))

	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.NotFound, Message: "not found user"},
			}))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.ServerError, Message: err.Error()},
			}))
		}
	}

	return updateUser(c, user)
}

// AddScore godoc
//
// @Description	 add score to user skill
// @Tags         User
// @Produce		 json
// @Param        AddScoreToSkillInput		body		DTO.AddScoreToSkillInput		true   "AddScoreToSkillInput"
// @Success		 200
// @Failure      500
// @Failure      404
// @Router		 /api/v1/users/me/skills [post]
func AddScore(c *fiber.Ctx) error {
	var payload *DTO.AddScoreToSkillInput
	log.Info("update score body is:" + string(c.Body()))
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.UnprocessableEntity, Message: err.Error()},
		}))
	}

	updateSkillErrors := validation.ValidateStruct(payload)
	if updateSkillErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse(updateSkillErrors))
	}

	r := repository.NewUserRepository()

	user := c.Locals("user").(*models.User)

	err := r.UpdateSkillScore(user, payload.SkillName, payload.Score)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.NotFound, Message: "user not found"},
			}))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: err.Error()},
		}))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

// GetUserScores godoc
//
// @Description	 get score and the name of the user's skills
// @Tags         User
// @Produce		 json
// @Success		 200 {object} api.SuccessResponse[DTO.UserSkillsResponse]
// @Failure      500
// @Router		 /api/v1/users/me/skills [get]
func GetUserScores(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	userSkills, err := GetUserSkillResponse(user)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: err.Error()},
		}))
	}

	userSkillsResponse := DTO.UserSkillsResponse{
		Skills: userSkills,
	}

	return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(userSkillsResponse, ""))
}

func updateUser(c *fiber.Ctx, user *models.User) error {
	//user := c.Locals("user").(*models.User)
	var payload *DTO.UpdateUserInput
	var data map[string]interface{}

	err := json.Unmarshal(c.Body(), &data)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.UnprocessableEntity, Message: "Unprocessable entity"},
		}))
	}

	if err = c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.UnprocessableEntity, Message: "Unprocessable entity"},
		}))
	}

	userErrors := validation.ValidateStruct(payload)
	if userErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse(userErrors))
	}

	repo := repository.NewUserRepository()

	if birthdayRow, ok := data["birthday"].(string); ok {
		birthday, _ := time.Parse(time.RFC3339, birthdayRow)
		if birthday.Day() > 1 {
			user.Birthday = &birthday
		}
	}
	if grade, ok := data["grade"].(float64); ok {
		user.Grade = int(grade)
	}
	if gender, ok := data["gender"].(string); ok && (gender == "М" || gender == "Ж") {
		user.Gender = gender
	}
	if isAdmin, ok := data["is_admin"].(bool); ok {
		user.IsAdmin = isAdmin
	}
	err = repo.Update(user)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.NotFound, Message: "user not found"},
			}))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.ServerError, Message: "server error"},
			}))
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

func GetUserSkillResponse(user *models.User) ([]DTO.UserSkillResponse, error) {
	// get userSkill
	r := repository.NewUserRepository()
	userSkills, err := r.GetUserSkills(*user.ID)
	if err != nil {
		return nil, err
	}

	// get all skills
	skillR := repository.NewSkillRepository()
	skills, err := skillR.GetAll()
	if err != nil {
		return nil, err
	}

	userSkillDTO := make([]DTO.UserSkillResponse, len(skills))

	for i := 0; i < len(skills); i++ {
		score := 0
		userSkillsIndex := slices.IndexFunc(userSkills, func(skill *models.UserSkill) bool {
			return skill.SkillName == skills[i].Name
		})
		if userSkillsIndex >= 0 {
			score = userSkills[userSkillsIndex].Score
		}

		userSkillDTO[i] = DTO.UserSkillResponse{
			Name:         skills[i].Name,
			FriendlyName: skills[i].FriendlyName,
			Description:  skills[i].Description,
			Score:        score,
		}
	}

	return userSkillDTO, nil
}

func getUserByEmail(c *fiber.Ctx, email string) (*DTO.UserResponse, error) {
	userR := repository.NewUserRepository()
	user, err := userR.GetByEmail(email)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, c.Status(fiber.StatusNotFound).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.NotFound, Message: "user not found"},
		}))
	}
	if err != nil {
		return nil, c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: "couldn't get user"},
		}))
	}

	userSkills, err := GetUserSkillResponse(user)
	if err != nil {
		return nil, c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: err.Error()},
		}))
	}
	userResponse := DTO.FilterUserRecord(user, userSkills)
	return &userResponse, nil
}

func getUserById(c *fiber.Ctx, id string) (*DTO.UserResponse, error) {
	userR := repository.NewUserRepository()
	user, err := userR.GetUserById(id)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, c.Status(fiber.StatusNotFound).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.NotFound, Message: "user not found"},
			}))
		} else {
			return nil, c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.ServerError, Message: "couldn't get user"},
			}))
		}
	}

	userSkills, err := GetUserSkillResponse(user)
	if err != nil {
		return nil, c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: err.Error()},
		}))
	}
	userResponse := DTO.FilterUserRecord(user, userSkills)
	return &userResponse, nil
}
