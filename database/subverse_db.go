package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/navid-m/versed/feeds"
	"github.com/navid-m/versed/models"
)

// CreateSubverse creates a new subverse
func CreateSubverse(db *sql.DB, name string) (*models.Subverse, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return nil, fmt.Errorf("subverse name cannot be empty")
	}

	query := `INSERT INTO subverses (name) VALUES (?)`
	result, err := db.Exec(query, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create subverse: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get subverse ID: %w", err)
	}

	subverse := &models.Subverse{
		ID:        int(id),
		Name:      name,
		CreatedAt: time.Now(),
	}

	return subverse, nil
}

// GetSubverses retrieves all subverses
func GetSubverses(db *sql.DB) ([]models.Subverse, error) {
	query := `SELECT id, name, created_at FROM subverses ORDER BY name`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get subverses: %w", err)
	}
	defer rows.Close()

	var subverses []models.Subverse
	for rows.Next() {
		var subverse models.Subverse
		err := rows.Scan(&subverse.ID, &subverse.Name, &subverse.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subverse: %w", err)
		}
		subverses = append(subverses, subverse)
	}

	return subverses, nil
}

// AddFeedToSubverse adds a feed source to a subverse
func AddFeedToSubverse(db *sql.DB, subverseID, feedSourceID int) error {
	query := `INSERT INTO subverse_feeds (subverse_id, feed_source_id) VALUES (?, ?)`
	_, err := db.Exec(query, subverseID, feedSourceID)
	return err
}

// RemoveFeedFromSubverse removes a feed source from a subverse
func RemoveFeedFromSubverse(db *sql.DB, subverseID, feedSourceID int) error {
	query := `DELETE FROM subverse_feeds WHERE subverse_id = ? AND feed_source_id = ?`
	_, err := db.Exec(query, subverseID, feedSourceID)
	return err
}

// GetSubverseFeeds gets all feed sources associated with a subverse
func GetSubverseFeeds(db *sql.DB, subverseID int) ([]feeds.FeedSource, error) {
	query := `
		SELECT fs.id, fs.name, fs.url, fs.last_updated, fs.update_interval
		FROM feed_sources fs
		INNER JOIN subverse_feeds sf ON fs.id = sf.feed_source_id
		WHERE sf.subverse_id = ?
		ORDER BY fs.name
	`
	rows, err := db.Query(query, subverseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subverse feeds: %w", err)
	}
	defer rows.Close()

	var sources []feeds.FeedSource
	for rows.Next() {
		var source feeds.FeedSource
		err := rows.Scan(&source.ID, &source.Name, &source.URL, &source.LastUpdated, &source.UpdateInterval)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feed: %w", err)
		}
		sources = append(sources, source)
	}

	return sources, nil
}

// GetSubverseFeedItems gets feed items from feeds associated with a subverse
func GetSubverseFeedItems(db *sql.DB, subverseID int, limit int) ([]feeds.FeedItem, error) {
	query := `
		SELECT fi.id, fi.source_id, fi.title, fi.url, fi.description, fi.author, fi.published_at, fi.score, fi.comments_count, fi.created_at, fs.name
		FROM feed_items fi
		INNER JOIN feed_sources fs ON fi.source_id = fs.id
		INNER JOIN subverse_feeds sf ON fs.id = sf.feed_source_id
		WHERE sf.subverse_id = ?
		ORDER BY fi.published_at DESC
		LIMIT ?
	`
	rows, err := db.Query(query, subverseID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get subverse feed items: %w", err)
	}
	defer rows.Close()

	var items []feeds.FeedItem
	for rows.Next() {
		var item feeds.FeedItem
		err := rows.Scan(&item.ID, &item.SourceID, &item.Title, &item.URL, &item.Description, &item.Author, &item.PublishedAt, &item.Score, &item.CommentsCount, &item.CreatedAt, &item.SourceName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feed item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func UpdateSubversePostCount(db *sql.DB, subverseID int) error {
	var column string
	err := db.QueryRow("SELECT 1 FROM pragma_table_info('subverses') WHERE name='post_count'").Scan(&column)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		} else {
			return err
		}
	}

	query := `UPDATE subverses SET post_count = post_count + 1 WHERE id = ?`
	_, err = db.Exec(query, subverseID)
	return err
}
