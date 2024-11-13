package main

import (
	"Games/internal/config"
	"Games/internal/database"
	"Games/internal/routes"
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {
	conf, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalln("Failed to load environment variables! \n", err.Error())
	}

	app := fiber.New()

	dbErr := database.InitDB(&conf)

	if dbErr != nil {
		panic(dbErr)
	}

	list := app.Group("/list")

	list.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to the Todo-List-API Tutorial :)")
	}) // "/" - Default route to return the given string.

	list.Get("/tasks", routes.GetAllTasks) //Get endpoint for fetching all the tasks.

	list.Get("/task/:id", routes.GetTask) //Get endpoint for fetching a single task.

	list.Post("/add_task", routes.AddTask) //Post endpoint for add a new task.

	list.Delete("/delete_task/:id", routes.DeleteTask) //Delete endpoint for removing an existing task.

	list.Patch("/update_task/:id", routes.UpdateTask) //Patch endpoint for updating an existing task.

	app.Listen(":8000")
}
