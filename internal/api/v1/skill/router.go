package skill

import (
	"Games/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func AddRoutes(router fiber.Router) {
	skill := router.Group("/skills")
	skill.Post("/", middleware.DeserializeUser, CreateSkill)
	skill.Get("/", middleware.DeserializeUser, GetSkills)
	skill.Delete("/:name", middleware.DeserializeUser, DeleteSkill)
	skill.Patch("/:name", middleware.DeserializeUser, UpdateSkill)
}
