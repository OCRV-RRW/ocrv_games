package utils

import (
	"github.com/gofiber/fiber/v2"
	"strings"
)

func GetToken(c *fiber.Ctx) string {
	var accessToken string
	authorization := c.Get("Authorization")

	if strings.HasPrefix(authorization, "Bearer ") {
		accessToken = strings.TrimPrefix(authorization, "Bearer ")
	} else if c.Cookies("access_token") != "" {
		accessToken = c.Cookies("access_token")
	}

	return accessToken
}
