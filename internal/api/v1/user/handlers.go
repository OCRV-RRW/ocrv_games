package user

import (
	"Games/internal/DTO/userDTO"
	"Games/internal/api"
	"Games/internal/models"
	"Games/internal/repository"
	"Games/internal/validation"
	"errors"
	"github.com/gofiber/fiber/v2"
)

// GetMe godoc
//
// @Description	 get current user
// @Tags         User
// @Produce		 json
// @Success		 200 {object}  api.SuccessResponse[userDTO.UserResponseDTO]
// @Failure      401
// @Router		 /api/v1/users/me [get]
func GetMe(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(
		userDTO.UserResponseDTO{User: userDTO.FilterUserRecord(user)}, ""))
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
	err := r.DeleteUser(c.Params("id"))
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
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

// GetUser godoc
//
// @Description	 get user by id
// @Tags         User
// @Produce		 json
// @Param        id  query     string     false  "user id"
// @Success		 200 {object} api.SuccessResponse[userDTO.UsersResponse]
// @Failure      500
// @Router		 /api/v1/users/ [get]
func GetUser(c *fiber.Ctx) error {
	repo := repository.NewUserRepository()

	id := c.Query("id")
	if id != "" {
		user, err := repo.GetUserById(id)
		if err != nil {
			if errors.Is(err, repository.ErrRecordNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(api.NewErrorResponse([]*api.Error{
					{Code: api.NotFound, Message: "user not found"},
				}))
			} else {
				return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
					{Code: api.ServerError, Message: "couldn't get user"},
				}))
			}
		}

		return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(fiber.Map{
			"users": []userDTO.UserResponse{userDTO.FilterUserRecord(user)}}, ""))
	}

	email := c.Query("email")
	if email != "" {
		user, err := repo.GetByEmail(email)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.NotFound, Message: "user not found"},
			}))
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.ServerError, Message: "couldn't get user"},
			}))
		}
		return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(fiber.Map{
			"users": []userDTO.UserResponse{userDTO.FilterUserRecord(user)}}, ""))
	}

	users, err := repo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: "couldn't get user"},
		}))
	}

	var userRecords = make([]userDTO.UserResponseDTO, len(users))
	for i := 0; i < len(userRecords); i++ {
		userRecords[i].User = userDTO.FilterUserRecord(&users[i])
	}

	return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(fiber.Map{"users": userRecords}, ""))
}

// UpdateMe godoc
//
// @Description	 update user
// @Tags         User
// @Produce		 json
// @Param       SignInInput		body		userDTO.UpdateUserInput		true   "UpdateUserInput"
// @Success		 200
// @Failure      500
// @Failure      404
// @Router		 /api/v1/users/me [patch]
func UpdateMe(c *fiber.Ctx) error {
	var payload *userDTO.UpdateUserInput
	user := c.Locals("user").(*models.User)

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.UnprocessableEntity, Message: err.Error()},
		}))
	}

	userErrors := validation.ValidateStruct(payload)
	if userErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse(userErrors))
	}

	repo := repository.NewUserRepository()

	if payload.Age != 0 {
		user.Age = payload.Age
	}
	if payload.Grade != 0 {
		user.Grade = payload.Grade
	}
	if payload.Gender == "М" || payload.Gender == "Ж" {
		user.Gender = payload.Gender
	}
	err := repo.Update(user)
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
