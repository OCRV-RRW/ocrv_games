package auth

import (
	"Games/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func AddRoutes(router fiber.Router) {
	auth := router.Group("auth")
	auth.Post("/register", SignUpUser)
	auth.Post("/refresh", RefreshAccessToken)
	auth.Post("/login", SignInUser)
	auth.Get("/logout", middleware.DeserializeUser, LogoutUser)
	auth.Post("/verify-email/:verificationCode", VerifyEmail)
	auth.Post("/forgot-password", ForgotPassword)
	auth.Patch("/reset-password/:resetToken", ResetPassword)
}
