package models

import "time"

// Subverse represents a sub-forum or category like /s/news
type Subverse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
