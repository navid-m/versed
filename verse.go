package main

import (
	"log"
	"path/filepath"
	"verse/database"
	"verse/feeds"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/mustache/v2"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	if err := database.InitDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	err := feeds.ResetAllFeedTimestamps(database.GetDB())
	if err != nil {
		log.Printf("Warning: Failed to reset feed timestamps: %v", err)
	}
	defer database.CloseConnection()
	feeds.DebugFeeds(database.GetDB())
	store := session.New(session.Config{
		KeyLookup: "cookie:session_id",
	})
	viewsPath, _ := filepath.Abs("./views")
	engine := mustache.New(viewsPath, ".mustache")
	var (
		app = fiber.New(fiber.Config{
			Views: engine,
		})
	)

	scheduler := NewFeedScheduler(database.GetDB())
	scheduler.Start()
	defer scheduler.Stop()

	app.Use(func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err == nil {
			userID := sess.Get("user_id")
			userEmail := sess.Get("user_email")
			if userID != nil && userEmail != nil {
				c.Locals("userID", userID)
				c.Locals("userEmail", userEmail)
			}
		}
		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {
		userEmail := c.Locals("userEmail")

		feedItems, err := feeds.GetAllFeedItems(database.GetDB(), 20)
		if err != nil {
			log.Printf("Failed to get feed items: %v", err)
		}

		data := fiber.Map{
			"FeedItems": feedItems,
		}
		if userEmail != nil {
			data["Email"] = userEmail
		}
		return c.Render("index", data)
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
		var (
			email    = c.FormValue("email")
			password = c.FormValue("password")
		)
		if email == "" || password == "" {
			return c.Status(400).SendString("Email and password are required")
		}
		user, err := database.GetUserByEmail(email)
		if err != nil || user.Password != password {
			return c.Status(401).SendString("Invalid credentials")
		}

		sess, err := store.Get(c)
		if err != nil {
			return c.Status(500).SendString("Session error")
		}
		sess.Set("user_id", user.ID)
		sess.Set("user_email", user.Email)
		if err := sess.Save(); err != nil {
			return c.Status(500).SendString("Failed to save session")
		}

		return c.Redirect("/")
	})

	app.Get("/signout", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err == nil {
			sess.Delete("user_id")
			sess.Delete("user_email")
			sess.Save()
		}
		return c.Redirect("/")
	})

	app.Get("/api/feeds", func(c *fiber.Ctx) error {
		limit := min(c.QueryInt("limit", 50), 200)
		items, err := feeds.GetAllFeedItems(database.GetDB(), limit)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to retrieve feed items",
			})
		}

		return c.JSON(fiber.Map{
			"items": items,
			"count": len(items),
		})
	})

	app.Get("/api/feeds/:source", func(c *fiber.Ctx) error {
		sourceName := c.Params("source")
		limit := min(c.QueryInt("limit", 30), 100)
		source, err := feeds.GetFeedSourceByName(database.GetDB(), sourceName)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Feed source not found",
			})
		}

		items, err := feeds.GetFeedItemsBySource(database.GetDB(), source.ID, limit)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to retrieve feed items",
			})
		}

		return c.JSON(fiber.Map{
			"source": sourceName,
			"items":  items,
			"count":  len(items),
		})
	})

	log.Fatal(app.Listen(":3000"))
}
