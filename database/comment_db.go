package database

import (
	"time"
)

// Comment represents a comment on a feed item
type Comment struct {
	ID        int       `json:"id"`
	ItemID    string    `json:"item_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateComment adds a new comment to the database
func CreateComment(itemID string, userID int, username, content string) (*Comment, error) {
	query := `
		INSERT INTO comments (item_id, user_id, username, content)
		VALUES (?, ?, ?, ?)`

	result, err := GetDB().Exec(query, itemID, userID, username, content)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Update comments_count in feed_items
	_, err = GetDB().Exec("UPDATE feed_items SET comments_count = comments_count + 1 WHERE id = ?", itemID)
	if err != nil {
		return nil, err
	}

	return GetCommentByID(int(id))
}

// GetCommentsByItemID retrieves all comments for a specific feed item
func GetCommentsByItemID(itemID string) ([]Comment, error) {
	query := `
		SELECT id, item_id, user_id, username, content, created_at, updated_at
		FROM comments
		WHERE item_id = ?
		ORDER BY created_at ASC`

	rows, err := GetDB().Query(query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(
			&comment.ID,
			&comment.ItemID,
			&comment.UserID,
			&comment.Username,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// GetCommentByID retrieves a single comment by its ID
func GetCommentByID(commentID int) (*Comment, error) {
	query := `
		SELECT id, item_id, user_id, username, content, created_at, updated_at
		FROM comments
		WHERE id = ?`

	var comment Comment
	err := GetDB().QueryRow(query, commentID).Scan(
		&comment.ID,
		&comment.ItemID,
		&comment.UserID,
		&comment.Username,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &comment, nil
}

// UpdateComment updates the content of a comment
func UpdateComment(commentID int, content string) error {
	query := `
		UPDATE comments
		SET content = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	_, err := GetDB().Exec(query, content, commentID)
	return err
}

// DeleteComment removes a comment from the database
func DeleteComment(commentID int) error {
	// First get the item_id to update the count
	var itemID string
	err := GetDB().QueryRow("SELECT item_id FROM comments WHERE id = ?", commentID).Scan(&itemID)
	if err != nil {
		return err
	}

	// Delete the comment
	query := `DELETE FROM comments WHERE id = ?`
	_, err = GetDB().Exec(query, commentID)
	if err != nil {
		return err
	}

	// Update comments_count in feed_items
	_, err = GetDB().Exec("UPDATE feed_items SET comments_count = comments_count - 1 WHERE id = ?", itemID)
	return err
}

// GetCommentCountByItemID returns the number of comments for a feed item
func GetCommentCountByItemID(itemID string) (int, error) {
	query := `SELECT COUNT(*) FROM comments WHERE item_id = ?`

	var count int
	err := GetDB().QueryRow(query, itemID).Scan(&count)
	return count, err
}
