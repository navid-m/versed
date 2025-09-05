package feeds

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"log"
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

// Interface for different feed sources.
type FeedParser interface {
	ParseFeed(content []byte) ([]FeedItem, error)
	GetFeedURL() string
	GetSourceName() string
}

// HTTPClient for fetching feeds
var client = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

func ResetAllFeedTimestamps(db *sql.DB) error {
	query := `UPDATE feed_sources SET last_updated = datetime('2000-01-01 00:00:00')`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to reset feed timestamps: %w", err)
	}
	log.Println("Reset all feed timestamps to force updates")
	return nil
}

// Also add this simplified test to see what's in the database:
func DebugFeeds(db *sql.DB) {
	log.Println("=== DATABASE DEBUG ===")

	// Check feed sources
	sources, err := GetAllFeedSources(db)
	if err != nil {
		log.Printf("ERROR getting feed sources: %v", err)
		return
	}
	log.Printf("Feed sources in database: %d", len(sources))
	for _, source := range sources {
		log.Printf("  - %s (ID: %d, URL: %s, LastUpdated: %v)",
			source.Name, source.ID, source.URL, source.LastUpdated)
	}

	// Check feed items
	items, err := GetAllFeedItems(db, 10)
	if err != nil {
		log.Printf("ERROR getting feed items: %v", err)
		return
	}
	log.Printf("Feed items in database: %d", len(items))
	for _, item := range items {
		log.Printf("  - %s (Source: %d)", item.Title, item.SourceID)
	}
}

// Fetches RSS content from URL.
func FetchFeed(url string) ([]byte, error) {
	fmt.Println("REQUEST MADE!")
	log.Printf("Making HTTP request to: %s", url)
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("HTTP request failed for %s: %v", url, err)
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("HTTP response status: %d for URL: %s", resp.StatusCode, url)
	if resp.StatusCode != http.StatusOK {
		log.Printf("HTTP error status %d for URL: %s", resp.StatusCode, url)
		return nil, fmt.Errorf("feed returned status %d", resp.StatusCode)
	}

	body := make([]byte, 0)
	buf := make([]byte, 1024)
	totalBytes := 0
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
			totalBytes += n
		}
		if err != nil {
			if err == io.EOF {
				log.Printf("Successfully read %d bytes from %s", totalBytes, url)
				break
			}
			log.Printf("Error reading response body from %s: %v", url, err)
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
	}

	return body, nil
}

// Parses RSS feed using gofeed.
func ParseFeedWithParser(content []byte, sourceID int, sourceName string) ([]FeedItem, error) {
	parser := gofeed.NewParser()
	feed, err := parser.ParseString(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed: %w", err)
	}

	var items []FeedItem
	for _, item := range feed.Items {
		id := generateItemID(item.Link)
		var publishedAt time.Time
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		} else {
			publishedAt = time.Now()
		}
		var authorName string
		if item.Author != nil {
			authorName = item.Author.Name
		}
		feedItem := FeedItem{
			ID:            id,
			SourceID:      sourceID,
			Title:         item.Title,
			URL:           item.Link,
			Description:   item.Description,
			Author:        authorName,
			PublishedAt:   publishedAt,
			Score:         0,
			CommentsCount: 0,
			CreatedAt:     time.Now(),
		}
		items = append(items, feedItem)
	}

	return items, nil
}

// Creates a unique ID for feed items.
func generateItemID(url string) string {
	hash := sha256.Sum256([]byte(url))
	return fmt.Sprintf("%x", hash)[:16]
}

// Saves feed items to database.
func SaveFeedItems(db *sql.DB, items []FeedItem) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
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

