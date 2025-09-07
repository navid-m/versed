package database

import (
	"database/sql"
	"strings"
)

// Primary query for searching feed items.
const FeedItemsQuery = `SELECT fi.id, fi.source_id, fi.title, fi.url, fi.description, fi.author, fi.published_at, fi.score, fi.comments_count, fi.created_at, fs.name as source_name
	FROM feed_items fi
	JOIN feed_sources fs ON fi.source_id = fs.id
	WHERE LOWER(fi.title) LIKE ? OR LOWER(fi.description) LIKE ? OR LOWER(fi.author) LIKE ?
	ORDER BY fi.published_at DESC
	LIMIT 50`

// Scheduler query.
const FeedInsertionQuery = `INSERT INTO feed_sources (name, url, last_updated, update_interval) 
			VALUES (?, ?, datetime('2000-01-01 00:00:00'), 3600)`

// Primarily for search purposes
func GetFeedItemsToQuery(query string) (*sql.Rows, error) {
	var (
		searchQuery = `%` + strings.ToLower(query) + `%`
	)
	rows, err := GetDB().Query(FeedItemsQuery, searchQuery, searchQuery, searchQuery)
	return rows, err
}
