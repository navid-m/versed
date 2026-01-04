package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	log.Println("Starting migration: Converting post IDs from INTEGER to TEXT (UUID)")

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	log.Println("Step 1: Creating new posts table with UUID support...")
	_, err = tx.Exec(`
		CREATE TABLE posts_new (
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
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create new posts table: %v", err)
	}

	log.Println("Step 2: Creating ID mapping table...")
	type IDMapping struct {
		OldID int
		NewID string
	}
	mappings := make(map[int]string)

	log.Println("Step 3: Copying posts with new UUIDs...")
	rows, err := tx.Query(`SELECT id, subverse_id, user_id, title, content, post_type, url, score, created_at, updated_at FROM posts`)
	if err != nil {
		log.Fatalf("Failed to query posts: %v", err)
	}
	defer rows.Close()

	postCount := 0
	for rows.Next() {
		var oldID, subverseID, userID, score int
		var title, postType string
		var content, url, createdAt, updatedAt sql.NullString

		err := rows.Scan(&oldID, &subverseID, &userID, &title, &content, &postType, &url, &score, &createdAt, &updatedAt)
		if err != nil {
			log.Fatalf("Failed to scan post: %v", err)
		}

		newID := uuid.New().String()
		mappings[oldID] = newID

		_, err = tx.Exec(`
			INSERT INTO posts_new (id, subverse_id, user_id, title, content, post_type, url, score, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			newID, subverseID, userID, title, content, postType, url, score, createdAt, updatedAt)
		if err != nil {
			log.Fatalf("Failed to insert post with new UUID: %v", err)
		}

		postCount++
		if postCount%100 == 0 {
			log.Printf("Migrated %d posts...", postCount)
		}
	}

	log.Printf("Migrated %d posts total", postCount)
	log.Println("Step 4: Updating post_votes with new UUIDs...")

	voteCount := 0
	for oldID, newID := range mappings {
		result, err := tx.Exec(`UPDATE post_votes SET post_id = ? WHERE post_id = ?`, newID, fmt.Sprintf("%d", oldID))
		if err != nil {
			log.Fatalf("Failed to update post_votes: %v", err)
		}
		rows, _ := result.RowsAffected()
		voteCount += int(rows)
	}
	log.Printf("Updated %d vote records", voteCount)

	log.Println("Step 5: Updating post_comments with new UUIDs...")
	commentCount := 0
	for oldID, newID := range mappings {
		result, err := tx.Exec(`UPDATE post_comments SET post_id = ? WHERE post_id = ?`, newID, fmt.Sprintf("%d", oldID))
		if err != nil {
			log.Fatalf("Failed to update post_comments: %v", err)
		}
		rows, _ := result.RowsAffected()
		commentCount += int(rows)
	}
	log.Printf("Updated %d comment records", commentCount)

	log.Println("Step 6: Replacing old posts table...")
	_, err = tx.Exec(`DROP TABLE posts`)
	if err != nil {
		log.Fatalf("Failed to drop old posts table: %v", err)
	}

	_, err = tx.Exec(`ALTER TABLE posts_new RENAME TO posts`)
	if err != nil {
		log.Fatalf("Failed to rename new posts table: %v", err)
	}

	log.Println("Step 7: Recreating indexes...")
	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_posts_subverse ON posts(subverse_id)`)
	if err != nil {
		log.Printf("Warning: Failed to create subverse index: %v", err)
	}

	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_posts_user ON posts(user_id)`)
	if err != nil {
		log.Printf("Warning: Failed to create user index: %v", err)
	}

	log.Println("Committing changes...")
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Println("✓ Migration completed successfully!")
	log.Printf("Summary: Migrated %d posts, %d votes, %d comments", postCount, voteCount, commentCount)
	log.Println("Post IDs have been converted from INTEGER to TEXT (UUID)")
}
