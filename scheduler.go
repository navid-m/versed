package main

import (
	"database/sql"
	"log"
	"time"
	"verse/feeds"

	_ "github.com/mattn/go-sqlite3"
)

// FeedScheduler manages periodic feed updates
type FeedScheduler struct {
	db          *sql.DB
	feedManager *feeds.FeedManager
	ticker      *time.Ticker
	stopChan    chan bool
}

// NewFeedScheduler creates a new feed scheduler
func NewFeedScheduler(db *sql.DB) *FeedScheduler {
	return &FeedScheduler{
		db:          db,
		feedManager: feeds.NewFeedManager(),
		stopChan:    make(chan bool),
	}
}

// Start begins the periodic feed update process
func (fs *FeedScheduler) Start() {
	log.Println("Starting feed scheduler...")

	// Initial feed update
	go fs.updateAllFeeds()

	// Set up periodic updates every hour
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

// Stop stops the scheduler
func (fs *FeedScheduler) Stop() {
	if fs.stopChan != nil {
		fs.stopChan <- true
	}
}

// updateAllFeeds fetches and caches feeds from all sources
func (fs *FeedScheduler) updateAllFeeds() {
	log.Println("Starting feed update process...")

	for _, source := range fs.feedManager.Sources {
		go fs.updateFeed(source)
	}
}

// updateFeed updates a single feed source
func (fs *FeedScheduler) updateFeed(source feeds.FeedSourceInterface) {
	sourceName := source.GetSourceName()
	feedURL := source.GetFeedURL()

	log.Printf("Updating feed: %s", sourceName)

	// Create or update feed source in database
	dbSource, err := feeds.CreateOrUpdateFeedSource(fs.db, sourceName, feedURL)
	if err != nil {
		log.Printf("Failed to create/update feed source %s: %v", sourceName, err)
		return
	}

	// Check if feed needs updating
	if !feeds.ShouldUpdateFeed(feeds.FeedSource(*dbSource)) {
		log.Printf("Feed %s is up to date", sourceName)
		return
	}

	// Fetch feed content
	content, err := feeds.FetchFeed(feedURL)
	if err != nil {
		log.Printf("Failed to fetch feed %s: %v", sourceName, err)
		return
	}

	// Parse feed content
	items, err := source.ParseFeed(content, dbSource.ID)
	if err != nil {
		log.Printf("Failed to parse feed %s: %v", sourceName, err)
		return
	}

	// Save feed items to database
	err = feeds.SaveFeedItems(fs.db, items)
	if err != nil {
		log.Printf("Failed to save feed items for %s: %v", sourceName, err)
		return
	}

	// Update last updated timestamp
	err = feeds.UpdateFeedSourceTimestamp(fs.db, dbSource.ID)
	if err != nil {
		log.Printf("Failed to update timestamp for %s: %v", sourceName, err)
		return
	}

	log.Printf("Successfully updated feed: %s (%d items)", sourceName, len(items))
}
