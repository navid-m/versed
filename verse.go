package main

import (
	"log"
	"verse/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/mustache/v2"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := database.InitDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	defer database.CloseConnection()

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

	app.Post("/signup", func(c *fiber.Ctx) error {
		email := c.FormValue("email")
		password := c.FormValue("password")

		if email == "" || password == "" {
			return c.Status(400).SendString("Email and password are required")
		}
		if err := database.CreateUser(email, password); err != nil {
			return c.Status(500).SendString("Failed to create user")
		}

		return c.Redirect("/")
	})

	app.Post("/signin", func(c *fiber.Ctx) error {
		email := c.FormValue("email")
		password := c.FormValue("password")
		if email == "" || password == "" {
			return c.Status(400).SendString("Email and password are required")
		}
		user, err := database.GetUserByEmail(email)
		if err != nil || user.Password != password {
			return c.Status(401).SendString("Invalid credentials")
		}
		return c.Render("index", fiber.Map{
			"Email": user.Email,
		})
	})

	log.Fatal(app.Listen(":3000"))
}
