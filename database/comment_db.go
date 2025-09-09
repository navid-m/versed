package database

import (
	"log"
	"time"
)

// Represents a comment on a feed item
type Comment struct {
	ID        int       `json:"id"`
	ItemID    string    `json:"item_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	ParentID  *int      `json:"parent_id"`
	Replies   []Comment `json:"replies,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Adds a new comment to the database
//
// Connected to POST request
func CreateComment(itemID string, userID int, username, content string, parentID *int) (*Comment, error) {
	log.Printf("=== CreateComment called with parentID: %v ===", parentID)
	if parentID != nil {
		log.Printf("=== CreateComment parentID value: %d ===", *parentID)
	} else {
		log.Printf("=== CreateComment parentID is nil ===")
	}

	query := `
		INSERT INTO comments (item_id, user_id, username, content, parent_id)
		VALUES (?, ?, ?, ?, ?)`

	log.Printf("=== CreateComment executing query with params: itemID=%s, userID=%d, username=%s, content=%s, parentID=%v ===",
		itemID, userID, username, content, parentID)

	result, err := GetDB().Exec(query, itemID, userID, username, content, parentID)
	if err != nil {
		log.Printf("=== CreateComment ERROR executing query: %v ===", err)
		return nil, err
	}
	log.Printf("=== CreateComment query executed successfully ===")

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("=== CreateComment ERROR getting last insert ID: %v ===", err)
		return nil, err
	}
	log.Printf("=== CreateComment inserted comment with ID: %d ===", id)

	_, err = GetDB().Exec("UPDATE feed_items SET comments_count = comments_count + 1 WHERE id = ?", itemID)
	if err != nil {
		log.Printf("=== CreateComment ERROR updating comments_count: %v ===", err)
		return nil, err
	}

	comment, err := GetCommentByID(int(id))
	if err != nil {
		log.Printf("=== CreateComment ERROR getting created comment: %v ===", err)
		return nil, err
	}
	log.Printf("=== CreateComment returning comment with ParentID: %v ===", comment.ParentID)

	return comment, nil
}

// Retrieves all comments for a specific feed item
func GetCommentsByItemID(itemID string) ([]Comment, error) {
	log.Printf("=== GetCommentsByItemID called with itemID: %s ===", itemID)
	query := `
		SELECT id, item_id, user_id, username, content, parent_id, created_at, updated_at
		FROM comments
		WHERE item_id = ?
		ORDER BY created_at ASC`

	rows, err := GetDB().Query(query, itemID)
	if err != nil {
		log.Printf("=== GetCommentsByItemID ERROR querying database: %v ===", err)
		return nil, err
	}
	defer rows.Close()

	var allComments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(
			&comment.ID,
			&comment.ItemID,
			&comment.UserID,
			&comment.Username,
			&comment.Content,
			&comment.ParentID,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			log.Printf("=== GetCommentsByItemID ERROR scanning row: %v ===", err)
			return nil, err
		}
		allComments = append(allComments, comment)
	}

	// Build hierarchical structure
	log.Printf("=== BUILDING HIERARCHY: Processing %d comments ===", len(allComments))
	for i, comment := range allComments {
		log.Printf("=== HIERARCHY INPUT: Comment %d: ID=%d, Content='%s', ParentID=%v ===", i+1, comment.ID, comment.Content, comment.ParentID)
	}
	topLevelComments := buildCommentHierarchy(allComments)

	log.Printf("=== HIERARCHY RESULT: %d top-level comments ===", len(topLevelComments))
	for i, comment := range topLevelComments {
		log.Printf("=== HIERARCHY OUTPUT: Top-level %d: ID=%d, Content='%s', Replies=%d ===", i+1, comment.ID, comment.Content, len(comment.Replies))
		for j, reply := range comment.Replies {
			log.Printf("=== HIERARCHY OUTPUT:   Reply %d.%d: ID=%d, Content='%s' ===", i+1, j+1, reply.ID, reply.Content)
		}
	}

	log.Printf("=== GetCommentsByItemID returning %d top-level comments ===", len(topLevelComments))
	return topLevelComments, nil
}

// Retrieves a single comment by its ID
func GetCommentByID(commentID int) (*Comment, error) {
	query := `
		SELECT id, item_id, user_id, username, content, parent_id, created_at, updated_at
		FROM comments
		WHERE id = ?`

	var comment Comment
	err := GetDB().QueryRow(query, commentID).Scan(
		&comment.ID,
		&comment.ItemID,
		&comment.UserID,
		&comment.Username,
		&comment.Content,
		&comment.ParentID,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &comment, nil
}

// Updates the content of a comment
func UpdateComment(commentID int, content string) error {
	query := `
		UPDATE comments
		SET content = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	_, err := GetDB().Exec(query, content, commentID)
	return err
}

// Removes a comment from the database
func DeleteComment(commentID int) error {
	var itemID string
	err := GetDB().QueryRow("SELECT item_id FROM comments WHERE id = ?", commentID).Scan(&itemID)
	if err != nil {
		return err
	}

	query := `DELETE FROM comments WHERE id = ?`
	_, err = GetDB().Exec(query, commentID)
	if err != nil {
		return err
	}

	_, err = GetDB().Exec("UPDATE feed_items SET comments_count = comments_count - 1 WHERE id = ?", itemID)
	return err
}

// Returns the number of comments for a feed item
func GetCommentCountByItemID(itemID string) (int, error) {
	query := `SELECT COUNT(*) FROM comments WHERE item_id = ?`
	var count int
	err := GetDB().QueryRow(query, itemID).Scan(&count)
	return count, err
}

// Builds hierarchical comment structure from flat comments array
func buildCommentHierarchy(comments []Comment) []Comment {
	var (
		commentMap       = make(map[int]*Comment)
		replyMap         = make(map[int][]*Comment)
		topLevelComments []Comment
	)

	for i := range comments {
		comment := &comments[i]
		commentMap[comment.ID] = comment
		replyMap[comment.ID] = []*Comment{}
	}

	for i := range comments {
		comment := comments[i]

		if comment.ParentID == nil {
			topLevelComments = append(topLevelComments, comment)
		} else {
			parentID := *comment.ParentID
			replyMap[parentID] = append(replyMap[parentID], &comment)
		}
	}

	var buildReplies func(comment *Comment)
	buildReplies = func(comment *Comment) {
		replies := replyMap[comment.ID]
		comment.Replies = make([]Comment, len(replies))

		for i, reply := range replies {
			comment.Replies[i] = *reply
		}

		for i := range comment.Replies {
			buildReplies(&comment.Replies[i])
		}
	}

	for i := range topLevelComments {
		buildReplies(&topLevelComments[i])
	}

	return topLevelComments
}
