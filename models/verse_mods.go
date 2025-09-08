package models

import "time"

// Subverse represents a sub-forum or category like /s/news
type Subverse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Post represents a user-created post in a subverse
type Post struct {
	ID         int       `json:"id"`
	SubverseID int       `json:"subverse_id"`
	UserID     int       `json:"user_id"`
	Username   string    `json:"username"`
	Title      string    `json:"title"`
	Content    string    `json:"content,omitempty"`
	PostType   string    `json:"post_type"` // "text" or "link"
	URL        string    `json:"url,omitempty"`
	Score      int       `json:"score"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// PostComment represents a comment on a post
type PostComment struct {
	ID        int          `json:"id"`
	PostID    int          `json:"post_id"`
	UserID    int          `json:"user_id"`
	Username  string       `json:"username"`
	Content   string       `json:"content"`
	ParentID  *int         `json:"parent_id,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Replies   []PostComment `json:"replies,omitempty"`
}
