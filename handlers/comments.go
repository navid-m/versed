package handlers

import (
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/navid-m/versed/database"
	"github.com/navid-m/versed/feeds"
)

// Retrieves all comments for a specific feed item
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

// Adds a new comment to a feed item
func CreateComment(c *fiber.Ctx) error {
	userIDLocal := c.Locals("userID")
	if userIDLocal == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var userID int
	switch v := userIDLocal.(type) {
	case int:
		userID = v
	case int64:
		userID = int(v)
	case int32:
		userID = int(v)
	default:
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	username, ok := c.Locals("userUsername").(string)
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
		Content     string `json:"content"`
		ParentIDStr string `json:"parent_id,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Printf("=== CreateComment ERROR parsing request body: %v ===", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	log.Printf("=== CreateComment parsed request: Content='%s', ParentIDStr='%s' ===", req.Content, req.ParentIDStr)

	if req.Content == "" {
		log.Printf("=== CreateComment ERROR: Content is empty ===")
		return c.Status(400).JSON(fiber.Map{
			"error": "Comment content is required",
		})
	}

	urlPattern := `(https?:\/\/)?([\w-]+\.)+[\w-]+(\/[\w- .\/?%&=]*)?`
	regex, err := regexp.Compile(urlPattern)
	if err != nil {
		log.Printf("=== CreateComment ERROR compiling URL regex: %v ===", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	if regex.MatchString(req.Content) {
		log.Printf("=== CreateComment ERROR: Comment contains URL ===")
		return c.Status(400).JSON(fiber.Map{
			"error": "Links are not allowed in comments",
		})
	}

	var parentID *int
	if req.ParentIDStr != "" {
		if parsedID, err := strconv.Atoi(req.ParentIDStr); err != nil {
			log.Printf("=== CreateComment ERROR parsing parent_id '%s': %v ===", req.ParentIDStr, err)
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid parent_id format",
			})
		} else {
			parentID = &parsedID
			log.Printf("=== CreateComment parsed parent_id: %d ===", *parentID)
			if req.Content != "" {
				req.Content = fmt.Sprintf("@%d %s", *parentID, req.Content)
				log.Printf("=== CreateComment added @mention: '%s' ===", req.Content)
			}
		}
	}

	log.Printf("=== CreateComment calling database with parentID: %v ===", parentID)

	comment, err := database.CreateComment(itemID, userID, username, req.Content, parentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create comment",
		})
	}

	return c.Status(201).JSON(comment)
}

// Updates an existing comment
func UpdateComment(c *fiber.Ctx) error {
	userIDLocal := c.Locals("userID")
	if userIDLocal == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var userID int
	switch v := userIDLocal.(type) {
	case int:
		userID = v
	case int64:
		userID = int(v)
	case int32:
		userID = int(v)
	default:
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

	updatedComment, err := database.GetCommentByID(commentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve updated comment",
		})
	}

	return c.JSON(updatedComment)
}

// Removes a comment
func DeleteComment(c *fiber.Ctx) error {
	userIDLocal := c.Locals("userID")
	if userIDLocal == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var userID int
	switch v := userIDLocal.(type) {
	case int:
		userID = v
	case int64:
		userID = int(v)
	case int32:
		userID = int(v)
	default:
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

// Retrieves a single comment by its ID
func GetComment(c *fiber.Ctx) error {
	commentIDStr := c.Params("commentId")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	comment, err := database.GetCommentByID(commentID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Comment not found",
		})
	}

	return c.JSON(comment)
}

// Retrieves a single post with its comments for viewing
func GetPostView(c *fiber.Ctx) error {
	itemID := c.Params("itemId")
	if itemID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Item ID is required",
		})
	}

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
