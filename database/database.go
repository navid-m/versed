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
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
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
			username TEXT,
			password TEXT NOT NULL,
			ip_address TEXT,
			is_admin BOOLEAN DEFAULT 0
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
		`CREATE TABLE IF NOT EXISTS user_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, name),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_category_feeds (
			user_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			feed_source_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, category_id, feed_source_id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (category_id) REFERENCES user_categories(id),
			FOREIGN KEY (feed_source_id) REFERENCES feed_sources(id)
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			session_id TEXT PRIMARY KEY,
			data TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			item_id TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			username TEXT NOT NULL,
			content TEXT NOT NULL,
			parent_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (item_id) REFERENCES feed_items(id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS banned_ips (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ip_address TEXT NOT NULL UNIQUE,
			banned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			banned_by INTEGER NOT NULL,
			reason TEXT,
			is_active BOOLEAN DEFAULT 1,
			unbanned_at DATETIME,
			unbanned_by INTEGER,
			FOREIGN KEY (banned_by) REFERENCES users(id),
						FOREIGN KEY (unbanned_by) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS subverses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS subverse_feeds (
			subverse_id INTEGER NOT NULL,
			feed_source_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (subverse_id, feed_source_id),
			FOREIGN KEY (subverse_id) REFERENCES subverses(id),
			FOREIGN KEY (feed_source_id) REFERENCES feed_sources(id)
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
			id TEXT PRIMARY KEY,
			subverse_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT,
			post_type TEXT NOT NULL CHECK(post_type IN ('text', 'link')),
			url TEXT,
			score INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (subverse_id) REFERENCES subverses(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS post_votes (
			user_id INTEGER NOT NULL,
			post_id TEXT NOT NULL,
			vote_type TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, post_id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS post_comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			username TEXT NOT NULL,
			content TEXT NOT NULL,
			parent_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (parent_id) REFERENCES post_comments(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS hidden_posts (
			user_id INTEGER NOT NULL,
			item_id TEXT NOT NULL,
			hidden_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, item_id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (item_id) REFERENCES feed_items(id)
		)`,
	}

	// Check if ip_address column exists in users table and add it if not
	if err := ensureIPAddressColumn(); err != nil {
		return err
	}

	// Check if parent_id column exists in comments table and add it if not
	if err := ensureParentIDColumn(); err != nil {
		return err
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

// ensureIPAddressColumn checks if the ip_address column exists in the users table and adds it if not
func ensureIPAddressColumn() error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='ip_address'").Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		// Column does not exist, add it
		_, err = db.Exec("ALTER TABLE users ADD COLUMN ip_address TEXT")
		if err != nil {
			return err
		}
	}
	return nil
}

// ensureParentIDColumn checks if the parent_id column exists in the comments table and adds it if not
func ensureParentIDColumn() error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('comments') WHERE name='parent_id'").Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		// Column does not exist, add it
		_, err = db.Exec("ALTER TABLE comments ADD COLUMN parent_id INTEGER")
		if err != nil {
			return err
		}
	}
	return nil
}
