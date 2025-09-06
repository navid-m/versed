package handlers

import (
	"verse/database"
	"verse/feeds"

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

	var (
		searchQuery = `%` + strings.ToLower(query) + `%`
		sqlQuery    = `SELECT fi.id, fi.source_id, fi.title, fi.url, fi.description, fi.author, fi.published_at, fi.score, fi.comments_count, fi.created_at, fs.name as source_name
				FROM feed_items fi
				JOIN feed_sources fs ON fi.source_id = fs.id
				WHERE LOWER(fi.title) LIKE ? OR LOWER(fi.description) LIKE ? OR LOWER(fi.author) LIKE ?
				ORDER BY fi.published_at DESC
				LIMIT 50`
	)

	rows, err := database.GetDB().Query(sqlQuery, searchQuery, searchQuery, searchQuery)
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
