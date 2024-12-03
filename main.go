package main

import (
	"Games/internal/api/v1/auth"
	"Games/internal/api/v1/game"
	"Games/internal/api/v1/tag"
	"Games/internal/api/v1/user"
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
// @host localhost:8000
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

	cfg := swagger.Config{
		BasePath: "/",
		FilePath: "./docs/swagger.json",
		Path:     "swagger",
		Title:    "Swagger API Docs",
	}

	app.Use(swagger.New(cfg))

	app.Mount("/api/v1", micro)
	app.Use(logger.New())

	//app.Use(cors.New(cors.Config{
	//	AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin",
	//	AllowOrigins:     "*",
	//	AllowCredentials: true,
	//	AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	//}))

	micro.Route("/", auth.AddRoutes)
	micro.Route("/game", game.AddRoutes)
	micro.Route("/tag", tag.AddRoutes)
	micro.Route("/", user.AddRoutes)

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
