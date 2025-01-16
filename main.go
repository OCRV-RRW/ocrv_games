package main

import (
	"Games/internal/api/v1/auth"
	"Games/internal/api/v1/game"
	"Games/internal/api/v1/skill"
	"Games/internal/api/v1/user"
	"Games/internal/config"
	"Games/internal/database"
	"fmt"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"log"
)

// @title GamePlatform API
// @version 1.0
// @description This is game platform swagger
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url h
// ttp://www.apache.org/licenses/LICENSE-2.0.html
// @host https://ocrv-game
// @BasePath /
func main() {
	conf, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalln("Failed to load environment variables! \n", err.Error())
	}
	database.InitDB(&conf)
	database.ConnectRedis(&conf)

	app := fiber.New()
	micro := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowHeaders: "Origin,Content-Type,Accept,Content-Length,Accept-Language," +
			"Accept-Encoding,Connection,Access-Control-Allow-Origin,Authorization",
		AllowOrigins:     "http://localhost:3000,http://localhost:8000",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	cfg := swagger.Config{
		BasePath: "/game-platform",
		FilePath: "./docs/swagger.json",
		Path:     "swagger",
		Title:    "Swagger API Docs",
	}

	app.Use(swagger.New(cfg))
	app.Mount("/api/v1", micro)
	app.Use(logger.New())

	micro.Route("/", auth.AddRoutes)
	micro.Route("/", game.AddRoutes)
	micro.Route("/", skill.AddRoutes)
	micro.Route("/", user.AddRoutes)

	micro.Get("/healthchecker", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "good",
		})
	})

	micro.All("*", func(c *fiber.Ctx) error {
		path := c.Path()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "fail",
			"message": fmt.Sprintf("Path: %v does not exists on this server", path),
		})
	})

	log.Fatal(app.Listen(fmt.Sprintf(":%d", conf.AppPort)))
}
