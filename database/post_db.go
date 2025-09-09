package database

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
)

// Build reading list query using Squirrel for feed items
var FeedReadingListQueryBuilder = squirrel.Select(
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
	Join("reading_list rl ON fi.id = rl.item_id").
	Join("feed_sources fs ON fi.source_id = fs.id").
	OrderBy("rl.created_at DESC")

// Build reading list query using Squirrel for posts
var PostReadingListQueryBuilder = squirrel.Select(
	"p.id",
	"'0' as source_id",
	"p.title",
	"p.url",
	"p.content as description",
	"u.username as author",
	"p.created_at as published_at",
	"p.score",
	"'0' as comments_count",
	"p.created_at",
	"'Subverse Post' as source_name",
).From("posts p").
	Join("reading_list rl ON CAST(p.id AS TEXT) = rl.item_id").
	Join("users u ON p.user_id = u.id").
	OrderBy("rl.created_at DESC")

// Retrieve feed items from reading list.
func RetrieveFeedReadingList(userID int) (*sql.Rows, error) {
	sqlQuery, args, err := FeedReadingListQueryBuilder.Where(
		squirrel.Eq{"rl.user_id": userID},
	).ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := GetDB().Query(sqlQuery, args...)
	return rows, err
}

// Retrieve posts from reading list.
func RetrievePostReadingList(userID int) (*sql.Rows, error) {
	sqlQuery, args, err := PostReadingListQueryBuilder.Where(
		squirrel.Eq{"rl.user_id": userID},
	).ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := GetDB().Query(sqlQuery, args...)
	return rows, err
}

// Retrieve combined reading list (feed items + posts).
func RetrieveReadingList(userID int) (*sql.Rows, error) {
	query := `
		SELECT fi.id, fi.source_id, fi.title, fi.url, fi.description, fi.author,
		       fi.published_at, fi.score,
		       COALESCE(feed_comment_counts.comment_count, 0) as comments_count,
		       fi.created_at, fs.name as source_name
		FROM feed_items fi
		JOIN reading_list rl ON fi.id = rl.item_id
		JOIN feed_sources fs ON fi.source_id = fs.id
		LEFT JOIN (
		    SELECT item_id, COUNT(*) as comment_count
		    FROM comments
		    GROUP BY item_id
		) feed_comment_counts ON fi.id = feed_comment_counts.item_id
		WHERE rl.user_id = $1

		UNION ALL

		SELECT p.id, '0' as source_id, p.title, p.url, p.content as description, u.username as author,
		       p.created_at as published_at, p.score,
		       COALESCE(post_comment_counts.comment_count, 0) as comments_count,
		       p.created_at, 'Subverse Post' as source_name
		FROM posts p
		JOIN reading_list rl ON CAST(p.id AS TEXT) = rl.item_id
		JOIN users u ON p.user_id = u.id
		LEFT JOIN (
		    SELECT post_id, COUNT(*) as comment_count
		    FROM post_comments
		    GROUP BY post_id
		) post_comment_counts ON CAST(p.id AS TEXT) = post_comment_counts.post_id
		WHERE rl.user_id = $1

		ORDER BY 10 DESC
	`

	rows, err := GetDB().Query(query, userID)
	return rows, err
}

var SubverseQuery = "SELECT id, name, created_at FROM subverses WHERE id = ?"
