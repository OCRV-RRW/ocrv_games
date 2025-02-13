package middleware

import (
	"Games/internal/api"
	"Games/internal/config"
	"Games/internal/models"
	"Games/internal/repository"
	"Games/internal/token"
	"Games/internal/utils"
	"errors"
	"github.com/gofiber/fiber/v2"
)

func DeserializeUser(c *fiber.Ctx) error {
	access_token := utils.GetToken(c)

	if access_token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "You are not logged in"})
	}

	config, _ := config.LoadConfig(".")

	tokenClaims, err := token.ValidateToken(access_token, config.AccessTokenPublicKey)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	// Check token in redis
	//tokenRepo := token.NewAuthTokenRepository(database.RedisClient)
	//userid, err := tokenRepo.GetUserIdByTokenUuid(tokenClaims.TokenUuid)
	//if err != nil {
	//	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": "Token is invalid or session has expired"})
	//}

	userRepo := repository.NewUserRepository()
	user, err := userRepo.GetUserById(tokenClaims.UserID)

	if errors.Is(err, repository.ErrRecordNotFound) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": "the user belonging to this token no logger exists"})
	}

	c.Locals("user", user)
	c.Locals("access_token_uuid", tokenClaims.TokenUuid)

	return c.Next()
}

func AdminUser(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	if user.IsAdmin {
		return c.Next()
	}

	return c.Status(fiber.StatusForbidden).JSON(*api.NewErrorResponse([]*api.Error{
		{Code: api.Forbidden, Message: "Permission denied"},
	}))
}
