package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"verse/database"
	"verse/feeds"

	"github.com/gofiber/fiber/v2"
)

// GetUserCategories returns all categories for the authenticated user
func GetUserCategories(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)
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

// CreateUserCategory creates a new category for the authenticated user
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

// UpdateUserCategory updates an existing category
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

// DeleteUserCategory deletes a category and all its feed associations
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

// GetCategoryFeeds returns all feeds in a specific category
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

	// Verify the category belongs to the user
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

// AddFeedToCategory adds a feed source to a user's category
func AddFeedToCategory(c *fiber.Ctx) error {
	// Debug logging at the very beginning
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

// RemoveFeedFromCategory removes a feed source from a user's category
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

// CreateAndAddFeedToCategory creates a new feed source and adds it to a category
func CreateAndAddFeedToCategory(c *fiber.Ctx) error {
	// Debug logging at the very beginning
	fmt.Printf("=== CreateAndAddFeedToCategory handler called ===\n")
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
		Type string `json:"type"` // "reddit" or "rss"
		URL  string `json:"url"`
		Name string `json:"name"`
	}

	if err := c.BodyParser(&req); err != nil {
		fmt.Printf("Body parsing error: %v\n", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Debug logging
	fmt.Printf("Received request: type=%s, url=%s, name=%s\n", req.Type, req.URL, req.Name)

	req.URL = strings.TrimSpace(req.URL)
	req.Name = strings.TrimSpace(req.Name)

	fmt.Printf("After trimming: type=%s, url=%s, name=%s\n", req.Type, req.URL, req.Name)

	if req.URL == "" || req.Name == "" {
		fmt.Printf("Validation failed: URL='%s', Name='%s'\n", req.URL, req.Name)
		return c.Status(400).JSON(fiber.Map{
			"error": "URL and name are required",
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
