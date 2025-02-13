package user

import (
	"Games/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func AddRoutes(router fiber.Router) {
	user := router.Group("/users")
	me := user.Group("/me")

	me.Get("/", middleware.DeserializeUser, GetMe)
	me.Patch("/", middleware.DeserializeUser, UpdateMe)
	me.Post("/skills", middleware.DeserializeUser, AddScore)
	me.Get("/skills", middleware.DeserializeUser, GetUserScores)

	user.Delete("/:id", middleware.DeserializeUser, middleware.AdminUser, DeleteUser)
	user.Get("/", middleware.DeserializeUser, GetUser)
	user.Patch("/:id", middleware.DeserializeUser, UpdateUser)
}

// REGISTRATION

// registration
//
// verify email
//
// continue registration

//
