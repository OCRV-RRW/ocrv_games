package user

import (
	"Games/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func AddRoutes(router fiber.Router) {
	user := router.Group("/users")
	user.Get("/me", middleware.DeserializeUser, GetMe)
	user.Patch("/me", middleware.DeserializeUser, UpdateMe)
	user.Delete("/:id", middleware.DeserializeUser, DeleteUser)
	user.Get("/", middleware.DeserializeUser, GetUser)
	user.Post("/me/skills", middleware.DeserializeUser, AddScore)
	user.Get("/me/skills", middleware.DeserializeUser, GetUserScores)
}
