package game

import (
	"Games/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func AddRoutes(router fiber.Router) {
	game := router.Group("/games")
	game.Post("/", middleware.DeserializeUser, middleware.AdminUser, CreateGame)
	game.Get("/", middleware.DeserializeUser, GetGames)
	game.Delete("/:name", middleware.DeserializeUser, middleware.AdminUser, DeleteGame)
	game.Patch("/:name", middleware.DeserializeUser, middleware.AdminUser, UpdateGame)
	game.Post("/upload-preview/:name", UploadGamePreview)
}
