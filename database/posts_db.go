package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/navid-m/versed/models"
)

// CreatePost creates a new post in a subverse
func CreatePost(db *sql.DB, subverseID, userID int, username, title, content, postType, url string) (*models.Post, error) {
	if postType != "text" && postType != "link" {
		return nil, fmt.Errorf("invalid post type: must be 'text' or 'link'")
	}

	if postType == "link" && url == "" {
		return nil, fmt.Errorf("URL is required for link posts")
	}

	if postType == "text" && content == "" {
		return nil, fmt.Errorf("content is required for text posts")
	}

	query := `INSERT INTO posts (subverse_id, user_id, title, content, post_type, url, score, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?)`

	now := time.Now()
	result, err := db.Exec(query, subverseID, userID, title, content, postType, url, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get post ID: %w", err)
	}

	post := &models.Post{
		ID:         int(id),
		SubverseID: subverseID,
		UserID:     userID,
		Username:   username,
		Title:      title,
		Content:    content,
		PostType:   postType,
		URL:        url,
		Score:      0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	return post, nil
}

// GetPostByID retrieves a post by its ID
func GetPostByID(db *sql.DB, postID int) (*models.Post, error) {
	query := `SELECT p.id, p.subverse_id, p.user_id, u.username, p.title, p.content, p.post_type, p.url, p.score, p.created_at, p.updated_at
	          FROM posts p
	          JOIN users u ON p.user_id = u.id
	          WHERE p.id = ?`

	var post models.Post
	var content, url sql.NullString
	err := db.QueryRow(query, postID).Scan(
		&post.ID, &post.SubverseID, &post.UserID, &post.Username,
		&post.Title, &content, &post.PostType, &url, &post.Score,
		&post.CreatedAt, &post.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	if content.Valid {
		post.Content = content.String
	}
	if url.Valid {
		post.URL = url.String
	}

	return &post, nil
}

// GetPostsBySubverse retrieves posts for a specific subverse
func GetPostsBySubverse(db *sql.DB, subverseID int, limit, offset int) ([]models.Post, error) {
	query := `SELECT p.id, p.subverse_id, p.user_id, u.username, p.title, p.content, p.post_type, p.url, p.score, p.created_at, p.updated_at
	          FROM posts p
	          JOIN users u ON p.user_id = u.id
	          WHERE p.subverse_id = ?
	          ORDER BY p.created_at DESC
	          LIMIT ? OFFSET ?`

	rows, err := db.Query(query, subverseID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var content, url sql.NullString
		err := rows.Scan(
			&post.ID, &post.SubverseID, &post.UserID, &post.Username,
			&post.Title, &content, &post.PostType, &url, &post.Score,
			&post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		if content.Valid {
			post.Content = content.String
		}
		if url.Valid {
			post.URL = url.String
		}

		posts = append(posts, post)
	}

	return posts, nil
}

// UpdatePost updates a post's title and content
func UpdatePost(db *sql.DB, postID, userID int, title, content string) error {
	query := `UPDATE posts SET title = ?, content = ?, updated_at = ? WHERE id = ? AND user_id = ?`
	result, err := db.Exec(query, title, content, time.Now(), postID, userID)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post not found or user not authorized")
	}

	return nil
}

// DeletePost deletes a post and all its comments
func DeletePost(db *sql.DB, postID, userID int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete comments first (cascade should handle this, but let's be explicit)
	_, err = tx.Exec(`DELETE FROM post_comments WHERE post_id = ?`, postID)
	if err != nil {
		return fmt.Errorf("failed to delete post comments: %w", err)
	}

	// Delete the post
	result, err := tx.Exec(`DELETE FROM posts WHERE id = ? AND user_id = ?`, postID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post not found or user not authorized")
	}

	return tx.Commit()
}

// CreatePostComment creates a new comment on a post
func CreatePostComment(db *sql.DB, postID, userID int, username, content string, parentID *int) (*models.PostComment, error) {
	query := `INSERT INTO post_comments (post_id, user_id, username, content, parent_id, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	result, err := db.Exec(query, postID, userID, username, content, parentID, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get comment ID: %w", err)
	}

	comment := &models.PostComment{
		ID:        int(id),
		PostID:    postID,
		UserID:    userID,
		Username:  username,
		Content:   content,
		ParentID:  parentID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return comment, nil
}

// GetPostComments retrieves comments for a specific post
func GetPostComments(db *sql.DB, postID int) ([]models.PostComment, error) {
	query := `SELECT id, post_id, user_id, username, content, parent_id, created_at, updated_at
	          FROM post_comments
	          WHERE post_id = ?
	          ORDER BY created_at ASC`

	rows, err := db.Query(query, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	defer rows.Close()

	commentMap := make(map[int]*models.PostComment)

	for rows.Next() {
		var comment models.PostComment
		var parentID sql.NullInt64
		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &comment.Username,
			&comment.Content, &parentID, &comment.CreatedAt, &comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		if parentID.Valid {
			parentIDInt := int(parentID.Int64)
			comment.ParentID = &parentIDInt
		}

		comment.Replies = []models.PostComment{}
		commentMap[comment.ID] = &comment
	}

	// Organize into tree structure
	var rootComments []models.PostComment
	for _, comment := range commentMap {
		if comment.ParentID == nil {
			rootComments = append(rootComments, *comment)
		} else {
			if parent, exists := commentMap[*comment.ParentID]; exists {
				parent.Replies = append(parent.Replies, *comment)
			}
		}
	}

	return rootComments, nil
}

// UpdatePostComment updates a comment's content
func UpdatePostComment(db *sql.DB, commentID, userID int, content string) error {
	query := `UPDATE post_comments SET content = ?, updated_at = ? WHERE id = ? AND user_id = ?`
	result, err := db.Exec(query, content, time.Now(), commentID, userID)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("comment not found or user not authorized")
	}

	return nil
}

// DeletePostComment deletes a comment and all its replies
func DeletePostComment(db *sql.DB, commentID, userID int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete replies first
	_, err = tx.Exec(`DELETE FROM post_comments WHERE parent_id = ?`, commentID)
	if err != nil {
		return fmt.Errorf("failed to delete comment replies: %w", err)
	}

	// Delete the comment
	result, err := tx.Exec(`DELETE FROM post_comments WHERE id = ? AND user_id = ?`, commentID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("comment not found or user not authorized")
	}

	return tx.Commit()
}
