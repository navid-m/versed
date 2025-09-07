package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"verse/database"
	"verse/feeds"

	"github.com/gofiber/fiber/v2"
)

// Returns all categories for the authenticated user
func GetUserCategories(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	db := database.GetDB()

	categories, err := database.GetUserCategories(db, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get user categories",
		})
	}

	return c.JSON(fiber.Map{
		"categories": categories,
		"count":      len(categories),
	})
}

// Creates a new category for the authenticated user
func CreateUserCategory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)
	db := database.GetDB()

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Category name is required",
		})
	}

	category, err := database.CreateUserCategory(db, userID, req.Name, req.Description)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create category",
		})
	}

	return c.Status(201).JSON(category)
}

// Updates an existing category
func UpdateUserCategory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)
	categoryIDStr := c.Params("id")
	db := database.GetDB()

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid category ID",
		})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Category name is required",
		})
	}

	err = database.UpdateUserCategory(db, userID, categoryID, req.Name, req.Description)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update category",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Category updated successfully",
	})
}

// Deletes a category and all its feed associations
func DeleteUserCategory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)
	categoryIDStr := c.Params("id")
	db := database.GetDB()

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid category ID",
		})
	}

	err = database.DeleteUserCategory(db, userID, categoryID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete category",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Category deleted successfully",
	})
}

// Returns all feeds in a specific category
func GetCategoryFeeds(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)
	categoryIDStr := c.Params("id")
	db := database.GetDB()

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid category ID",
		})
	}

	_, err = database.GetUserCategoryByID(db, userID, categoryID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Category not found",
		})
	}

	feeds, err := database.GetFeedsInUserCategory(db, userID, categoryID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get category feeds",
		})
	}

	return c.JSON(fiber.Map{
		"feeds": feeds,
		"count": len(feeds),
	})
}

// Returns feed items from all feeds in a specific category
func GetCategoryFeedItems(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)
	categoryIDStr := c.Params("id")
	db := database.GetDB()

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid category ID",
		})
	}

	_, err = database.GetUserCategoryByID(db, userID, categoryID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Category not found",
		})
	}
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
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get category feed items",
		})
	}
	defer rows.Close()

	var items []feeds.FeedItem
	for rows.Next() {
		var item feeds.FeedItem
		var sourceName string
		err := rows.Scan(&item.ID, &item.SourceID, &item.Title, &item.URL, &item.Description,
			&item.Author, &item.PublishedAt, &item.Score, &item.CommentsCount, &item.CreatedAt, &sourceName)
		if err != nil {
			continue
		}
		item.SourceName = sourceName
		items = append(items, item)
	}

	return c.JSON(fiber.Map{
		"items": items,
		"count": len(items),
	})
}

