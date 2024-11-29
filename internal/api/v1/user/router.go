package user

import (
	"Games/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func AddRoutes(router fiber.Router) {
	user := router.Group("/user")
	user.Get("/me", middleware.DeserializeUser, GetMe)
	user.Delete("/:id", middleware.DeserializeUser, DeleteUser)
	user.Get("/", middleware.DeserializeUser, GetUser)
}
