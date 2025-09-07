package handlers

import (
	"github.com/navid-m/versed/database"
	"github.com/navid-m/versed/feeds"

	"strings"

	"github.com/gofiber/fiber/v2"
)

// Searches for feed items based on query string
func SearchFeedItems(c *fiber.Ctx) error {
	query := c.Query("q", "")
	if strings.TrimSpace(query) == "" {
		return c.JSON(fiber.Map{
			"items": []feeds.FeedItem{},
			"count": 0,
		})
	}

	rows, err := database.GetFeedItemsToQuery(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to search feed items",
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
