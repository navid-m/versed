package database

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
)

// Build reading list query using Squirrel
var ReadingListQueryBuilder = squirrel.Select(
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

// Retrieve some reading list.
func RetrieveReadingList(userID int) (*sql.Rows, error) {
	sqlQuery, args, err := ReadingListQueryBuilder.Where(
		squirrel.Eq{"rl.user_id": userID},
	).ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := GetDB().Query(sqlQuery, args...)
	return rows, err
}
