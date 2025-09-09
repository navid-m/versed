package database

import (
	"database/sql"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v2"
)

// Build feed items search query using Squirrel
var FeedItemsQueryBuilder = squirrel.Select(
	"fi.id",
	"fi.source_id",
	"fi.title",
	"fi.url",
	"fi.description",
	"fi.author",
	"fi.published_at",
	"fi.score",
	"fi.comments_count",
	"fi.created_at",
	"fs.name as source_name",
).From("feed_items fi").
	Join("feed_sources fs ON fi.source_id = fs.id").
	OrderBy("fi.published_at DESC").
	Limit(50)

// Build feed insertion query using Squirrel
var FeedInsertionBuilder = squirrel.Insert("feed_sources").
	Columns("name", "url", "last_updated", "update_interval").
	Values(squirrel.Expr("?, ?, datetime('2000-01-01 00:00:00'), 3600"))

// Primarily for search purposes
func GetFeedItemsToQuery(query string) (*sql.Rows, error) {
	var (
		searchQuery = `%` + strings.ToLower(query) + `%`
	)
	sqlQuery, args, err := FeedItemsQueryBuilder.Where(
		squirrel.Or{
			squirrel.Expr("LOWER(fi.title) LIKE ?", searchQuery),
			squirrel.Expr("LOWER(fi.description) LIKE ?", searchQuery),
			squirrel.Expr("LOWER(fi.author) LIKE ?", searchQuery),
		},
	).ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := GetDB().Query(sqlQuery, args...)
	return rows, err
}

func BuildFiQuery(userID int, categoryID int, c *fiber.Ctx) (string, []interface{}, error) {
	sq := squirrel.Select(
		"fi.id", "fi.source_id", "fi.title", "fi.url", "fi.description", "fi.author", "fi.published_at", "fi.score", "fi.comments_count", "fi.created_at", "fs.name as source_name",
	).From("feed_items fi").
		Join("feed_sources fs ON fi.source_id = fs.id").
		Join("user_category_feeds ucf ON fs.id = ucf.feed_source_id").
		Where(squirrel.Eq{"ucf.user_id": userID, "ucf.category_id": categoryID}).
		OrderBy("fi.published_at DESC").
		Limit(50)

	sql, args, err := sq.ToSql()
	if err != nil {
		return "", nil, c.Status(500).JSON(fiber.Map{
			"error": "Failed to build query",
		})
	}
	return sql, args, nil
}

func GraphFeedQuery(db *sql.DB, userID int, catID int) (*sql.Rows, error) {
	postRows, err := db.Query(`
				SELECT fi.id, fi.title
				FROM feed_items fi
				JOIN user_category_feeds ucf ON fi.source_id = ucf.feed_source_id
				WHERE ucf.user_id = ? AND ucf.category_id = ?
				LIMIT 50
			`, userID, catID)
	return postRows, err
}

var PostFeedQuery = `
SELECT fi.id, fi.source_id, fi.title, fi.url, fi.description, fi.author,
	   fi.published_at, fi.score, fi.comments_count, fi.created_at, fs.name as source_name
FROM feed_items fi
JOIN feed_sources fs ON fi.source_id = fs.id
WHERE fi.id = ?`

var FeedsQuery = `
SELECT fs.id, fs.name, fs.url, fs.last_updated
FROM feed_sources fs
JOIN user_category_feeds ucf ON fs.id = ucf.feed_source_id
WHERE ucf.user_id = ? AND ucf.category_id = ?
`

var SurpFeedsQuery = `
SELECT fs.id, fs.name, fs.url
FROM feed_sources fs
JOIN user_category_feeds ucf ON fs.id = ucf.feed_source_id
WHERE ucf.user_id = ? AND ucf.category_id = ?
`

var PostFeedNextQuery = `
SELECT fi.id, fi.source_id, fi.title, fi.url, fi.description, fi.author, fi.published_at, fi.score, fi.comments_count, fi.created_at, fs.name as source_name
FROM feed_items fi
JOIN feed_sources fs ON fi.source_id = fs.id
JOIN user_category_feeds ucf ON fs.id = ucf.feed_source_id
WHERE ucf.user_id = ? AND ucf.category_id = ?
ORDER BY fi.published_at DESC
LIMIT 50
`

var ResetQuery = `
UPDATE feed_sources
SET last_updated = datetime('2000-01-01 00:00:00')
WHERE id IN (
	SELECT feed_source_id
	FROM user_category_feeds
	WHERE user_id = ? AND category_id = ?
)
`

var FeedItemsQueryVariation = `SELECT fi.id, fi.source_id, fi.title, fi.url, fi.description, fi.author, fi.published_at, COALESCE(fi.score, 0) as score, COALESCE(fi.comments_count, 0) as comments_count, fi.created_at, fs.name as source_name
FROM feed_items fi
JOIN feed_sources fs ON fi.source_id = fs.id
WHERE fi.id = ?`