// Updates the last_updated timestamp for a feed source
func UpdateFeedSourceTimestamp(db *sql.DB, sourceID int) error {
	query := `UPDATE feed_sources SET last_updated = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.Exec(query, sourceID)
	return err
}

// Gets a feed source by name.
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

// Creates or updates a feed source - FIXED to preserve timestamps
func CreateOrUpdateFeedSource(db *sql.DB, name, url string) (*FeedSource, error) {
	// Try to get existing source
	existing, err := GetFeedSourceByName(db, name)
	if err == nil {
		// Source exists - return it WITHOUT any database updates
		log.Printf("Found existing source: %s (ID: %d, LastUpdated: %v)",
			name, existing.ID, existing.LastUpdated)
		return existing, nil
	}

	// Source doesn't exist, create new one
	log.Printf("Creating new feed source: %s", name)
	query := `INSERT INTO feed_sources (name, url, last_updated, update_interval) 
	          VALUES (?, ?, datetime('2000-01-01 00:00:00'), 3600)`
	result, err := db.Exec(query, name, url)
	if err != nil {
		return nil, fmt.Errorf("failed to insert feed source: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	source := &FeedSource{
		ID:             int(id),
		Name:           name,
		URL:            url,
		LastUpdated:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdateInterval: 3600,
	}

	log.Printf("Created new feed source with ID: %d", source.ID)
	return source, nil
}

// TEMPORARY DEBUG - Replace your ShouldUpdateFeed function with this one that forces updates
func ShouldUpdateFeed(source FeedSource) bool {
	timeSince := time.Since(source.LastUpdated)
	threshold := time.Duration(source.UpdateInterval) * time.Second

	log.Printf("DEBUG ShouldUpdateFeed - Source: %s", source.Name)
	log.Printf("  LastUpdated: %v", source.LastUpdated)
	log.Printf("  Time since: %v", timeSince)
	log.Printf("  Threshold: %v", threshold)
	log.Printf("  Should update: %v", timeSince > threshold)

	// TEMPORARY: Force all updates for debugging
	log.Printf("  FORCING UPDATE FOR DEBUGGING")
	return true
}

// Alternative: If you want to keep the normal logic but with debug info:
func ShouldUpdateFeedNormal(source FeedSource) bool {
	timeSince := time.Since(source.LastUpdated)
	threshold := time.Duration(source.UpdateInterval) * time.Second
	shouldUpdate := timeSince > threshold

	log.Printf("DEBUG ShouldUpdateFeed - Source: %s", source.Name)
	log.Printf("  LastUpdated: %v", source.LastUpdated)
	log.Printf("  Time since: %v", timeSince)
	log.Printf("  Threshold: %v (UpdateInterval: %d seconds)", threshold, source.UpdateInterval)
	log.Printf("  Should update: %v", shouldUpdate)

	return shouldUpdate
}

// Also add this manual test function to your main.go to bypass the scheduler:
func manualFeedTest(db *sql.DB) {
	log.Println("=== MANUAL FEED TEST ===")

	// Test just one feed manually
	redditFeed := &RedditFeed{Subreddit: "programming"}
	sourceName := redditFeed.GetSourceName()
	feedURL := redditFeed.GetFeedURL()

	log.Printf("Testing: %s", sourceName)
	log.Printf("URL: %s", feedURL)

	// Get the source from database
	dbSource, err := GetFeedSourceByName(db, sourceName)
	if err != nil {
		log.Printf("ERROR: Cannot find source in database: %v", err)
		return
	}

	log.Printf("Found source - ID: %d, LastUpdated: %v", dbSource.ID, dbSource.LastUpdated)

	// Fetch the feed
	log.Println("Fetching feed content...")
	content, err := FetchFeed(feedURL)
	if err != nil {
		log.Printf("ERROR: Failed to fetch: %v", err)
		return
	}

	log.Printf("SUCCESS: Fetched %d bytes", len(content))

	// Parse the feed
	log.Println("Parsing feed...")
	items, err := redditFeed.ParseFeed(content, dbSource.ID)
	if err != nil {
		log.Printf("ERROR: Failed to parse: %v", err)
		return
	}

	log.Printf("SUCCESS: Parsed %d items", len(items))
	for i, item := range items {
		if i >= 3 {
			break
		}
		log.Printf("  Item %d: %s", i+1, item.Title)
	}

	// Save items
	log.Println("Saving items to database...")
	err = SaveFeedItems(db, items)
	if err != nil {
		log.Printf("ERROR: Failed to save: %v", err)
		return
	}

	log.Printf("SUCCESS: Saved %d items", len(items))

	// Check what's in database now
	allItems, err := GetAllFeedItems(db, 10)
	if err != nil {
		log.Printf("ERROR: Failed to get items from database: %v", err)
		return
	}

	log.Printf("Total items now in database: %d", len(allItems))
	for i, item := range allItems {
		if i >= 3 {
			break
		}
		log.Printf("  DB Item %d: %s (Source: %d)", i+1, item.Title, item.SourceID)
	}
}

// Gets all feed sources.
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

// Gets feed items for a specific source.
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

// Gets all feed items sorted by published date.
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
