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
