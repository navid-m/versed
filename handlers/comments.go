package handlers

import (
	"strconv"

	"github.com/navid-m/versed/database"
	"github.com/navid-m/versed/feeds"

	"github.com/gofiber/fiber/v2"
)

// GetComments retrieves all comments for a specific feed item
func GetComments(c *fiber.Ctx) error {
	itemID := c.Params("itemId")

	if itemID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Item ID is required",
		})
	}

	comments, err := database.GetCommentsByItemID(itemID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get comments",
		})
	}

	return c.JSON(fiber.Map{
		"comments": comments,
		"count":    len(comments),
	})
}

// CreateComment adds a new comment to a feed item
func CreateComment(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	username, ok := c.Locals("username").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	itemID := c.Params("itemId")
	if itemID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Item ID is required",
		})
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Content == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Comment content is required",
		})
	}

	comment, err := database.CreateComment(itemID, userID, username, req.Content)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create comment",
		})
	}

	return c.Status(201).JSON(comment)
}

// UpdateComment updates an existing comment
func UpdateComment(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	commentIDStr := c.Params("commentId")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Content == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Comment content is required",
		})
	}

	// Check if the comment belongs to the user
	comment, err := database.GetCommentByID(commentID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Comment not found",
		})
	}

	if comment.UserID != userID {
		return c.Status(403).JSON(fiber.Map{
			"error": "You can only update your own comments",
		})
	}

	err = database.UpdateComment(commentID, req.Content)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update comment",
		})
	}

	// Return the updated comment
	updatedComment, err := database.GetCommentByID(commentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve updated comment",
		})
	}

	return c.JSON(updatedComment)
}

// DeleteComment removes a comment
func DeleteComment(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	commentIDStr := c.Params("commentId")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	// Check if the comment belongs to the user
	comment, err := database.GetCommentByID(commentID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Comment not found",
		})
	}

	if comment.UserID != userID {
		return c.Status(403).JSON(fiber.Map{
			"error": "You can only delete your own comments",
		})
	}

	err = database.DeleteComment(commentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete comment",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Comment deleted successfully",
	})
}

// GetPostView retrieves a single post with its comments for viewing
func GetPostView(c *fiber.Ctx) error {
	itemID := c.Params("itemId")
	if itemID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Item ID is required",
		})
	}

	// Get the feed item details
	db := database.GetDB()
	var item feeds.FeedItem
	var sourceName string

	query := `
		SELECT fi.id, fi.source_id, fi.title, fi.url, fi.description, fi.author,
			   fi.published_at, fi.score, fi.comments_count, fi.created_at, fs.name as source_name
		FROM feed_items fi
		JOIN feed_sources fs ON fi.source_id = fs.id
		WHERE fi.id = ?`

	err := db.QueryRow(query, itemID).Scan(
		&item.ID, &item.SourceID, &item.Title, &item.URL, &item.Description,
		&item.Author, &item.PublishedAt, &item.Score, &item.CommentsCount,
		&item.CreatedAt, &sourceName,
	)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Post not found",
		})
	}

	item.SourceName = sourceName

	// Get comments for this item
	comments, err := database.GetCommentsByItemID(itemID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get comments",
		})
	}

	return c.JSON(fiber.Map{
		"post":     item,
		"comments": comments,
	})
}
