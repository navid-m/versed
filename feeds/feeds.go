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
	SourceName    string    `json:"source_name"`
	Title         string    `json:"title"`
	URL           string    `json:"url"`
	Description   string    `json:"description,omitempty"`
	Author        string    `json:"author,omitempty"`
	PublishedAt   time.Time `json:"published_at"`
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

func DebugFeeds(db *sql.DB) {
	log.Println("=== DATABASE DEBUG ===")

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
	log.Printf("=== FetchFeed called ===")
	log.Printf("Fetching RSS content from: %s", url)
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
		VALUES (?, ?, ?, ?, ?, ?, ?, 
			COALESCE((SELECT score FROM feed_items WHERE id = ?), ?), 
			?, ?)
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
			item.ID,
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
	existing, err := GetFeedSourceByName(db, name)
	if err == nil {
		log.Printf("Found existing source: %s (ID: %d, LastUpdated: %v)",
			name, existing.ID, existing.LastUpdated)
		return existing, nil
	}

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

func ShouldUpdateFeed(source FeedSource) bool {
	timeSince := time.Since(source.LastUpdated)
	threshold := time.Duration(source.UpdateInterval) * time.Second

	log.Printf("=== ShouldUpdateFeed Check ===")
	log.Printf("Source: %s (ID: %d)", source.Name, source.ID)
	log.Printf("LastUpdated: %v", source.LastUpdated)
	log.Printf("Time since: %v", timeSince)
	log.Printf("Threshold: %v", threshold)
	log.Printf("UpdateInterval: %d seconds", source.UpdateInterval)
	shouldUpdate := timeSince > threshold
	log.Printf("Should update: %v", shouldUpdate)
	log.Printf("FORCING UPDATE FOR DEBUGGING")

	return true // Force update for debugging
}

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
		item.SourceName = sourceName
		items = append(items, item)
	}
	return items, nil
}

// Gets all feed items sorted by published date, excluding hidden ones for a user
func GetAllFeedItemsForUser(db *sql.DB, userID int, limit int) ([]FeedItem, error) {
	query := `SELECT fi.id, fi.source_id, fi.title, fi.url, fi.description, fi.author, fi.published_at, fi.score, fi.comments_count, fi.created_at, fs.name as source_name
			FROM feed_items fi
			JOIN feed_sources fs ON fi.source_id = fs.id
			LEFT JOIN hidden_posts hp ON fi.id = hp.item_id AND hp.user_id = ?
			WHERE hp.item_id IS NULL
			ORDER BY fi.published_at DESC
			LIMIT ?`
	rows, err := db.Query(query, userID, limit)
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
		item.SourceName = sourceName
		items = append(items, item)
	}
	return items, nil
}

// Gets all feed items with pagination support
func GetAllFeedItemsWithPagination(db *sql.DB, limit, offset int) ([]FeedItem, error) {
	query := `SELECT fi.id, fi.source_id, fi.title, fi.url, fi.description, fi.author, fi.published_at, fi.score, fi.comments_count, fi.created_at, fs.name as source_name
			FROM feed_items fi
			JOIN feed_sources fs ON fi.source_id = fs.id
			ORDER BY fi.published_at DESC
			LIMIT ? OFFSET ?`
	rows, err := db.Query(query, limit, offset)
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
		item.SourceName = sourceName
		items = append(items, item)
	}
	return items, nil
}

// Gets all feed items with pagination support, excluding hidden ones for a user
func GetAllFeedItemsWithPaginationForUser(db *sql.DB, userID int, limit, offset int) ([]FeedItem, error) {
	query := `SELECT fi.id, fi.source_id, fi.title, fi.url, fi.description, fi.author, fi.published_at, fi.score, fi.comments_count, fi.created_at, fs.name as source_name
			FROM feed_items fi
			JOIN feed_sources fs ON fi.source_id = fs.id
			LEFT JOIN hidden_posts hp ON fi.id = hp.item_id AND hp.user_id = ?
			WHERE hp.item_id IS NULL
			ORDER BY fi.published_at DESC
			LIMIT ? OFFSET ?`
	rows, err := db.Query(query, userID, limit, offset)
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
		item.SourceName = sourceName
		items = append(items, item)
	}
	return items, nil
}

func HandleVote(db *sql.DB, itemID string, userID int, voteType string) (int, error) {
	var existingVoteType sql.NullString
	err := db.QueryRow("SELECT vote_type FROM upvotes WHERE user_id = ? AND item_id = ?", userID, itemID).Scan(&existingVoteType)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to check existing vote: %w", err)
	}
	var currentScore int
	err = db.QueryRow("SELECT score FROM feed_items WHERE id = ?", itemID).Scan(&currentScore)
	if err != nil {
		return 0, fmt.Errorf("failed to get current score: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if existingVoteType.Valid {
		if existingVoteType.String == voteType {
			_, err = tx.Exec("DELETE FROM upvotes WHERE user_id = ? AND item_id = ?", userID, itemID)
			if err != nil {
				return 0, fmt.Errorf("failed to delete vote: %w", err)
			}
			if voteType == "upvote" {
				currentScore--
			} else {
				currentScore++
			}
		} else {
			_, err = tx.Exec("UPDATE upvotes SET vote_type = ? WHERE user_id = ? AND item_id = ?", voteType, userID, itemID)
			if err != nil {
				return 0, fmt.Errorf("failed to update vote: %w", err)
			}
			if existingVoteType.String == "upvote" {
				currentScore -= 2
			} else {
				currentScore += 2
			}
		}
	} else {
		_, err = tx.Exec("INSERT INTO upvotes (user_id, item_id, vote_type) VALUES (?, ?, ?)", userID, itemID, voteType)
		if err != nil {
			return 0, fmt.Errorf("failed to insert vote: %w", err)
		}
		if voteType == "upvote" {
			currentScore++
		} else {
			currentScore--
		}
	}

	_, err = tx.Exec("UPDATE feed_items SET score = ? WHERE id = ?", currentScore, itemID)
	if err != nil {
		return 0, fmt.Errorf("failed to update score: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return currentScore, nil
}
