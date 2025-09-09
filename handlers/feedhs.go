package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/navid-m/versed/database"
	"github.com/navid-m/versed/feeds"
)

func FeedSourceHandler(c *fiber.Ctx) error {
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
}

func FeedsHandler(c *fiber.Ctx) error {
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
}
