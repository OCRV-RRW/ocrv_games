package routes

import "github.com/gofiber/fiber/v2"

func GetTask(c *fiber.Ctx) error {
	return c.SendString("A Single Task") // for getting a single task.
}

func GetAllTasks(c *fiber.Ctx) error {
	return c.SendString("ALL Tasks") // for getting all the tasks.
}

func AddTask(c *fiber.Ctx) error {
	return c.SendString("Added a Task") // for adding a new task.
}

func DeleteTask(c *fiber.Ctx) error {
	return c.SendString("Deleted a Task") // for deleting a task.
}

func UpdateTask(c *fiber.Ctx) error {
	return c.SendString("Updated a Task") // for updating a task.
}