// Adds a feed source to a user's category
func AddFeedToCategory(c *fiber.Ctx) error {
	fmt.Printf("=== AddFeedToCategory handler called ===\n")
	fmt.Printf("Method: %s, Path: %s\n", c.Method(), c.Path())
	fmt.Printf("Params: %v\n", c.AllParams())
	fmt.Printf("Body: %s\n", string(c.Body()))

	userID := c.Locals("userID").(int)
	categoryIDStr := c.Params("id")
	db := database.GetDB()

	fmt.Printf("UserID: %d, CategoryIDStr: %s\n", userID, categoryIDStr)

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		fmt.Printf("Invalid category ID: %v\n", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid category ID",
		})
	}

	var req struct {
		FeedSourceID int `json:"feed_source_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		fmt.Printf("Body parsing error: %v\n", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	fmt.Printf("Parsed request: FeedSourceID=%d\n", req.FeedSourceID)

	if req.FeedSourceID <= 0 {
		fmt.Printf("Validation failed: FeedSourceID=%d is invalid\n", req.FeedSourceID)
		return c.Status(400).JSON(fiber.Map{
			"error": "Valid feed source ID is required",
		})
	}

	err = database.AddFeedToUserCategory(db, userID, categoryID, req.FeedSourceID)
	if err != nil {
		fmt.Printf("Failed to add feed to category: %v\n", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to add feed to category",
		})
	}

	fmt.Printf("=== Successfully added feed to category ===\n")
	return c.JSON(fiber.Map{
		"message": "Feed added to category successfully",
	})
}

// Removes a feed source from a user's category
func RemoveFeedFromCategory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)
	categoryIDStr := c.Params("categoryId")
	feedSourceIDStr := c.Params("feedId")
	db := database.GetDB()

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid category ID",
		})
	}

	feedSourceID, err := strconv.Atoi(feedSourceIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid feed source ID",
		})
	}

	err = database.RemoveFeedFromUserCategory(db, userID, categoryID, feedSourceID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to remove feed from category",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Feed removed from category successfully",
	})
}

// Creates a new feed source and adds it to a category
func CreateAndAddFeedToCategory(c *fiber.Ctx) error {
	// Debug logging at the very beginning
	fmt.Printf("=== CreateAndAddFeedToCategory handler called ===\n")
	fmt.Printf("Method: %s, Path: %s\n", c.Method(), c.Path())
	fmt.Printf("Params: %v\n", c.AllParams())
	fmt.Printf("Headers: %v\n", c.GetReqHeaders())
	rawBody := string(c.Body())
	fmt.Printf("Raw Body: %s\n", rawBody)

	userID := c.Locals("userID").(int)
	categoryIDStr := c.Params("id")
	db := database.GetDB()

	fmt.Printf("UserID: %d, CategoryIDStr: %s\n", userID, categoryIDStr)

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		fmt.Printf("Invalid category ID: %v\n", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid category ID",
		})
	}

	var req struct {
		Type string `json:"type"` // "reddit" or "rss"
		URL  string `json:"url"`
		Name string `json:"name"`
	}

	if err := c.BodyParser(&req); err != nil {
		fmt.Printf("Body parsing error: %v\n", err)
		fmt.Printf("Failed to parse body: %s\n", rawBody)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Debug logging
	fmt.Printf("Parsed request: type='%s', url='%s', name='%s'\n", req.Type, req.URL, req.Name)

	req.URL = strings.TrimSpace(req.URL)
	req.Name = strings.TrimSpace(req.Name)

	fmt.Printf("After trimming: type='%s', url='%s', name='%s'\n", req.Type, req.URL, req.Name)

	if req.URL == "" || req.Name == "" {
		fmt.Printf("Validation failed: URL='%s', Name='%s'\n", req.URL, req.Name)
		return c.Status(400).JSON(fiber.Map{
			"error": "URL and name are required",
		})
	}

	if req.Type == "" {
		fmt.Printf("Validation failed: Type='%s'\n", req.Type)
		return c.Status(400).JSON(fiber.Map{
			"error": "Feed type is required",
		})
	}

	// Create the feed source
	var source *feeds.FeedSource
	if req.Type == "reddit" {
		// Extract subreddit name from URL
		if !strings.Contains(req.URL, "reddit.com/r/") {
			fmt.Printf("Reddit validation failed: URL='%s' doesn't contain 'reddit.com/r/'\n", req.URL)
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid Reddit URL format",
			})
		}
		source, err = feeds.CreateOrUpdateFeedSource(db, req.Name, req.URL)
	} else {
		source, err = feeds.CreateOrUpdateFeedSource(db, req.Name, req.URL)
	}

	if err != nil {
		fmt.Printf("Failed to create feed source: %v\n", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create feed source",
		})
	}

	// Add to category
	err = database.AddFeedToUserCategory(db, userID, categoryID, source.ID)
	if err != nil {
		fmt.Printf("Failed to add feed to category: %v\n", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to add feed to category",
		})
	}

	fmt.Printf("=== Successfully added feed to category ===\n")
	return c.Status(201).JSON(fiber.Map{
		"feed_source": source,
		"message":     "Feed created and added to category successfully",
	})
}
