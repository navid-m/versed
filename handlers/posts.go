package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/navid-m/versed/database"
)

// CreatePost handles creating a new post in a subverse
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

	// Get subverse by name
	var subverseID int
	err := db.QueryRow("SELECT id FROM subverses WHERE name = ?", subverseName).Scan(&subverseID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Subverse not found",
		})
	}

	post, err := database.CreatePost(db, subverseID, userID.(int), username.(string), req.Title, req.Content, req.PostType, req.URL)
	if err != nil {
		log.Printf("Failed to create post: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create post",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(post)
}

// GetPost handles retrieving a single post
func GetPost(c *fiber.Ctx) error {
	postID, err := c.ParamsInt("postID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	db := database.GetDB()
	post, err := database.GetPostByID(db, postID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Post not found",
		})
	}

	return c.JSON(post)
}

// GetSubversePosts handles retrieving posts for a subverse
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

	// Get subverse by name
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
		"posts":  posts,
		"count":  len(posts),
		"limit":  limit,
		"offset": offset,
	})
}

// UpdatePost handles updating a post
func UpdatePost(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	postID, err := c.ParamsInt("postID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

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
	err = database.UpdatePost(db, postID, userID.(int), req.Title, req.Content)
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

// DeletePost handles deleting a post
func DeletePost(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	postID, err := c.ParamsInt("postID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	db := database.GetDB()
	err = database.DeletePost(db, postID, userID.(int))
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

// CreatePostComment handles creating a new comment on a post
func CreatePostComment(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	username := c.Locals("userUsername")
	if userID == nil || username == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	postID, err := c.ParamsInt("postID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	var req struct {
		Content  string `json:"content"`
		ParentID *int   `json:"parent_id,omitempty"`
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
	comment, err := database.CreatePostComment(db, postID, userID.(int), username.(string), req.Content, req.ParentID)
	if err != nil {
		log.Printf("Failed to create comment: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create comment",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(comment)
}

// GetPostComments handles retrieving comments for a post
func GetPostComments(c *fiber.Ctx) error {
	postID, err := c.ParamsInt("postID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

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

// UpdatePostComment handles updating a comment
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
	err = database.UpdatePostComment(db, commentID, userID.(int), req.Content)
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

// DeletePostComment handles deleting a comment
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
	err = database.DeletePostComment(db, commentID, userID.(int))
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
