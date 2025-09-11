package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

const VERSION = "1.0.0"

func main() {
	app := fiber.New()

	app.Get("/version", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"version": VERSION,
		})
	})

	log.Fatal(app.Listen(":8080"))
}