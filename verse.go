package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"
	"verse/database"
	"verse/feeds"
	"verse/handlers"

	"github.com/Masterminds/squirrel"
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
		customStorage = database.NewDBSessionStorage(database.GetDB())
		store         = session.New(session.Config{
			Storage:    customStorage,
			KeyLookup:  "cookie:session_id",
			Expiration: 24 * time.Hour,
		})
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
			userUsername := sess.Get("user_username")
			if userID != nil && userEmail != nil {
				c.Locals("userID", userID)
				c.Locals("userEmail", userEmail)
				c.Locals("userUsername", userUsername)
			}
		}
		return c.Next()
	})

	app.Static("/static", "./static")

	app.Get("/", func(c *fiber.Ctx) error {
		userEmail := c.Locals("userEmail")
		userUsername := c.Locals("userUsername")

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
		if userUsername != nil {
			data["Username"] = userUsername
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
		username := c.FormValue("username")
		password := c.FormValue("password")
		if email == "" || username == "" || password == "" {
			return c.Status(400).SendString("Email, username, and password are required")
		}
		if err := database.CreateUser(email, username, password); err != nil {
			fmt.Println(err)
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
		sess.Set("user_username", user.Username)
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
			sess.Delete("user_username")
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

		saved, err := database.SaveToReadingList(userID, saveRequest.ItemID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to save item to reading list",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"saved":   saved,
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

	app.Get("/reading-list", func(c *fiber.Ctx) error {
		userEmail := c.Locals("userEmail")
		userUsername := c.Locals("userUsername")
		if userEmail == nil {
			return c.Redirect("/signin")
		}

		userID := c.Locals("userID").(int)
		sqlQ, args, _ := database.ReadingListQueryBuilder.Where(squirrel.Eq{"rl.user_id": userID}).ToSql()
		rows, err := database.GetDB().Query(sqlQ, args...)
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
			"Username":  userUsername,
		}

		return c.Render("reading-list", data)
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

	app.Get("/u/:username/c/:categoryName", func(c *fiber.Ctx) error {
		username := c.Params("username")
		categoryName := c.Params("categoryName")

		log.Printf("=== URL Route Handler: /u/%s/c/%s ===", username, categoryName)

		userEmail := c.Locals("userEmail")
		userUsername := c.Locals("userUsername")

		log.Printf("User info - Email: %v, Username: %v", userEmail, userUsername)

		userID := c.Locals("userID").(int)
		db := database.GetDB()

		log.Printf("UserID: %d", userID)

		var categoryID int
		err := db.QueryRow("SELECT id FROM user_categories WHERE user_id = ? AND LOWER(name) = LOWER(?)", userID, categoryName).Scan(&categoryID)
		if err != nil {
			log.Printf("Category not found: %v (user_id=%d, name='%s')", err, userID, categoryName)
			return c.Status(404).SendString("Category not found")
		}

		log.Printf("Found category ID: %d", categoryID)

		feedsQuery := `
			SELECT fs.id, fs.name, fs.url, fs.last_updated
			FROM feed_sources fs
			JOIN user_category_feeds ucf ON fs.id = ucf.feed_source_id
			WHERE ucf.user_id = ? AND ucf.category_id = ?
		`
		feedRows, err := db.Query(feedsQuery, userID, categoryID)
		feedCount := 0
		if err != nil {
			log.Printf("Failed to get category feeds: %v", err)
		} else {
			log.Printf("Checking feeds associated with category...")
			for feedRows.Next() {
				var feedID int
				var feedName, feedURL string
				var lastUpdated time.Time
				err := feedRows.Scan(&feedID, &feedName, &feedURL, &lastUpdated)
				if err != nil {
					log.Printf("Feed row scan error: %v", err)
					continue
				}
				feedCount++
				log.Printf("Feed %d: %s (%s) - Last updated: %v", feedID, feedName, feedURL, lastUpdated)
			}
			feedRows.Close()
		}
		log.Printf("Total feeds in category: %d", feedCount)
		query := `
			SELECT fi.id, fi.source_id, fi.title, fi.url, fi.description, fi.author, fi.published_at, fi.score, fi.comments_count, fi.created_at, fs.name as source_name
			FROM feed_items fi
			JOIN feed_sources fs ON fi.source_id = fs.id
			JOIN user_category_feeds ucf ON fs.id = ucf.feed_source_id
			WHERE ucf.user_id = ? AND ucf.category_id = ?
			ORDER BY fi.published_at DESC
			LIMIT 50
		`

		rows, err := db.Query(query, userID, categoryID)
		if err != nil {
			log.Printf("Database query error: %v", err)
		} else {
			log.Printf("Database query executed successfully")
		}
		defer func() {
			if rows != nil {
				rows.Close()
			}
		}()

		var items []feeds.FeedItem
		if rows != nil {
			for rows.Next() {
				var item feeds.FeedItem
				var sourceName string
				err := rows.Scan(&item.ID, &item.SourceID, &item.Title, &item.URL, &item.Description,
					&item.Author, &item.PublishedAt, &item.Score, &item.CommentsCount, &item.CreatedAt, &sourceName)
				if err != nil {
					log.Printf("Row scan error: %v", err)
					continue
				}
				item.SourceName = sourceName
				items = append(items, item)
			}
		}

		log.Printf("Found %d feed items for category", len(items))
		if len(items) > 0 {
			log.Printf("Sample item: %s", items[0].Title)
		}

		if feedCount > 0 && len(items) == 0 {
			log.Printf("No feed items found, triggering feed processing for category feeds...")

			resetQuery := `
				UPDATE feed_sources
				SET last_updated = datetime('2000-01-01 00:00:00')
				WHERE id IN (
					SELECT feed_source_id
					FROM user_category_feeds
					WHERE user_id = ? AND category_id = ?
				)
			`
			_, err = db.Exec(resetQuery, userID, categoryID)
			if err != nil {
				log.Printf("Failed to reset feed timestamps: %v", err)
			} else {
				log.Printf("Reset timestamps for %d feeds in category", feedCount)
			}
		}

		for i, item := range items {
			if strings.TrimSpace(item.Description) == "" {
				items[i].Description = "No description."
			}
			if strings.TrimSpace(item.Description) == "Comments" {
				items[i].Description = "No description."
			}
		}

		data := fiber.Map{
			"FeedItems":    items,
			"CategoryName": categoryName,
			"Username":     username,
		}

		if userEmail != nil {
			data["Email"] = userEmail
		}
		if userUsername != nil {
			data["Username"] = userUsername
		}

		log.Printf("Rendering template with %d items, category: %s", len(items), categoryName)
		return c.Render("index", data)
	})

	app.Get("/api/categories", handlers.GetUserCategories)
	app.Post("/api/categories", handlers.CreateUserCategory)
	app.Put("/api/categories/:id", handlers.UpdateUserCategory)
	app.Delete("/api/categories/:id", handlers.DeleteUserCategory)
	app.Get("/api/categories/:id/feeds", handlers.GetCategoryFeeds)
	app.Get("/api/categories/:id/items", handlers.GetCategoryFeedItems)
	app.Post("/api/categories/:id/feeds", handlers.AddFeedToCategory)
	app.Delete("/api/categories/:categoryId/feeds/:feedId", handlers.RemoveFeedFromCategory)
	app.Post("/api/categories/:id/feeds/create", handlers.CreateAndAddFeedToCategory)

	app.Get("/profile", func(c *fiber.Ctx) error {
		userID := c.Locals("userID")
		userEmail := c.Locals("userEmail")
		userUsername := c.Locals("userUsername")

		if userID == nil || userEmail == nil {
			return c.Redirect("/signin")
		}

		data := fiber.Map{
			"Email":    userEmail,
			"Username": userUsername,
		}

		return c.Render("profile", data)
	})

	app.Post("/profile/update", func(c *fiber.Ctx) error {
		userID := c.Locals("userID")
		if userID == nil {
			return c.Redirect("/signin")
		}

		var (
			email           = c.FormValue("email")
			username        = c.FormValue("username")
			currentPassword = c.FormValue("current_password")
			newPassword     = c.FormValue("new_password")
			confirmPassword = c.FormValue("confirm_password")
		)

		if email == "" || username == "" || currentPassword == "" {
			return c.Status(400).SendString("Email, username, and current password are required")
		}
		user, err := database.GetUserByEmail(c.Locals("userEmail").(string))
		if err != nil || user.Password != currentPassword {
			return c.Status(401).SendString("Invalid current password")
		}
		password := user.Password
		if newPassword != "" {
			if newPassword != confirmPassword {
				return c.Status(400).SendString("New passwords do not match")
			}
			password = newPassword
		}
		err = database.UpdateUser(userID.(int), email, username, password)
		if err != nil {
			return c.Status(500).SendString("Failed to update profile")
		}
		sess, err := store.Get(c)
		if err == nil {
			sess.Set("user_email", email)
			sess.Set("user_username", username)
			sess.Save()
		}

		return c.Redirect("/profile?success=1")
	})

	log.Fatal(app.Listen(":3000"))
}
