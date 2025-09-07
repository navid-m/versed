package database

import (
	"database/sql"
	"strings"

	"github.com/Masterminds/squirrel"
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
