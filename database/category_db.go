package database

import (
	"database/sql"
	"fmt"
	"time"

	"verse/models"
	"verse/feeds"
)

// CreateUserCategory creates a new category for a user
func CreateUserCategory(db *sql.DB, userID int, name, description string) (*models.UserCategory, error) {
	query := `INSERT INTO user_categories (user_id, name, description) VALUES (?, ?, ?)`
	result, err := db.Exec(query, userID, name, description)
	if err != nil {
		return nil, fmt.Errorf("failed to create user category: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get category ID: %w", err)
	}

	category := &models.UserCategory{
		ID:          int(id),
		UserID:      userID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}

	return category, nil
}

// GetUserCategories retrieves all categories for a user
func GetUserCategories(db *sql.DB, userID int) ([]models.UserCategory, error) {
	query := `SELECT id, user_id, name, description, created_at FROM user_categories WHERE user_id = ? ORDER BY name`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user categories: %w", err)
	}
	defer rows.Close()

	var categories []models.UserCategory
	for rows.Next() {
		var category models.UserCategory
		err := rows.Scan(&category.ID, &category.UserID, &category.Name, &category.Description, &category.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// GetUserCategoryByID retrieves a specific category for a user
func GetUserCategoryByID(db *sql.DB, userID, categoryID int) (*models.UserCategory, error) {
	query := `SELECT id, user_id, name, description, created_at FROM user_categories WHERE id = ? AND user_id = ?`
	row := db.QueryRow(query, categoryID, userID)

	var category models.UserCategory
	err := row.Scan(&category.ID, &category.UserID, &category.Name, &category.Description, &category.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

// UpdateUserCategory updates a category's name and description
func UpdateUserCategory(db *sql.DB, userID, categoryID int, name, description string) error {
	query := `UPDATE user_categories SET name = ?, description = ? WHERE id = ? AND user_id = ?`
	result, err := db.Exec(query, name, description, categoryID, userID)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found or no changes made")
	}

	return nil
}

// DeleteUserCategory deletes a category and all its feed associations
func DeleteUserCategory(db *sql.DB, userID, categoryID int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// First delete all feed associations for this category
	_, err = tx.Exec(`DELETE FROM user_category_feeds WHERE user_id = ? AND category_id = ?`, userID, categoryID)
	if err != nil {
		return fmt.Errorf("failed to delete category feeds: %w", err)
	}

	// Then delete the category
	result, err := tx.Exec(`DELETE FROM user_categories WHERE id = ? AND user_id = ?`, categoryID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return tx.Commit()
}

// AddFeedToUserCategory adds a feed source to a user's category
func AddFeedToUserCategory(db *sql.DB, userID, categoryID, feedSourceID int) error {
	// Check if the category belongs to the user
	_, err := GetUserCategoryByID(db, userID, categoryID)
	if err != nil {
		return fmt.Errorf("invalid category: %w", err)
	}

	// Check if the feed source exists
	var exists bool
	err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM feed_sources WHERE id = ?)`, feedSourceID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check feed source: %w", err)
	}
	if !exists {
		return fmt.Errorf("feed source not found")
	}

	query := `INSERT OR IGNORE INTO user_category_feeds (user_id, category_id, feed_source_id) VALUES (?, ?, ?)`
	_, err = db.Exec(query, userID, categoryID, feedSourceID)
	if err != nil {
		return fmt.Errorf("failed to add feed to category: %w", err)
	}

	return nil
}

// RemoveFeedFromUserCategory removes a feed source from a user's category
func RemoveFeedFromUserCategory(db *sql.DB, userID, categoryID, feedSourceID int) error {
	query := `DELETE FROM user_category_feeds WHERE user_id = ? AND category_id = ? AND feed_source_id = ?`
	result, err := db.Exec(query, userID, categoryID, feedSourceID)
	if err != nil {
		return fmt.Errorf("failed to remove feed from category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("feed not found in category")
	}

	return nil
}

// GetFeedsInUserCategory gets all feed sources in a user's category
func GetFeedsInUserCategory(db *sql.DB, userID, categoryID int) ([]feeds.FeedSource, error) {
	query := `SELECT fs.id, fs.name, fs.url, fs.last_updated, fs.update_interval
	          FROM feed_sources fs
	          JOIN user_category_feeds ucf ON fs.id = ucf.feed_source_id
	          WHERE ucf.user_id = ? AND ucf.category_id = ?
	          ORDER BY fs.name`
	rows, err := db.Query(query, userID, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feeds in category: %w", err)
	}
	defer rows.Close()

	var sources []feeds.FeedSource
	for rows.Next() {
		var source feeds.FeedSource
		err := rows.Scan(&source.ID, &source.Name, &source.URL, &source.LastUpdated, &source.UpdateInterval)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feed source: %w", err)
		}
		sources = append(sources, source)
	}

	return sources, nil
}

// GetUserCategoriesForFeed gets all categories for a user that contain a specific feed
func GetUserCategoriesForFeed(db *sql.DB, userID, feedSourceID int) ([]models.UserCategory, error) {
	query := `SELECT uc.id, uc.user_id, uc.name, uc.description, uc.created_at
	          FROM user_categories uc
	          JOIN user_category_feeds ucf ON uc.id = ucf.category_id
	          WHERE uc.user_id = ? AND ucf.feed_source_id = ?
	          ORDER BY uc.name`
	rows, err := db.Query(query, userID, feedSourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories for feed: %w", err)
	}
	defer rows.Close()

	var categories []models.UserCategory
	for rows.Next() {
		var category models.UserCategory
		err := rows.Scan(&category.ID, &category.UserID, &category.Name, &category.Description, &category.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}
