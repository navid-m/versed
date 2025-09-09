package handlers

import (
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/navid-m/versed/database"
	"github.com/navid-m/versed/models"
)

// Handles creating a new post in a subverse
func CreatePost(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	username := c.Locals("userUsername")
	if userID == nil || username == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	subverseName := c.Params("subverseName")
	if subverseName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Subverse name is required",
		})
	}

	var req struct {
		Title    string `json:"title"`
		Content  string `json:"content,omitempty"`
		PostType string `json:"post_type"`
		URL      string `json:"url,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title is required",
		})
	}

	if req.PostType != "text" && req.PostType != "link" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Post type must be 'text' or 'link'",
		})
	}

	if req.PostType == "link" && req.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "URL is required for link posts",
		})
	}

	if req.PostType == "text" && req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Content is required for text posts",
		})
	}

	db := database.GetDB()

	var subverseID int
	err := db.QueryRow("SELECT id FROM subverses WHERE name = ?", subverseName).Scan(&subverseID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Subverse not found",
		})
	}

	log.Printf("Creating post in subverse '%s' (ID: %d)", subverseName, subverseID)
	log.Printf("Getting post with ID: %s", req.Title)
	post, err := database.CreatePost(db, subverseID, userID.(int), username.(string), req.Title, req.Content, req.PostType, req.URL)
	if err != nil {
		log.Printf("Failed to create post: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create post",
		})
	}
	log.Printf("Successfully created post with ID %d", post.ID)

	err = database.UpdateSubversePostCount(db, subverseID)
	if err != nil {
		log.Printf("Failed to update subverse post count: %v", err)
	}

	return c.Status(fiber.StatusCreated).JSON(post)
}

// Handles retrieving a single post
func GetPost(c *fiber.Ctx) error {
	postIDStr := c.Params("postID")
	if postIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	postID := strings.TrimSpace(postIDStr)

	db := database.GetDB()
	post, err := database.GetPostByID(db, postID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Post not found",
		})
	}

	var subverse models.Subverse
	err = db.QueryRow("SELECT id, name, created_at FROM subverses WHERE id = ?", post.SubverseID).Scan(
		&subverse.ID, &subverse.Name, &subverse.CreatedAt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get subverse information",
		})
	}

	comments, err := database.GetPostComments(db, postID)
	if err != nil {
		log.Printf("Failed to get comments: %v", err)
		comments = []models.PostComment{}
	}
	userEmail := c.Locals("userEmail")
	userUsername := c.Locals("userUsername")
	userID := c.Locals("userID")

	data := fiber.Map{
		"Post":          post,
		"Subverse":      subverse,
		"Comments":      comments,
		"CommentsCount": len(comments),
	}

	if userEmail != nil {
		data["Email"] = userEmail
	}
	if userUsername != nil {
		data["Username"] = userUsername
	}
	if userID != nil {
		data["UserID"] = userID
	}

	return c.Render("subverse-post", data)
}

// Handles retrieving posts for a subverse
func GetSubversePosts(c *fiber.Ctx) error {
	subverseName := c.Params("subverseName")
	if subverseName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Subverse name is required",
		})
	}

	limit := c.QueryInt("limit", 20)
	if limit > 50 {
		limit = 50
	}
	offset := c.QueryInt("offset", 0)

	db := database.GetDB()

	var subverseID int
	err := db.QueryRow("SELECT id FROM subverses WHERE name = ?", subverseName).Scan(&subverseID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Subverse not found",
		})
	}

	posts, err := database.GetPostsBySubverse(db, subverseID, limit, offset)
	if err != nil {
		log.Printf("Failed to get subverse posts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get posts",
		})
	}

	return c.JSON(fiber.Map{
		"Posts":  posts,
		"count":  len(posts),
		"limit":  limit,
		"offset": offset,
	})
}

// Handles updating a post
func UpdatePost(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	postIDStr := c.Params("postID")
	if postIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}
	postID := strings.TrimSpace(postIDStr)

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title is required",
		})
	}

	db := database.GetDB()
	err := database.UpdatePost(db, postID, userID.(int), req.Title, req.Content)
	if err != nil {
		log.Printf("Failed to update post: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update post",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Post updated successfully",
	})
}

// Handles deleting a post
func DeletePost(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	postIDStr := c.Params("postID")
	if postIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	postID := strings.TrimSpace(postIDStr)

	db := database.GetDB()
	err := database.DeletePost(db, postID, userID.(int))
	if err != nil {
		log.Printf("Failed to delete post: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete post",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Post deleted successfully",
	})
}

// Handles creating a new comment on a post
func CreatePostComment(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	username := c.Locals("userUsername")
	if userID == nil || username == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	postIDStr := c.Params("postID")
	if postIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	postID := strings.TrimSpace(postIDStr)

	var req struct {
		Content  string  `json:"content"`
		ParentID *string `json:"parent_id,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Printf("BodyParser error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	log.Printf("Received comment request: content='%s', parent_id=%v", req.Content, req.ParentID)

	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Content is required",
		})
	}

	db := database.GetDB()
	comment, err := database.CreatePostComment(db, postID, userID.(int), username.(string), req.Content, req.ParentID)
	if err != nil {
		log.Printf("Failed to create comment: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create comment",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(comment)
}

// Handles retrieving comments for a post
func GetPostComments(c *fiber.Ctx) error {
	postIDStr := c.Params("postID")
	if postIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	postID := strings.TrimSpace(postIDStr)

	db := database.GetDB()
	comments, err := database.GetPostComments(db, postID)
	if err != nil {
		log.Printf("Failed to get post comments: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get comments",
		})
	}

	return c.JSON(fiber.Map{
		"comments": comments,
		"count":    len(comments),
	})
}

