// Versed - A content aggregator
//
// Copyright (c) 2025 Navid Momtahen
// License AGPL-3.0
// https://github.com/navid-m/versed

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/django/v3"

	"github.com/navid-m/versed/database"
	"github.com/navid-m/versed/feeds"
	"github.com/navid-m/versed/handlers"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.Printf("=== MAIN FUNCTION STARTED ===")
	if err := database.InitDatabase(); err != nil {
		log.Printf("=== DATABASE INIT ERROR: %v ===", err)
		log.Fatal("Failed to initialize database:", err)
	}
	log.Printf("=== DATABASE INITIALIZED SUCCESSFULLY ===")
	err := feeds.ResetAllFeedTimestamps(database.GetDB())
	if err != nil {
		log.Printf("Warning: Failed to reset feed timestamps: %v", err)
	}
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

	log.Printf("=== VIEWS PATH: %s ===", viewsPath)
	log.Printf("=== TEMPLATE ENGINE CREATED ===")
	log.Printf("=== STARTING FEED SCHEDULER ===")
	scheduler.Start()
	defer scheduler.Stop()

	log.Printf("=== SETTING UP MIDDLEWARE ===")
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

	log.Printf("=== SETTING UP IP BAN MIDDLEWARE ===")
	app.Use(func(c *fiber.Ctx) error {
		clientIP := c.IP()

		isBanned, err := database.IsIPBanned(clientIP)
		if err != nil {
			log.Printf("Error checking IP ban status for %s: %v", clientIP, err)
			return c.Next()
		}

		if isBanned {
			log.Printf("Blocked access from banned IP: %s", clientIP)
			return c.Status(403).SendString("Access denied. Your IP address has been banned.")
		}
		return c.Next()
	})

	log.Printf("=== SETTING UP STATIC FILE MIDDLEWARE ===")
	app.Use("/static", func(c *fiber.Ctx) error {
		path := c.Path()
		if strings.HasPrefix(path, "/static/js/admin") ||
			strings.HasPrefix(path, "/static/css/admin") ||
			strings.Contains(path, "/admin/") {

			userID := c.Locals("userID")
			if userID == nil {
				return c.Status(404).SendString("Cannot GET /static/js/")
			}

			isAdmin, err := database.IsUserAdmin(userID.(int))
			if err != nil {
				log.Printf("Error checking admin status for user %v: %v", userID, err)
				return c.Status(404).SendString("Cannot GET /static/js/")
			}

			if !isAdmin {
				return c.Status(404).SendString("Cannot GET /static/js/")
			}
		}

		return c.Next()
	})

	app.Static("/static", "./static")

	log.Printf("=== SETTING UP ROUTES ===")
	app.Get("/", func(c *fiber.Ctx) error {
		userEmail := c.Locals("userEmail")
		userUsername := c.Locals("userUsername")
		userID := c.Locals("userID")

		var feedItems []feeds.FeedItem
		var err error

		if userID != nil {
			feedItems, err = feeds.GetAllFeedItemsForUser(database.GetDB(), userID.(int), 20)
		} else {
			feedItems, err = feeds.GetAllFeedItems(database.GetDB(), 20)
		}

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
		err = django.New("./views", ".html").Render(c, "index", data)
		fmt.Println(err)
		return c.Render("index", data)
	})
	app.Get("/signin", func(c *fiber.Ctx) error {
		return c.Render("signin", fiber.Map{})
	})
	app.Get("/signup", func(c *fiber.Ctx) error {
		return c.Render("signup", fiber.Map{})
	})
	app.Get("/about", func(c *fiber.Ctx) error {
		userEmail := c.Locals("userEmail")
		userUsername := c.Locals("userUsername")

		data := fiber.Map{}
		if userEmail != nil {
			data["Email"] = userEmail
		}
		if userUsername != nil {
			data["Username"] = userUsername
		}

		return c.Render("about", data)
	})

	app.Post("/signup", func(c *fiber.Ctx) error {
		email := c.FormValue("email")
		username := c.FormValue("username")
		password := c.FormValue("password")
		if email == "" || username == "" || password == "" {
			return c.Status(400).SendString("Email, username, and password are required")
		}

		if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
			return c.Status(400).SendString("Enter a valid email address")
		}

		if len(username) < 3 || len(username) > 20 {
			return c.Status(400).SendString("Username must be between 3 and 20 characters")
		}
		for _, r := range username {
			if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' {
				return c.Status(400).SendString("Username can only contain letters, numbers, and underscores")
			}
		}

		if len(password) < 3 {
			return c.Status(400).SendString("Password must be at least 3 characters long")
		}

		err := database.CreateUser(email, username, password, c.IP())
		if err != nil {
			switch err.Error() {
			case "email already in use":
				return c.Status(400).SendString("This email is already registered")
			case "username already taken":
				return c.Status(400).SendString("This username is already taken")
			default:
				log.Printf("Error creating user: %v", err)
				return c.Status(500).SendString("An error occurred while creating your account")
			}
		}

		user, err := database.GetUserByEmail(email)
		if err != nil {
			log.Printf("Error getting new user after signup: %v", err)
			return c.Redirect("/signin")
		}

		sess, err := store.Get(c)
		if err != nil {
			log.Printf("Error getting session: %v", err)
			return c.Redirect("/signin")
		}

		sess.Set("user_id", user.ID)
		sess.Set("user_email", user.Email)
		sess.Set("user_username", user.Username)

		if err := sess.Save(); err != nil {
			log.Printf("Error saving session: %v", err)
			return c.Redirect("/signin")
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
		if err != nil {
			log.Printf("Login failed for email %s: %v", email, err)
			return c.Status(401).SendString("Invalid credentials")
		}
		err = database.VerifyPassword(user.Password, password)
		if err != nil {
			log.Printf("Invalid password attempt for user %s", email)
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
		userID := c.Locals("userID")
		page := c.QueryInt("page", 1)
		limit := min(c.QueryInt("limit", 20), 50)
		offset := (page - 1) * limit

		var items []feeds.FeedItem
		var err error

		if userID != nil {
			items, err = feeds.GetAllFeedItemsWithPaginationForUser(database.GetDB(), userID.(int), limit, offset)
		} else {
			items, err = feeds.GetAllFeedItemsWithPagination(database.GetDB(), limit, offset)
		}

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

	app.Get("/api/search", handlers.SearchFeedItems)

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

		err = database.RemoveFromReadingList(userID, removeRequest.ItemID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to remove item from reading list",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"removed": true,
		})
	})

	app.Post("/api/posts/:itemId/hide", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		itemID := c.Params("itemId")
		if itemID == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Item ID is required",
			})
		}

		err := database.HideFeedItem(userID, itemID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to hide post",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Post hidden successfully",
		})
	})

	app.Post("/api/posts/:itemId/unhide", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		itemID := c.Params("itemId")
		if itemID == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Item ID is required",
			})
		}

		err := database.UnhideFeedItem(userID, itemID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to unhide post",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Post unhidden successfully",
		})
	})

	app.Get("/api/posts/:itemId/hidden", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		itemID := c.Params("itemId")
		if itemID == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Item ID is required",
			})
		}

		hidden, err := database.IsFeedItemHidden(userID, itemID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to check if post is hidden",
			})
		}

		return c.JSON(fiber.Map{
			"hidden": hidden,
		})
	})

	app.Get("/api/posts/hidden", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		itemIDs, err := database.GetHiddenFeedItems(userID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to get hidden posts",
			})
		}

		var hiddenItems []feeds.FeedItem
		for _, itemID := range itemIDs {
			query := `SELECT fi.id, fi.source_id, fi.title, fi.url, fi.description, fi.author, fi.published_at, COALESCE(fi.score, 0) as score, COALESCE(fi.comments_count, 0) as comments_count, fi.created_at, fs.name as source_name
				FROM feed_items fi
				JOIN feed_sources fs ON fi.source_id = fs.id
				WHERE fi.id = ?`
			var item feeds.FeedItem
			var sourceName string
			err := database.GetDB().QueryRow(query, itemID).Scan(
				&item.ID, &item.SourceID, &item.Title, &item.URL, &item.Description,
				&item.Author, &item.PublishedAt, &item.Score, &item.CommentsCount, &item.CreatedAt, &sourceName)
			if err != nil {
				continue
			}
			item.SourceName = sourceName
			hiddenItems = append(hiddenItems, item)
		}

		return c.JSON(fiber.Map{
			"hiddenItems": hiddenItems,
			"count":       len(hiddenItems),
		})
	})

	app.Get("/reading-list", func(c *fiber.Ctx) error {
		userEmail := c.Locals("userEmail")
		userUsername := c.Locals("userUsername")
		if userEmail == nil {
			return c.Redirect("/signin")
		}

		userID := c.Locals("userID").(int)
		rows, err := database.RetrieveReadingList(userID)
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

		if userEmail == nil {
			return c.Redirect("/signin")
		}

		userID := c.Locals("userID").(int)
		db := database.GetDB()

		log.Printf("UserID: %d", userID)

		categoryName = strings.TrimSpace(categoryName)
		categoryName = strings.ReplaceAll(categoryName, "-", " ")

		var categoryID int
		err := db.QueryRow("SELECT id FROM user_categories WHERE user_id = ? AND LOWER(name) = LOWER(?)", userID, categoryName).Scan(&categoryID)
		if err != nil {
			log.Printf("Category not found: %v (user_id=%d, name='%s')", err, userID, categoryName)
			return c.Status(404).SendString("Category not found")
		}

		log.Printf("Found category ID: %d", categoryID)

		feedRows, err := db.Query(database.FeedsQuery, userID, categoryID)
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

		rows, err := db.Query(database.PostFeedNextQuery, userID, categoryID)
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

			result, err := db.Exec(database.ResetQuery, userID, categoryID)
			if err != nil {
				log.Printf("Failed to reset feed timestamps: %v", err)
			} else {
				rowsAffected, _ := result.RowsAffected()
				log.Printf("Reset timestamps for %d feeds in category", rowsAffected)
				if rowsAffected > 0 {
					log.Printf("Attempting immediate feed processing...")
					var feedList []struct {
						id   int
						name string
						url  string
					}

					feedRows, err := db.Query(database.SurpFeedsQuery, userID, categoryID)
					if err != nil {
						log.Printf("Failed to get category feeds: %v", err)
					} else {
						for feedRows.Next() {
							var feed struct {
								id   int
								name string
								url  string
							}
							err := feedRows.Scan(&feed.id, &feed.name, &feed.url)
							if err != nil {
								log.Printf("Feed row scan error: %v", err)
								continue
							}
							feedList = append(feedList, feed)
						}
						feedRows.Close()
					}

					for _, feed := range feedList {
						log.Printf("Processing feed: %s (%s)", feed.name, feed.url)

						source := feeds.FeedSource{
							ID:             feed.id,
							Name:           feed.name,
							URL:            feed.url,
							LastUpdated:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
							UpdateInterval: 3600,
						}

						if feeds.ShouldUpdateFeed(source) {
							log.Printf("Feed needs update, fetching...")

							content, err := feeds.FetchFeed(feed.url)
							if err != nil {
								log.Printf("Failed to fetch feed %s: %v", feed.url, err)
								continue
							}

							parsedItems, err := feeds.ParseFeedWithParser(content, feed.id, feed.name)
							if err != nil {
								log.Printf("Failed to parse feed %s: %v", feed.url, err)
								continue
							}

							log.Printf("Parsed %d items from feed %s", len(parsedItems), feed.name)
							if len(parsedItems) > 0 {
								err = feeds.SaveFeedItems(db, parsedItems)
								if err != nil {
									log.Printf("Failed to save feed items for %s: %v", feed.name, err)
									continue
								}
								err = feeds.UpdateFeedSourceTimestamp(db, feed.id)
								if err != nil {
									log.Printf("Failed to update timestamp for feed %s: %v", feed.name, err)
								} else {
									log.Printf("Successfully processed and saved %d items for feed %s", len(parsedItems), feed.name)
								}
							}
						}
					}
				}
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

		data["Email"] = userEmail
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

	app.Get("/api/graph", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		log.Printf("Graph API called for user %d", userID)

		db := database.GetDB()
		categoryRows, err := db.Query("SELECT id, name FROM user_categories WHERE user_id = ?", userID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to get categories",
			})
		}
		defer categoryRows.Close()

		var nodes []fiber.Map
		var links []fiber.Map
		categoryMap := make(map[int]string)

		nodes = append(nodes, fiber.Map{
			"id":   "root",
			"name": "Categories",
			"type": "root",
		})

		for categoryRows.Next() {
			var catID int
			var catName string
			err := categoryRows.Scan(&catID, &catName)
			if err != nil {
				continue
			}

			nodes = append(nodes, fiber.Map{
				"id":   fmt.Sprintf("cat_%d", catID),
				"name": catName,
				"type": "category",
			})
			categoryMap[catID] = catName

			links = append(links, fiber.Map{
				"source": "root",
				"target": fmt.Sprintf("cat_%d", catID),
			})

			postRows, err := database.GraphFeedQuery(db, userID, catID)
			if err != nil {
				log.Printf("Error querying posts for category %d: %v", catID, err)
				continue
			}

			postCount := 0
			for postRows.Next() {
				var postID string
				var postTitle string
				err := postRows.Scan(&postID, &postTitle)
				if err != nil {
					log.Printf("Error scanning post row: %v", err)
					continue
				}
				postCount++

				nodes = append(nodes, fiber.Map{
					"id":   fmt.Sprintf("post_%s", postID),
					"name": postTitle,
					"type": "post",
				})

				links = append(links, fiber.Map{
					"source": fmt.Sprintf("cat_%d", catID),
					"target": fmt.Sprintf("post_%s", postID),
				})
			}
			postRows.Close()
			log.Printf("Category %d (%s): %d posts", catID, catName, postCount)
		}

		log.Printf("Graph API returning %d nodes, %d links", len(nodes), len(links))

		return c.JSON(fiber.Map{
			"nodes": nodes,
			"links": links,
		})
	})

	app.Get("/post/:itemId", func(c *fiber.Ctx) error {
		itemID := c.Params("itemId")
		userEmail := c.Locals("userEmail")
		userUsername := c.Locals("userUsername")
		userID := c.Locals("userID")

		db := database.GetDB()
		var post feeds.FeedItem
		var sourceName string

		err := db.QueryRow(database.PostFeedQuery, itemID).Scan(
			&post.ID, &post.SourceID, &post.Title, &post.URL, &post.Description,
			&post.Author, &post.PublishedAt, &post.Score, &post.CommentsCount,
			&post.CreatedAt, &sourceName,
		)

		fmt.Println("helloo")
		if err != nil {
			fmt.Println("hello2")
			postObj, postErr := database.GetPostByID(db, itemID)
			if postErr != nil {
				return c.Status(404).SendString("Post not found")
			}

			post.ID = postObj.ID
			post.Title = postObj.Title
			post.Description = postObj.Content
			post.Author = postObj.Username
			post.Score = postObj.Score
			post.CommentsCount = 0
			post.CreatedAt = postObj.CreatedAt
			sourceName = "Forum"

			comments, commentErr := database.GetCommentsByItemID(itemID)
			if commentErr != nil {
				log.Printf("Failed to get comments: %v", commentErr)
				comments = []database.Comment{}
			}

			data := fiber.Map{
				"post":     post,
				"comments": comments,
			}

			if userEmail != nil {
				data["Email"] = userEmail
			}
			if userUsername != nil {
				data["Username"] = userUsername
			}
			if userID != nil {
				data["userID"] = userID
			}

			return c.Render("post", data)
		}

		post.SourceName = sourceName

		comments, err := database.GetCommentsByItemID(itemID)
		fmt.Println(comments)
		log.Printf("=== POST PAGE DEBUG: Retrieved %d comments for item %s ===", len(comments), itemID)
		for i, comment := range comments {
			log.Printf("=== POST PAGE DEBUG: Comment %d: ID=%d, Content='%s', ParentID=%v, Replies=%d ===",
				i+1, comment.ID, comment.Content, comment.ParentID, len(comment.Replies))
			for j, reply := range comment.Replies {
				log.Printf("=== POST PAGE DEBUG:   Reply %d.%d: ID=%d, Content='%s' ===",
					i+1, j+1, reply.ID, reply.Content)
			}
		}

		if err != nil {
			log.Printf("Failed to get comments: %v", err)
			comments = []database.Comment{}
		}

		log.Printf("=== POST PAGE DEBUG: About to render template with %d comments ===", len(comments))

		data := fiber.Map{
			"post":     post,
			"comments": comments,
		}

		if userEmail != nil {
			data["Email"] = userEmail
		}
		if userUsername != nil {
			data["Username"] = userUsername
		}
		if userID != nil {
			data["userID"] = userID
		}
		return c.Render("post", data)
	})

	app.Get("/api/posts/:itemId", handlers.GetPostView)
	app.Get("/api/posts/:itemId/comments", handlers.GetComments)
	app.Post("/api/posts/:itemId/comments", handlers.CreateComment)
	app.Get("/api/comments/:commentId", handlers.GetComment)
	app.Put("/api/comments/:commentId", handlers.UpdateComment)
	app.Delete("/api/comments/:commentId", handlers.DeleteComment)

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

	app.Get("/graph", func(c *fiber.Ctx) error {
		userEmail := c.Locals("userEmail")
		userUsername := c.Locals("userUsername")

		if userEmail == nil {
			return c.Redirect("/signin")
		}

		data := fiber.Map{
			"Email":    userEmail,
			"Username": userUsername,
		}

		return c.Render("graph", data)
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

	adminMiddleware := func(c *fiber.Ctx) error {
		userID := c.Locals("userID")
		if userID == nil {
			return c.Status(401).Redirect("/signin")
		}

		isAdmin, err := database.IsUserAdmin(userID.(int))
		if err != nil {
			log.Printf("Error checking admin status for user %v: %v", userID, err)
			return c.Status(500).SendString("Internal server error")
		}

		if !isAdmin {
			return c.Status(403).SendString("Access denied. Admin privileges required.")
		}

		c.Locals("isAdmin", isAdmin)
		return c.Next()
	}

	app.Get("/admin", adminMiddleware, func(c *fiber.Ctx) error {
		userEmail := c.Locals("userEmail")
		userUsername := c.Locals("userUsername")

		data := fiber.Map{}
		if userEmail != nil {
			data["Email"] = userEmail
		}
		if userUsername != nil {
			data["Username"] = userUsername
		}

		return c.Render("admin", data)
	})

	app.Get("/api/admin/users", adminMiddleware, func(c *fiber.Ctx) error {
		users, err := database.GetAllUsers()
		if err != nil {
			log.Printf("Error getting users: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to retrieve users",
			})
		}

		return c.JSON(fiber.Map{
			"users": users,
		})
	})

	app.Get("/api/admin/banned-ips", adminMiddleware, func(c *fiber.Ctx) error {
		bannedIPs, err := database.GetAllBannedIPs()
		if err != nil {
			log.Printf("Error getting banned IPs: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to retrieve banned IPs",
			})
		}

		return c.JSON(fiber.Map{
			"bannedIPs": bannedIPs,
		})
	})

	app.Post("/api/admin/ban-ip", adminMiddleware, func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(int)

		var banRequest struct {
			IPAddress string `json:"ipAddress"`
			Reason    string `json:"reason"`
		}

		if err := c.BodyParser(&banRequest); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if banRequest.IPAddress == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "IP address is required",
			})
		}

		err := database.BanIP(banRequest.IPAddress, banRequest.Reason, userID)
		if err != nil {
			log.Printf("Error banning IP %s: %v", banRequest.IPAddress, err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to ban IP address",
			})
		}

		log.Printf("Admin %d banned IP: %s (reason: %s)", userID, banRequest.IPAddress, banRequest.Reason)
		return c.JSON(fiber.Map{
			"success": true,
			"message": "IP address banned successfully",
		})
	})

	app.Post("/api/admin/unban-ip", adminMiddleware, func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(int)

		var unbanRequest struct {
			IPAddress string `json:"ipAddress"`
		}

		if err := c.BodyParser(&unbanRequest); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if unbanRequest.IPAddress == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "IP address is required",
			})
		}

		err := database.UnbanIP(unbanRequest.IPAddress, userID)
		if err != nil {
			log.Printf("Error unbanning IP %s: %v", unbanRequest.IPAddress, err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to unban IP address",
			})
		}

		log.Printf("Admin %d unbanned IP: %s", userID, unbanRequest.IPAddress)
		return c.JSON(fiber.Map{
			"success": true,
			"message": "IP address unbanned successfully",
		})
	})

	app.Post("/api/admin/subverses", adminMiddleware, handlers.CreateSubverse)
	app.Get("/api/subverses", handlers.GetSubverses)
	app.Get("/s/:subverseName", handlers.ViewSubverse)

	app.Get("/api/admin/subverses/:subverseId/feeds", adminMiddleware, handlers.GetSubverseFeeds)
	app.Post("/api/admin/subverses/:subverseId/feeds", adminMiddleware, handlers.AddFeedToSubverse)
	app.Delete("/api/admin/subverses/:subverseId/feeds/:feedId", adminMiddleware, handlers.RemoveFeedFromSubverse)

	app.Get("/s/:subverseName/posts", handlers.GetSubversePosts)
	app.Get("/s/:subverseName/posts/search", handlers.SearchPosts)
	app.Post("/s/:subverseName/posts", handlers.CreatePost)
	app.Get("/posts/:postID", handlers.GetPost)
	app.Put("/posts/:postID", handlers.UpdatePost)
	app.Delete("/posts/:postID", handlers.DeletePost)
	app.Post("/api/posts/:postID/vote", handlers.VotePost)

	app.Get("/api/user/status", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.JSON(fiber.Map{
				"isAdmin": false,
			})
		}

		isAdmin, err := database.IsUserAdmin(userID)
		if err != nil {
			log.Printf("Error checking admin status for user %d: %v", userID, err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to check admin status",
			})
		}

		return c.JSON(fiber.Map{
			"isAdmin": isAdmin,
		})
	})

	port := 3000
	if len(os.Args) > 1 {
		if os.Args[1] == "prod" {
			log.Fatal(app.ListenTLS(":443", "/etc/letsencrypt/live/versed.cc/fullchain.pem", "/etc/letsencrypt/live/versed.cc/privkey.pem"))
		}
	} else {
		log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
	}
}
