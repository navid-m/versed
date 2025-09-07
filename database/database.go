package database

import (
	"database/sql"
)

var db *sql.DB

func InitDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		return err
	}
	return createTables()
}

func GetDB() *sql.DB {
	return db
}

func CloseConnection() error {
	return db.Close()
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS feed_sources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			url TEXT NOT NULL,
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_interval INTEGER DEFAULT 3600
		)`,
		`CREATE TABLE IF NOT EXISTS feed_items (
			id TEXT PRIMARY KEY,
			source_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			url TEXT NOT NULL,
			description TEXT,
			author TEXT,
			published_at DATETIME,
			score INTEGER DEFAULT 0,
			comments_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (source_id) REFERENCES feed_sources(id)
		)`,
		`CREATE TABLE IF NOT EXISTS upvotes (
			user_id INTEGER NOT NULL,
			item_id TEXT NOT NULL,
			vote_type TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, item_id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (item_id) REFERENCES feed_items(id)
		)`,
		`CREATE TABLE IF NOT EXISTS reading_list (
			user_id INTEGER NOT NULL,
			item_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, item_id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (item_id) REFERENCES feed_items(id)
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			session_id TEXT PRIMARY KEY,
			data TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}
