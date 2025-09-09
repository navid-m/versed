package handlers

import (
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/django/v3"
	"github.com/navid-m/versed/database"
	"github.com/navid-m/versed/feeds"
	"github.com/navid-m/versed/models"
)

func SignUpHandler(c *fiber.Ctx, store *session.Store) error {
	email := c.FormValue("email")
	username := c.FormValue("username")
	password := c.FormValue("password")
	if email == "" || username == "" || password == "" {
		return c.Status(400).SendString("Email, username, and password are required")
	}

	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return c.Status(400).SendString("Enter a valid email address")
	}

	if len(username) < 3 || len(username) > 20 {
		return c.Status(400).SendString("Username must be between 3 and 20 characters")
	}
	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' {
			return c.Status(400).SendString("Username can only contain letters, numbers, and underscores")
		}
	}

	if len(password) < 3 {
		return c.Status(400).SendString("Password must be at least 3 characters long")
	}

	err := database.CreateUser(email, username, password, c.IP())
	if err != nil {
		switch err.Error() {
		case "email already in use":
			return c.Status(400).SendString("This email is already registered")
		case "username already taken":
			return c.Status(400).SendString("This username is already taken")
		default:
			log.Printf("Error creating user: %v", err)
			return c.Status(500).SendString("An error occurred while creating your account")
		}
	}

	user, err := database.GetUserByEmail(email)
	if err != nil {
		log.Printf("Error getting new user after signup: %v", err)
		return c.Redirect("/signin")
	}

	sess, err := store.Get(c)
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return c.Redirect("/signin")
	}

	sess.Set("user_id", user.ID)
	sess.Set("user_email", user.Email)
	sess.Set("user_username", user.Username)

	if err := sess.Save(); err != nil {
		log.Printf("Error saving session: %v", err)
		return c.Redirect("/signin")
	}

	return c.Redirect("/")
}

func AboutHandler(c *fiber.Ctx) error {
	userEmail := c.Locals("userEmail")
	userUsername := c.Locals("userUsername")

	data := fiber.Map{}
	if userEmail != nil {
		data["Email"] = userEmail
	}
	if userUsername != nil {
		data["Username"] = userUsername
	}

	return c.Render("about", data)
}

func IndexHandler(c *fiber.Ctx) error {
	userEmail := c.Locals("userEmail")
	userUsername := c.Locals("userUsername")
	userID := c.Locals("userID")

	var feedItems []feeds.FeedItem
	var err error

	if userID != nil {
		feedItems, err = feeds.GetAllFeedItemsForUser(database.GetDB(), userID.(int), 20)
	} else {
		feedItems, err = feeds.GetAllFeedItems(database.GetDB(), 20)
	}

	if err != nil {
		log.Printf("Failed to get feed items: %v", err)
	}

	for i, f := range feedItems {
		if strings.TrimSpace(f.Description) == "" {
			feedItems[i].Description = "No description."
		}
		if strings.TrimSpace(f.Description) == "Comments" {
			feedItems[i].Description = "No description."
		}
	}

	data := fiber.Map{
		"FeedItems": feedItems,
	}

	if userEmail != nil {
		data["Email"] = userEmail
	}
	if userUsername != nil {
		data["Username"] = userUsername
	}
	err = django.New("./views", ".html").Render(c, "index", data)
	fmt.Println(err)
	return c.Render("index", data)
}

func SignOutHandler(store *session.Store, c *fiber.Ctx) error {
	sess, err := store.Get(c)
	if err == nil {
		sess.Delete("user_id")
		sess.Delete("user_email")
		sess.Delete("user_username")
		sess.Save()
	}
	return c.Redirect("/")
}

