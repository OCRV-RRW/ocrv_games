package main

import (
	"Games/internal/api/v1/auth"
	"Games/internal/config"
	"Games/internal/database"
	"fmt"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"log"
)

// @title Fiber Example API
// @version 1.0
// @description This is a sample swagger for Fiber
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /
func main() {
	conf, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalln("Failed to load environment variables! \n", err.Error())
	}
	database.InitDB(&conf)

	app := fiber.New()
	micro := fiber.New()

	cfg := swagger.Config{
		BasePath: "/",
		FilePath: "./docs/swagger.json",
		Path:     "swagger",
		Title:    "Swagger API Docs",
	}

	app.Use(swagger.New(cfg))

	//app.Get("/swagger/*", swagger.New(swagger.Config{ // custom
	//	URL:         "http://example.com/doc.json",
	//	DeepLinking: false,
	//	// Expand ("list") or Collapse ("none") tag groups by default
	//	DocExpansion: "none",
	//	// Prefill OAuth ClientId on Authorize popup
	//	OAuth: &swagger.OAuthConfig{
	//		AppName:  "OAuth Provider",
	//		ClientId: "21bb4edc-05a7-4afc-86f1-2e151e4ba6e2",
	//	},
	//	// Ability to change OAuth2 redirect uri location
	//	OAuth2RedirectUrl: "http://localhost:8080/swagger/oauth2-redirect.html",
	//}))

	app.Mount("/api", micro)
	app.Use(logger.New())
	//app.Use(cors.New(cors.Config{
	//	AllowOrigins:     "http://localhost:3000",
	//	AllowHeaders:     "Origin, Content-Type, Accept",
	//	AllowMethods:     "GET, POST",
	//	AllowCredentials: true,
	//}))

	micro.Route("/auth", auth.AddRoutes)

	//micro.Get("/users/me", middleware.mith.GetMe)

	micro.Get("/healthchecker", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "JWT Authentication with Golang, Fiber, and GORM",
		})
	})

	micro.All("*", func(c *fiber.Ctx) error {
		path := c.Path()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "fail",
			"message": fmt.Sprintf("Path: %v does not exists on this server", path),
		})
	})

	log.Fatal(app.Listen(":8000"))
}
