package handlers

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/navid-m/versed/database"
	"github.com/navid-m/versed/models"
)

// CreateSubverse handles the creation of a new subverse
func CreateSubverse(c *fiber.Ctx) error {
	// This is an admin-only function
	isAdmin := c.Locals("isAdmin").(bool)
	if !isAdmin {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	db := database.GetDB()

	var req struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Subverse name is required",
		})
	}

	subverse, err := database.CreateSubverse(db, req.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create subverse",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(subverse)
}

// GetSubverses handles retrieving all subverses
func GetSubverses(c *fiber.Ctx) error {
	db := database.GetDB()

	subverses, err := database.GetSubverses(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get subverses",
		})
	}

	return c.JSON(fiber.Map{
		"subverses": subverses,
		"count":     len(subverses),
	})
}

// AddFeedToSubverse handles adding a feed to a subverse
func AddFeedToSubverse(c *fiber.Ctx) error {
	// This is an admin-only function
	isAdmin := c.Locals("isAdmin").(bool)
	if !isAdmin {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	subverseID, err := c.ParamsInt("subverseId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subverse ID",
		})
	}

	var req struct {
		FeedSourceID int `json:"feed_source_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.FeedSourceID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Feed source ID is required",
		})
	}

	db := database.GetDB()
	err = database.AddFeedToSubverse(db, subverseID, req.FeedSourceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add feed to subverse",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Feed added to subverse successfully",
	})
}

// RemoveFeedFromSubverse handles removing a feed from a subverse
func RemoveFeedFromSubverse(c *fiber.Ctx) error {
	// This is an admin-only function
	isAdmin := c.Locals("isAdmin").(bool)
	if !isAdmin {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	subverseID, err := c.ParamsInt("subverseId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subverse ID",
		})
	}

	feedSourceID, err := c.ParamsInt("feedId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid feed source ID",
		})
	}

	db := database.GetDB()
	err = database.RemoveFeedFromSubverse(db, subverseID, feedSourceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove feed from subverse",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Feed removed from subverse successfully",
	})
}

// GetSubverseFeeds handles getting all feeds for a subverse
func GetSubverseFeeds(c *fiber.Ctx) error {
	subverseID, err := c.ParamsInt("subverseId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subverse ID",
		})
	}

	db := database.GetDB()
	feeds, err := database.GetSubverseFeeds(db, subverseID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get subverse feeds",
		})
	}

	return c.JSON(fiber.Map{
		"feeds": feeds,
		"count": len(feeds),
	})
}

// ViewSubverse handles viewing a specific subverse page
func ViewSubverse(c *fiber.Ctx) error {
	subverseName := c.Params("subverseName")
	if subverseName == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Subverse name is required")
	}

	db := database.GetDB()

	// Get subverse by name
	var subverse models.Subverse
	err := db.QueryRow("SELECT id, name, created_at FROM subverses WHERE name = ?", subverseName).Scan(
		&subverse.ID, &subverse.Name, &subverse.CreatedAt)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Subverse not found")
	}

	// Get posts for this subverse
	posts, err := database.GetPostsBySubverse(db, subverse.ID, 20, 0)
	if err != nil {
		log.Printf("Failed to get subverse posts: %v", err)
		posts = []models.Post{}
	}

	userEmail := c.Locals("userEmail")
	userUsername := c.Locals("userUsername")
	userID := c.Locals("userID")

	data := fiber.Map{
		"Subverse":     subverse,
		"Posts":        posts,
		"SubverseName": subverseName,
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

	return c.Render("subverse", data)
}
