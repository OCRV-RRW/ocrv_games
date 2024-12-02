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
	auth.Get("/verifyemail/:verificationCode", VerifyEmail)
	auth.Post("/forgotpassword", ForgotPassword)
	auth.Post("/resetpassword/:resetToken", ResetPassword)
}
