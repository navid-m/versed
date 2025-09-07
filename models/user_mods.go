package models

import "time"

// Represents some user of the system
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserCategory represents a user-created category for organizing feeds
type UserCategory struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserCategoryFeed represents the relationship between a user's category and a feed source
type UserCategoryFeed struct {
	UserID       int       `json:"user_id"`
	CategoryID   int       `json:"category_id"`
	FeedSourceID int       `json:"feed_source_id"`
	CreatedAt    time.Time `json:"created_at"`
}
