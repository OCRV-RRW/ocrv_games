package user

import (
	"Games/internal/DTO"
	"Games/internal/DTO/userDTO"
	"Games/internal/repository"
	"errors"
	"github.com/gofiber/fiber/v2"
)

// GetMe godoc
//
// @Description	 get current user
// @Tags         User
// @Produce		 json
// @Success		 200 {object}  DTO.DefaultResponse[userDTO.UserResponseDTO]
// @Failure      401
// @Router		 /api/v1/users/me [get]
func GetMe(c *fiber.Ctx) error {
	user := c.Locals("user").(userDTO.UserResponse)
	return c.Status(fiber.StatusOK).JSON(DTO.DefaultResponse[userDTO.UserResponseDTO]{
		Data:   userDTO.UserResponseDTO{User: user},
		Status: "success",
	})
}

// DeleteUser godoc
//
// @Description	 delete user by id
// @Tags         User
// @Produce		 json
// @Param        id   path string true "User ID"
// @Success		 200
// @Failure      502
// @Router		 /api/v1/users/ [delete]
func DeleteUser(c *fiber.Ctx) error {
	r := repository.NewUserRepository()
	err := r.DeleteUser(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": "couldn't delete the user"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

// GetUser godoc
//
// @Description	 get user by id
// @Tags         User
// @Produce		 json
// @Param        id   path string true "User ID"
// @Success		 200 {object} DTO.DefaultResponse[[]userDTO.UserResponseDTO]
// @Failure      502
// @Router		 /api/v1/users/ [get]
func GetUser(c *fiber.Ctx) error {
	repo := repository.NewUserRepository()

	id := c.Query("id")
	if id != "" {
		user, err := repo.GetUserById(id)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "user not found"})
		}
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": "couldn't get the user"})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": fiber.Map{"user": userDTO.FilterUserRecord(user)}})
	}

	email := c.Query("email")
	if email != "" {
		user, err := repo.GetByEmail(email)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "user not found"})
		}
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": "couldn't get the user"})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": userDTO.FilterUserRecord(user)})
	}

	users, err := repo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail"})
	}

	var userRecords = make([]userDTO.UserResponseDTO, len(users))
	for i := 0; i < len(userRecords); i++ {
		userRecords[i].User = userDTO.FilterUserRecord(&users[i])
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": userRecords})
}
