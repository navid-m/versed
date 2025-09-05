package feeds

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

type FeedSource struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	URL            string    `json:"url"`
	LastUpdated    time.Time `json:"last_updated"`
	UpdateInterval int       `json:"update_interval"`
}

type FeedItem struct {
	ID            string    `json:"id"`
	SourceID      int       `json:"source_id"`
	Title         string    `json:"title"`
	URL           string    `json:"url"`
	Description   string    `json:"description,omitempty"`
	Author        string    `json:"author,omitempty"`
	PublishedAt   time.Time `json:"published_at,omitempty"`
	Score         int       `json:"score,omitempty"`
	CommentsCount int       `json:"comments_count,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// FeedParser interface for different feed sources
type FeedParser interface {
	ParseFeed(content []byte) ([]FeedItem, error)
	GetFeedURL() string
	GetSourceName() string
}

// HTTPClient for fetching feeds
var client = &http.Client{
	Timeout: 30 * time.Second,
}

// FetchFeed fetches RSS content from URL
func FetchFeed(url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed returned status %d", resp.StatusCode)
	}

	// Read response body
	body := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	return body, nil
}

// ParseFeedWithParser parses RSS feed using gofeed
func ParseFeedWithParser(content []byte, sourceID int, sourceName string) ([]FeedItem, error) {
	parser := gofeed.NewParser()
	feed, err := parser.ParseString(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed: %w", err)
	}

	var items []FeedItem
	for _, item := range feed.Items {
		// Generate unique ID based on URL
		id := generateItemID(item.Link)

		// Parse published date
		var publishedAt time.Time
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		} else {
			publishedAt = time.Now()
		}

		feedItem := FeedItem{
			ID:            id,
			SourceID:      sourceID,
			Title:         item.Title,
			URL:           item.Link,
			Description:   item.Description,
			Author:        item.Author.Name,
			PublishedAt:   publishedAt,
			Score:         0, // Will be populated by specific parsers
			CommentsCount: 0, // Will be populated by specific parsers
			CreatedAt:     time.Now(),
		}
		items = append(items, feedItem)
	}

	return items, nil
}

// generateItemID creates a unique ID for feed items
func generateItemID(url string) string {
	hash := sha256.Sum256([]byte(url))
	return fmt.Sprintf("%x", hash)[:16]
}

// SaveFeedItems saves feed items to database
func SaveFeedItems(db *sql.DB, items []FeedItem) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO feed_items
		(id, source_id, title, url, description, author, published_at, score, comments_count, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		_, err = stmt.Exec(
			item.ID,
			item.SourceID,
			item.Title,
			item.URL,
			item.Description,
			item.Author,
			item.PublishedAt,
			item.Score,
			item.CommentsCount,
			item.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateFeedSourceTimestamp updates the last_updated timestamp for a feed source
func UpdateFeedSourceTimestamp(db *sql.DB, sourceID int) error {
	query := `UPDATE feed_sources SET last_updated = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.Exec(query, sourceID)
	return err
}

// GetFeedSourceByName gets a feed source by name
func GetFeedSourceByName(db *sql.DB, name string) (*FeedSource, error) {
	query := `SELECT id, name, url, last_updated, update_interval FROM feed_sources WHERE name = ?`
	row := db.QueryRow(query, name)

	var source FeedSource
	err := row.Scan(&source.ID, &source.Name, &source.URL, &source.LastUpdated, &source.UpdateInterval)
	if err != nil {
		return nil, err
	}
	return &source, nil
}

// CreateOrUpdateFeedSource creates or updates a feed source
func CreateOrUpdateFeedSource(db *sql.DB, name, url string) (*FeedSource, error) {
	// Try to get existing source
	src, err := GetFeedSourceByName(db, name)
	if err == nil {
		// Update existing
		query := `UPDATE feed_sources SET url = ?, last_updated = CURRENT_TIMESTAMP WHERE id = ?`
		_, err := db.Exec(query, url, src.ID)
		if err != nil {
			return nil, err
		}
		src.URL = url
		src.LastUpdated = time.Now()
		return src, nil
	}

	// Create new source
	query := `INSERT INTO feed_sources (name, url) VALUES (?, ?)`
	result, err := db.Exec(query, name, url)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	var source FeedSource
	source = FeedSource{
		ID:             int(id),
		Name:           name,
		URL:            url,
		LastUpdated:    time.Now(),
		UpdateInterval: 3600, // 1 hour default
	}
	return &source, nil
}

// GetAllFeedSources gets all feed sources
func GetAllFeedSources(db *sql.DB) ([]FeedSource, error) {
	query := `SELECT id, name, url, last_updated, update_interval FROM feed_sources`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []FeedSource
	for rows.Next() {
		var source FeedSource
		err := rows.Scan(&source.ID, &source.Name, &source.URL, &source.LastUpdated, &source.UpdateInterval)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, nil
}

// GetFeedItemsBySource gets feed items for a specific source
func GetFeedItemsBySource(db *sql.DB, sourceID int, limit int) ([]FeedItem, error) {
	query := `SELECT id, source_id, title, url, description, author, published_at, score, comments_count, created_at 
	          FROM feed_items 
	          WHERE source_id = ? 
	          ORDER BY published_at DESC 
	          LIMIT ?`
	rows, err := db.Query(query, sourceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []FeedItem
	for rows.Next() {
		var item FeedItem
		err := rows.Scan(&item.ID, &item.SourceID, &item.Title, &item.URL, &item.Description,
			&item.Author, &item.PublishedAt, &item.Score, &item.CommentsCount, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// GetAllFeedItems gets all feed items sorted by published date
func GetAllFeedItems(db *sql.DB, limit int) ([]FeedItem, error) {
	query := `SELECT fi.id, fi.source_id, fi.title, fi.url, fi.description, fi.author, fi.published_at, fi.score, fi.comments_count, fi.created_at, fs.name as source_name
	          FROM feed_items fi
	          JOIN feed_sources fs ON fi.source_id = fs.id
	          ORDER BY fi.published_at DESC
	          LIMIT ?`
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []FeedItem
	for rows.Next() {
		var item FeedItem
		var sourceName string
		err := rows.Scan(&item.ID, &item.SourceID, &item.Title, &item.URL, &item.Description,
			&item.Author, &item.PublishedAt, &item.Score, &item.CommentsCount, &item.CreatedAt, &sourceName)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// ShouldUpdateFeed checks if a feed should be updated based on its last update time
func ShouldUpdateFeed(source FeedSource) bool {
	return time.Since(source.LastUpdated) > time.Duration(source.UpdateInterval)*time.Second
}
