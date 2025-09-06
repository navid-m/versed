package main

import (
	"log"
	"path/filepath"
	"strings"
	"verse/database"
	"verse/feeds"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/django/v3"
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

	var (
		store        = session.New(session.Config{KeyLookup: "cookie:session_id"})
		viewsPath, _ = filepath.Abs("./views")
		engine       = django.New(viewsPath, ".html")
		app          = fiber.New(fiber.Config{Views: engine})
		scheduler    = NewFeedScheduler(database.GetDB())
	)

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

		for i, f := range feedItems {
			if strings.TrimSpace(f.Description) == "" {
				feedItems[i].Description = "No description."
			}
			if strings.TrimSpace(f.Description) == "Comments" {
				feedItems[i].Description = "No description."
			}
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
		page := c.QueryInt("page", 1)
		limit := min(c.QueryInt("limit", 20), 50)
		offset := (page - 1) * limit
		items, err := feeds.GetAllFeedItemsWithPagination(database.GetDB(), limit, offset)
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

	app.Post("/api/vote", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		var voteRequest struct {
			FeedID   string `json:"feed_id"`
			VoteType string `json:"vote_type"`
		}

		if err := c.BodyParser(&voteRequest); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		newScore, err := feeds.HandleVote(database.GetDB(), voteRequest.FeedID, int(userID), voteRequest.VoteType)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"new_score": newScore,
		})
	})

	app.Post("/api/reading-list/save", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		var saveRequest struct {
			ItemID string `json:"item_id"`
		}

		if err := c.BodyParser(&saveRequest); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		err := database.SaveToReadingList(userID, saveRequest.ItemID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to save item to reading list",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
		})
	})

	app.Post("/api/reading-list/remove", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		var removeRequest struct {
			ItemID string `json:"item_id"`
		}

		if err := c.BodyParser(&removeRequest); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		err := database.RemoveFromReadingList(userID, removeRequest.ItemID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to remove item from reading list",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
		})
	})

	app.Get("/api/reading-list", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		rows, err := database.RetrieveReadingList(userID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to retrieve reading list",
			})
		}
		defer rows.Close()

		var items []feeds.FeedItem
		for rows.Next() {
			var item feeds.FeedItem
			var sourceName string

			err := rows.Scan(
				&item.ID, &item.SourceID, &item.Title, &item.URL, &item.Description,
				&item.Author, &item.PublishedAt, &item.Score, &item.CommentsCount, &item.CreatedAt, &sourceName,
			)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "Failed to scan reading list item",
				})
			}
			item.SourceName = sourceName
			items = append(items, item)
		}

		return c.JSON(fiber.Map{
			"items": items,
			"count": len(items),
		})
	})

	app.Get("/api/reading-list/check/:itemId", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		itemID := c.Params("itemId")
		saved, err := database.IsInReadingList(userID, itemID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to check reading list status",
			})
		}

		return c.JSON(fiber.Map{
			"saved": saved,
		})
	})

	app.Get("/reading-list", func(c *fiber.Ctx) error {
		userEmail := c.Locals("userEmail")
		if userEmail == nil {
			return c.Redirect("/signin")
		}

		userID := c.Locals("userID").(int)
		rows, err := database.GetDB().Query(database.ReadingListQuery, userID)
		if err != nil {
			log.Printf("Failed to get reading list: %v", err)
		}
		defer rows.Close()

		var feedItems []feeds.FeedItem
		for rows.Next() {
			var item feeds.FeedItem
			var sourceName string
			err := rows.Scan(&item.ID, &item.SourceID, &item.Title, &item.URL, &item.Description,
				&item.Author, &item.PublishedAt, &item.Score, &item.CommentsCount, &item.CreatedAt, &sourceName)
			if err != nil {
				log.Printf("Failed to scan reading list item: %v", err)
			}
			if strings.TrimSpace(item.Description) == "" {
				item.Description = "No description."
			}
			item.SourceName = sourceName
			feedItems = append(feedItems, item)
		}

		data := fiber.Map{
			"FeedItems": feedItems,
			"Email":     userEmail,
		}

		return c.Render("reading-list", data)
	})

	log.Fatal(app.Listen(":3000"))
}
