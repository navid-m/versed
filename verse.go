package main

import (
	"database/sql"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/mustache/v2"
	_ "github.com/mattn/go-sqlite3"
)

// Represents some user of the system
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// The singleton database instance
//
// TODO: This will be changed to a connection pool
var db *sql.DB

func initDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "./users.db")
	if err != nil {
		return err
	}
	return createTable()
}

func createTable() error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	)`
	_, err := db.Exec(query)
	return err
}

func createUser(email, password string) error {
	query := `INSERT INTO users (email, password) VALUES (?, ?)`
	_, err := db.Exec(query, email, password)
	return err
}

func getUserByEmail(email string) (*User, error) {
	query := `SELECT id, email, password FROM users WHERE email = ?`
	row := db.QueryRow(query, email)

	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func main() {
	if err := initDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	var (
		engine = mustache.New("./views", ".mustache")
		app    = fiber.New(fiber.Config{
			Views: engine,
		})
	)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})
	app.Get("/signin", func(c *fiber.Ctx) error {
		return c.Render("signin", fiber.Map{})
	})
	app.Get("/signup", func(c *fiber.Ctx) error {
		return c.Render("signup", fiber.Map{})
	})

	app.Post("/signup", func(c *fiber.Ctx) error {
		email := c.FormValue("email")
		password := c.FormValue("password")

		if email == "" || password == "" {
			return c.Status(400).SendString("Email and password are required")
		}

		if err := createUser(email, password); err != nil {
			return c.Status(500).SendString("Failed to create user")
		}

		return c.Redirect("/")
	})

	app.Post("/signin", func(c *fiber.Ctx) error {
		email := c.FormValue("email")
		password := c.FormValue("password")

		if email == "" || password == "" {
			return c.Status(400).SendString("Email and password are required")
		}

		user, err := getUserByEmail(email)
		if err != nil || user.Password != password {
			return c.Status(401).SendString("Invalid credentials")
		}

		return c.Render("index", fiber.Map{
			"Email": user.Email,
		})
	})

	log.Fatal(app.Listen(":3000"))
}