func SignInHandler(c *fiber.Ctx, store *session.Store) error {
	var (
		email    = c.FormValue("email")
		password = c.FormValue("password")
	)
	if email == "" || password == "" {
		return c.Status(400).JSON(fiber.Map{
			"toast": fiber.Map{
				"type":    "error",
				"message": "Email and password are required",
			},
		})
	}
	user, err := database.GetUserByEmail(email)
	if err != nil {
		log.Printf("Login failed for email %s: %v", email, err)
		return c.Status(401).JSON(fiber.Map{
			"toast": fiber.Map{
				"type":    "error",
				"message": "Invalid email or password",
			},
		})
	}
	err = database.VerifyPassword(user.Password, password)
	if err != nil {
		log.Printf("Invalid password attempt for user %s", email)
		return c.Status(401).JSON(fiber.Map{
			"toast": fiber.Map{
				"type":    "error",
				"message": "Invalid email or password",
			},
		})
	}

	sess, err := store.Get(c)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"toast": fiber.Map{
				"type":    "error",
				"message": "Session error",
			},
		})
	}
	sess.Set("user_id", user.ID)
	sess.Set("user_email", user.Email)
	sess.Set("user_username", user.Username)
	if err := sess.Save(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"toast": fiber.Map{
				"type":    "error",
				"message": "Failed to save session",
			},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

func GraphHandler(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	log.Printf("Graph API called for user %d", userID)

	db := database.GetDB()
	categoryRows, err := db.Query(database.CategoryQuery, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get categories",
		})
	}
	defer categoryRows.Close()

	var nodes []fiber.Map
	var links []fiber.Map
	categoryMap := make(map[int]string)

	nodes = append(nodes, fiber.Map{
		"id":   "root",
		"name": "Categories",
		"type": "root",
	})

	for categoryRows.Next() {
		var catID int
		var catName string
		err := categoryRows.Scan(&catID, &catName)
		if err != nil {
			continue
		}

		nodes = append(nodes, fiber.Map{
			"id":   fmt.Sprintf("cat_%d", catID),
			"name": catName,
			"type": "category",
		})
		categoryMap[catID] = catName

		links = append(links, fiber.Map{
			"source": "root",
			"target": fmt.Sprintf("cat_%d", catID),
		})

		postRows, err := database.GraphFeedQuery(db, userID, catID)
		if err != nil {
			log.Printf("Error querying posts for category %d: %v", catID, err)
			continue
		}

		postCount := 0
		for postRows.Next() {
			var postID string
			var postTitle string
			err := postRows.Scan(&postID, &postTitle)
			if err != nil {
				log.Printf("Error scanning post row: %v", err)
				continue
			}
			postCount++

			nodes = append(nodes, fiber.Map{
				"id":   fmt.Sprintf("post_%s", postID),
				"name": postTitle,
				"type": "post",
			})

			links = append(links, fiber.Map{
				"source": fmt.Sprintf("cat_%d", catID),
				"target": fmt.Sprintf("post_%s", postID),
			})
		}
		postRows.Close()
		log.Printf("Category %d (%s): %d posts", catID, catName, postCount)
	}

	log.Printf("Graph API returning %d nodes, %d links", len(nodes), len(links))

	return c.JSON(fiber.Map{
		"nodes": nodes,
		"links": links,
	})
}

func PostItemHandler(c *fiber.Ctx) error {
	itemID := c.Params("itemId")
	userEmail := c.Locals("userEmail")
	userUsername := c.Locals("userUsername")
	userID := c.Locals("userID")

	db := database.GetDB()
	var post feeds.FeedItem
	var sourceName string

	err := db.QueryRow(database.PostFeedQuery, itemID).Scan(
		&post.ID, &post.SourceID, &post.Title, &post.URL, &post.Description,
		&post.Author, &post.PublishedAt, &post.Score, &post.CommentsCount,
		&post.CreatedAt, &sourceName,
	)

	if err != nil {
		postObj, postErr := database.GetPostByID(db, itemID)
		if postErr != nil {
			return c.Status(404).SendString("Post not found")
		}

		var subverse models.Subverse
		subverseErr := db.QueryRow(database.SubverseQuery, postObj.SubverseID).Scan(
			&subverse.ID, &subverse.Name, &subverse.CreatedAt)
		if subverseErr != nil {
			return c.Status(500).SendString("Failed to get subverse information")
		}

		comments, commentErr := database.GetCommentsByItemID(itemID)
		if commentErr != nil {
			log.Printf("Failed to get comments: %v", commentErr)
			comments = []database.Comment{}
		}

		data := fiber.Map{
			"Post":          postObj,
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

	post.SourceName = sourceName

	comments, err := database.GetCommentsByItemID(itemID)
	fmt.Println(comments)
	log.Printf("=== POST PAGE DEBUG: Retrieved %d comments for item %s ===", len(comments), itemID)
	for i, comment := range comments {
		log.Printf("=== POST PAGE DEBUG: Comment %d: ID=%d, Content='%s', ParentID=%v, Replies=%d ===",
			i+1, comment.ID, comment.Content, comment.ParentID, len(comment.Replies))
		for j, reply := range comment.Replies {
			log.Printf("=== POST PAGE DEBUG:   Reply %d.%d: ID=%d, Content='%s' ===",
				i+1, j+1, reply.ID, reply.Content)
		}
	}

	if err != nil {
		log.Printf("Failed to get comments: %v", err)
		comments = []database.Comment{}
	}

	data := fiber.Map{
		"post":     post,
		"comments": comments,
	}

	if userEmail != nil {
		data["Email"] = userEmail
	}
	if userUsername != nil {
		data["Username"] = userUsername
	}
	if userID != nil {
		data["userID"] = userID
	}
	return c.Render("post", data)
}
