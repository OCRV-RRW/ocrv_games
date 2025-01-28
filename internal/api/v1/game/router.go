package game

import (
	"Games/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func AddRoutes(router fiber.Router) {
	game := router.Group("/games")
	game.Post("/", middleware.DeserializeUser, CreateGame)
	game.Get("/", middleware.DeserializeUser, GetGames)
	game.Delete("/:name", middleware.DeserializeUser, DeleteGame)
	game.Patch("/:name", middleware.DeserializeUser, UpdateGame)
}
