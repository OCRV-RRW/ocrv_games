package middleware

import (
	"Games/internal/DTO/userDTO"
	"Games/internal/config"
	"Games/internal/database"
	"Games/internal/models"
	"Games/internal/utils"
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"strings"
)

func DeserializeUser(c *fiber.Ctx) error {
	//var tokenString string
	//authorization := c.Get("Authorization")
	//
	//if strings.HasPrefix(authorization, "Bearer ") {
	//	tokenString = strings.TrimPrefix(authorization, "Bearer ")
	//} else if c.Cookies("token") != "" {
	//	tokenString = c.Cookies("token")
	//}
	//
	//if tokenString == "" {
	//	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "You are not logged in"})
	//}
	//
	//config, _ := config.LoadConfig(".")
	//
	//tokenByte, err := jwt.Parse(tokenString, func(jwtToken *jwt.Token) (interface{}, error) {
	//	if _, ok := jwtToken.Method.(*jwt.SigningMethodHMAC); !ok {
	//		return nil, fmt.Errorf("unexpected signing method: %s", jwtToken.Header["alg"])
	//	}
	//
	//	return []byte(config.JwtSecret), nil
	//})
	//if err != nil {
	//	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": fmt.Sprintf("invalidate token: %v", err)})
	//}
	//
	//claims, ok := tokenByte.Claims.(jwt.MapClaims)
	//if !ok || !tokenByte.Valid {
	//	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "invalid token claim"})
	//}
	//
	//var user models.User
	//database.DB.First(&user, "id = ?", fmt.Sprint(claims["sub"]))
	//
	//if user.ID.String() != claims["sub"] {
	//	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": "the user belonging to this token no logger exists"})
	//}
	//
	//c.Locals("user", userDTO.FilterUserRecord(&user))
	//

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

	tokenClaims, err := utils.ValidateToken(access_token, config.AccessTokenPublicKey)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	ctx := context.TODO()
	userid, err := database.RedisClient.Get(ctx, tokenClaims.TokenUuid).Result()
	if err == redis.Nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": "Token is invalid or session has expired"})
	}

	var user models.User
	err = database.DB.First(&user, "id = ?", userid).Error

	if err == gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": "the user belonging to this token no logger exists"})
	}

	c.Locals("user", userDTO.FilterUserRecord(&user))
	c.Locals("access_token_uuid", tokenClaims.TokenUuid)

	return c.Next()
}
