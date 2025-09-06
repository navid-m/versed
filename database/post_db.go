package database

import (
	"database/sql"
)

const ReadingListQuery = `SELECT fi.id, fi.source_id, fi.title, fi.url, fi.description, fi.author, fi.published_at, fi.score, fi.comments_count, fi.created_at, fs.name as source_name
FROM feed_items fi
JOIN reading_list rl ON fi.id = rl.item_id
JOIN feed_sources fs ON fi.s ource_id = fs.id
WHERE rl.user_id = ?
ORDER BY rl.created_at DESC`

func RetrieveReadingList(userID int) (*sql.Rows, error) {
	rows, err := GetDB().Query(ReadingListQuery, userID)
	return rows, err
}
