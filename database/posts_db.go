package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/navid-m/versed/models"
)

// generateUUID generates a new UUID string
func generateUUID() string {
	return uuid.New().String()
}
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

	// Generate UUID for the post ID
	postID := generateUUID()

	query := `INSERT INTO posts (id, subverse_id, user_id, title, content, post_type, url, score, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`

	now := time.Now()
	_, err := db.Exec(query, postID, subverseID, userID, title, content, postType, url, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	post := &models.Post{
		ID:         postID,
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
func GetPostByID(db *sql.DB, postID string) (*models.Post, error) {
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

		post.Content = content.String
		post.URL = url.String

		posts = append(posts, post)
	}

	log.Printf("GetPostsBySubverse: returning %d posts", len(posts))

	return posts, nil
}

// UpdatePost updates a post's title and content
func UpdatePost(db *sql.DB, postID string, userID int, title, content string) error {
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
func DeletePost(db *sql.DB, postID string, userID int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM post_comments WHERE post_id = ?`, postID)
	if err != nil {
		return fmt.Errorf("failed to delete post comments: %w", err)
	}

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
func CreatePostComment(db *sql.DB, postID string, userID int, username, content string, parentID *string) (*models.PostComment, error) {
	// Generate UUID for the comment ID
	commentID := generateUUID()

	query := `INSERT INTO post_comments (id, post_id, user_id, username, content, parent_id, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := db.Exec(query, commentID, postID, userID, username, content, parentID, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	comment := &models.PostComment{
		ID:        commentID,
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
func GetPostComments(db *sql.DB, postID string) ([]models.PostComment, error) {
	query := `SELECT id, post_id, user_id, username, content, parent_id, created_at, updated_at
	          FROM post_comments
	          WHERE post_id = ?
	          ORDER BY created_at ASC`

	rows, err := db.Query(query, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	defer rows.Close()

	commentMap := make(map[string]*models.PostComment)

	for rows.Next() {
		var comment models.PostComment
		var parentID sql.NullString
		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &comment.Username,
			&comment.Content, &parentID, &comment.CreatedAt, &comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		if parentID.Valid {
			parentIDStr := parentID.String
			comment.ParentID = &parentIDStr
		}

		comment.Replies = []models.PostComment{}
		commentMap[comment.ID] = &comment
	}

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

	_, err = tx.Exec(`DELETE FROM post_comments WHERE parent_id = ?`, commentID)
	if err != nil {
		return fmt.Errorf("failed to delete comment replies: %w", err)
	}

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

// VoteOnPost creates or updates a vote on a post and updates the post score
func VoteOnPost(db *sql.DB, userID int, postID string, voteType string) error {
	if voteType != "upvote" && voteType != "downvote" {
		return fmt.Errorf("invalid vote type: must be 'upvote' or 'downvote'")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	var existingVoteType string
	var hasVote bool
	err = tx.QueryRow("SELECT vote_type FROM post_votes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingVoteType)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing vote: %w", err)
	}
	hasVote = err != sql.ErrNoRows

	now := time.Now()

	if hasVote {
		if existingVoteType == voteType {
			_, err = tx.Exec("DELETE FROM post_votes WHERE user_id = ? AND post_id = ?", userID, postID)
			if err != nil {
				return fmt.Errorf("failed to remove vote: %w", err)
			}
		} else {
			_, err = tx.Exec("UPDATE post_votes SET vote_type = ?, updated_at = ? WHERE user_id = ? AND post_id = ?", voteType, now, userID, postID)
			if err != nil {
				return fmt.Errorf("failed to update vote: %w", err)
			}
		}
	} else {
		_, err = tx.Exec("INSERT INTO post_votes (user_id, post_id, vote_type, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", userID, postID, voteType, now, now)
		if err != nil {
			return fmt.Errorf("failed to create vote: %w", err)
		}
	}

	err = updatePostScore(tx, postID)
	if err != nil {
		return fmt.Errorf("failed to update post score: %w", err)
	}

	return tx.Commit()
}

// GetUserVoteOnPost gets the user's current vote on a specific post (if any)
func GetUserVoteOnPost(db *sql.DB, userID, postID int) (string, error) {
	var voteType string
	err := db.QueryRow("SELECT vote_type FROM post_votes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&voteType)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to get user vote: %w", err)
	}
	return voteType, nil
}

// updatePostScore recalculates and updates the score for a post based on votes
func updatePostScore(tx *sql.Tx, postID string) error {
	var upvotes, downvotes int

	err := tx.QueryRow("SELECT COUNT(*) FROM post_votes WHERE post_id = ? AND vote_type = 'upvote'", postID).Scan(&upvotes)
	if err != nil {
		return fmt.Errorf("failed to count upvotes: %w", err)
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM post_votes WHERE post_id = ? AND vote_type = 'downvote'", postID).Scan(&downvotes)
	if err != nil {
		return fmt.Errorf("failed to count downvotes: %w", err)
	}

	score := upvotes - downvotes

	_, err = tx.Exec("UPDATE posts SET score = ? WHERE id = ?", score, postID)
	if err != nil {
		return fmt.Errorf("failed to update post score: %w", err)
	}

	return nil
}

// Searches for posts within a specific subverse
func SearchPostsBySubverse(db *sql.DB, subverseID int, query string, limit, offset int) ([]models.Post, error) {
	if strings.TrimSpace(query) == "" {
		return GetPostsBySubverse(db, subverseID, limit, offset)
	}

	searchQuery := `SELECT p.id, p.subverse_id, p.user_id, u.username, p.title, p.content, p.post_type, p.url, p.score, p.created_at, p.updated_at
	               FROM posts p
	               JOIN users u ON p.user_id = u.id
	               WHERE p.subverse_id = ? AND (p.title LIKE ? OR p.content LIKE ?)
	               ORDER BY p.created_at DESC
	               LIMIT ? OFFSET ?`

	searchPattern := "%" + query + "%"
	rows, err := db.Query(searchQuery, subverseID, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts: %w", err)
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

		post.Content = content.String
		post.URL = url.String

		posts = append(posts, post)
	}

	return posts, nil
}
