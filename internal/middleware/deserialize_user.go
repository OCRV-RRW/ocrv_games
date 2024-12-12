package middleware

import (
	"Games/internal/DTO/userDTO"
	"Games/internal/config"
	"Games/internal/database"
	"Games/internal/repository"
	"Games/internal/token"
	"errors"
	"github.com/gofiber/fiber/v2"
	"strings"
)

func DeserializeUser(c *fiber.Ctx) error {
	var access_token string
	authorization := c.Get("Authorization")

	if strings.HasPrefix(authorization, "Bearer ") {
		access_token = strings.TrimPrefix(authorization, "Bearer ")
	} else if c.Cookies("access_token") != "" {
		access_token = c.Cookies("access_token")
	}

	if access_token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "You are not logged in"})
	}

	config, _ := config.LoadConfig(".")

	tokenClaims, err := token.ValidateToken(access_token, config.AccessTokenPublicKey)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	tokenRepo := token.NewAuthTokenRepository(database.RedisClient)
	userid, err := tokenRepo.GetUserIdByTokenUuid(tokenClaims.TokenUuid)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": "Token is invalid or session has expired"})
	}

	userRepo := repository.NewUserRepository()
	user, err := userRepo.GetUserById(userid)

	if errors.Is(err, repository.ErrRecordNotFound) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": "the user belonging to this token no logger exists"})
	}

	c.Locals("user", userDTO.FilterUserRecord(user))
	c.Locals("access_token_uuid", tokenClaims.TokenUuid)

	return c.Next()
}
