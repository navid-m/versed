package models

import "time"

// Represents some user of the system
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

// BannedIP represents a banned IP address
type BannedIP struct {
	ID         int        `json:"id"`
	IPAddress  string     `json:"ip_address"`
	BannedAt   time.Time  `json:"banned_at"`
	BannedBy   int        `json:"banned_by"`
	Reason     string     `json:"reason,omitempty"`
	IsActive   bool       `json:"is_active"`
	UnbannedAt *time.Time `json:"unbanned_at,omitempty"`
	UnbannedBy *int       `json:"unbanned_by,omitempty"`
}

// AdminUser represents an admin user for management
type AdminUser struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}
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

// Subverse represents a sub-forum or category like /s/news
type Subverse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
