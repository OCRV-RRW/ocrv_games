package tag

import (
	"Games/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func AddRoutes(router fiber.Router) {
	router.Post("/", middleware.DeserializeUser, CreateTag)
	router.Get("/", GetTags)
}
