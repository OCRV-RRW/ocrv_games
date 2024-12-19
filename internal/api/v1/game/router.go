package game

import (
	"Games/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func AddRoutes(router fiber.Router) {
	router.Post("/", middleware.DeserializeUser, CreateGame)
	router.Get("/", middleware.DeserializeUser, GetGames)
	router.Delete("/:name", middleware.DeserializeUser, DeleteGame)
}