// Handles updating a comment
func UpdatePostComment(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	commentID, err := c.ParamsInt("commentID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Content is required",
		})
	}

	db := database.GetDB()
	err = database.UpdatePostComment(db, strconv.Itoa(commentID), userID.(int), req.Content)
	if err != nil {
		log.Printf("Failed to update comment: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update comment",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Comment updated successfully",
	})
}

// Handles deleting a comment
func DeletePostComment(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	commentID, err := c.ParamsInt("commentID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	db := database.GetDB()
	err = database.DeletePostComment(db, strconv.Itoa(commentID), userID.(int))
	if err != nil {
		log.Printf("Failed to delete comment: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete comment",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Comment deleted successfully",
	})
}

// VotePost handles voting on a post (upvote/downvote)
func VotePost(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	postIDStr := c.Params("postID")
	if postIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	postID := strings.TrimSpace(postIDStr)

	var req struct {
		VoteType string `json:"vote_type"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.VoteType != "upvote" && req.VoteType != "downvote" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Vote type must be 'upvote' or 'downvote'",
		})
	}

	db := database.GetDB()

	err := database.VoteOnPost(db, userID.(int), postID, req.VoteType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to vote on post",
		})
	}

	post, err := database.GetPostByID(db, postID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get updated post",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"post_id": postID,
		"score":   post.Score,
	})
}

// Handles searching for posts within a subverse
func SearchPosts(c *fiber.Ctx) error {
	subverseName := c.Params("subverseName")
	if subverseName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Subverse name is required",
		})
	}

	var (
		subverseID int
		query      = c.Query("q", "")
		limit      = min(c.QueryInt("limit", 20), 50)
		offset     = c.QueryInt("offset", 0)
		db         = database.GetDB()
		err        = db.QueryRow("SELECT id FROM subverses WHERE name = ?", subverseName).Scan(&subverseID)
	)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Subverse not found",
		})
	}

	posts, err := database.SearchPostsBySubverse(db, subverseID, query, limit, offset)
	if err != nil {
		log.Printf("Failed to search posts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search posts",
		})
	}

	return c.JSON(fiber.Map{
		"Posts":  posts,
		"count":  len(posts),
		"limit":  limit,
		"offset": offset,
		"query":  query,
	})
}
