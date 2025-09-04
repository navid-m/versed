package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/mustache/v2"
)

func main() {
	var (
		engine = mustache.New("./views", ".mustache")
		app    = fiber.New(fiber.Config{
			Views: engine,
		})
	)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})
	app.Get("/signin", func(c *fiber.Ctx) error {
		return c.Render("signin", fiber.Map{})
	})
	app.Get("/signup", func(c *fiber.Ctx) error {
		return c.Render("signup", fiber.Map{})
	})

	log.Fatal(app.Listen(":3000"))
}
