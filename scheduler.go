package main

import (
	"database/sql"
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

	go fs.updateAllFeeds()

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

func (fs *FeedScheduler) updateFeed(source feeds.FeedSourceInterface) {
	sourceName := source.GetSourceName()
	feedURL := source.GetFeedURL()

	log.Printf("Updating feed: %s from URL: %s", sourceName, feedURL)

	dbSource, err := feeds.CreateOrUpdateFeedSource(fs.db, sourceName, feedURL)
	if err != nil {
		log.Printf("Failed to create/update feed source %s: %v", sourceName, err)
		return
	}
	if !feeds.ShouldUpdateFeed(feeds.FeedSource(*dbSource)) {
		log.Printf("Feed %s is up to date", sourceName)
		return
	}

	log.Printf("Fetching RSS content from %s", feedURL)
	content, err := feeds.FetchFeed(feedURL)
	if err != nil {
		log.Printf("Failed to fetch feed %s: %v", sourceName, err)
		return
	}
	log.Printf("Successfully fetched %d bytes from %s", len(content), sourceName)

	items, err := source.ParseFeed(content, dbSource.ID)
	if err != nil {
		log.Printf("Failed to parse feed %s: %v", sourceName, err)
		return
	}
	log.Printf("Parsed %d items from %s", len(items), sourceName)

	err = feeds.SaveFeedItems(fs.db, items)
	if err != nil {
		log.Printf("Failed to save feed items for %s: %v", sourceName, err)
		return
	}
	err = feeds.UpdateFeedSourceTimestamp(fs.db, dbSource.ID)
	if err != nil {
		log.Printf("Failed to update timestamp for %s: %v", sourceName, err)
		return
	}

	log.Printf("Successfully updated feed: %s (%d items)", sourceName, len(items))
}
