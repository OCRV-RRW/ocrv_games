package auth

import (
	"Games/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func AddRoutes(router fiber.Router) {
	router.Post("/register", SignUpUser)
	router.Post("/refresh", RefreshAccessToken)
	router.Post("/login", SignInUser)
	router.Get("/logout", middleware.DeserializeUser, LogoutUser)
}
