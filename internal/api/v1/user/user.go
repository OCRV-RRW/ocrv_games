package user

import (
	"Games/internal/DTO/userDTO"
	"github.com/gofiber/fiber/v2"
)

func GetMe(c *fiber.Ctx) error {
	user := c.Locals("user").(userDTO.UserResponse)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": fiber.Map{"user": user}})
}
