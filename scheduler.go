package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"verse/feeds"

	_ "github.com/mattn/go-sqlite3"
)

// Manages periodic feed updates
type FeedScheduler struct {
	db          *sql.DB
	feedManager *feeds.FeedManager
	ticker      *time.Ticker
	stopChan    chan bool
}

// Creates a new feed scheduler
func NewFeedScheduler(db *sql.DB) *FeedScheduler {
	return &FeedScheduler{
		db:          db,
		feedManager: feeds.NewFeedManager(),
		stopChan:    make(chan bool),
	}
}

// Begins the periodic feed update process
func (fs *FeedScheduler) Start() {
	log.Println("Starting feed scheduler...")

	fs.updateAllFeeds()

	fs.ticker = time.NewTicker(1 * time.Hour)

	go func() {
		for {
			select {
			case <-fs.ticker.C:
				fs.updateAllFeeds()
			case <-fs.stopChan:
				fs.ticker.Stop()
				return
			}
		}
	}()
}

// Stops the scheduler
func (fs *FeedScheduler) Stop() {
	if fs.stopChan != nil {
		fs.stopChan <- true
	}
}

// Fetches and caches feeds from all sources
func (fs *FeedScheduler) updateAllFeeds() {
	log.Println("Starting feed update process...")

	for _, source := range fs.feedManager.Sources {
		go fs.updateFeed(source)
	}
}
func debugDatabase(db *sql.DB) {
	log.Println("=== DATABASE DEBUG ===")

	// Check feed sources
	sources, err := feeds.GetAllFeedSources(db)
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
	items, err := feeds.GetAllFeedItems(db, 10)
	if err != nil {
		log.Printf("ERROR getting feed items: %v", err)
		return
	}
	log.Printf("Feed items in database: %d", len(items))
	for _, item := range items {
		log.Printf("  - %s (Source: %d)", item.Title, item.SourceID)
	}
}

// Replace your CreateOrUpdateFeedSource function in feeds/feeds.go with this:
func CreateOrUpdateFeedSource(db *sql.DB, name, url string) (*feeds.FeedSource, error) {
	existing, err := feeds.GetFeedSourceByName(db, name)
	if err == nil {
		log.Printf(
			"Found existing source: %s (ID: %d, LastUpdated: %v)",
			name,
			existing.ID,
			existing.LastUpdated,
		)
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

	source := &feeds.FeedSource{
		ID:             int(id),
		Name:           name,
		URL:            url,
		LastUpdated:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdateInterval: 3600,
	}

	log.Printf("Created new feed source with ID: %d", source.ID)
	return source, nil
}

// Quick fix to force update all existing feeds - add this function to feeds/feeds.go:
func ResetAllFeedTimestamps(db *sql.DB) error {
	query := `UPDATE feed_sources SET last_updated = datetime('2000-01-01 00:00:00')`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to reset feed timestamps: %w", err)
	}
	log.Println("Reset all feed timestamps to force updates")
	return nil
}

// Also fix the scheduler updateFeed function to add more logging:
func (fs *FeedScheduler) updateFeed(source feeds.FeedSourceInterface) {
	sourceName := source.GetSourceName()
	feedURL := source.GetFeedURL()
	log.Printf("=== Processing feed: %s ===", sourceName)
	log.Printf("URL: %s", feedURL)

	dbSource, err := feeds.CreateOrUpdateFeedSource(fs.db, sourceName, feedURL)
	if err != nil {
		log.Printf("ERROR: Failed to create/update feed source %s: %v", sourceName, err)
		return
	}
	log.Printf("DB Source - ID: %d, LastUpdated: %v", dbSource.ID, dbSource.LastUpdated)

	shouldUpdate := feeds.ShouldUpdateFeed(*dbSource)
	log.Printf("Should update feed %s: %v", sourceName, shouldUpdate)

	if !shouldUpdate {
		log.Printf("SKIPPING: Feed %s is up to date (last updated: %v)", sourceName, dbSource.LastUpdated)
		return
	}

	log.Printf("FETCHING: RSS content from %s", feedURL)
	content, err := feeds.FetchFeed(feedURL)
	if err != nil {
		log.Printf("ERROR: Failed to fetch feed %s: %v", sourceName, err)
		return
	}

	log.Printf("SUCCESS: Fetched %d bytes from %s", len(content), sourceName)
	items, err := source.ParseFeed(content, dbSource.ID)
	if err != nil {
		log.Printf("ERROR: Failed to parse feed %s: %v", sourceName, err)
		return
	}

	log.Printf("SUCCESS: Parsed %d items from %s", len(items), sourceName)

	// Log a few sample items
	for i, item := range items {
		if i >= 2 {
			break
		}
		log.Printf("  Sample item %d: %s", i+1, item.Title)
	}

	err = feeds.SaveFeedItems(fs.db, items)
	if err != nil {
		log.Printf("ERROR: Failed to save feed items for %s: %v", sourceName, err)
		return
	}

	log.Printf("SUCCESS: Saved %d items for %s", len(items), sourceName)

	err = feeds.UpdateFeedSourceTimestamp(fs.db, dbSource.ID)
	if err != nil {
		log.Printf("ERROR: Failed to update timestamp for %s: %v", sourceName, err)
		return
	}

	log.Printf("SUCCESS: Updated timestamp for %s", sourceName)
	log.Printf("=== Finished processing: %s ===\n", sourceName)
}
