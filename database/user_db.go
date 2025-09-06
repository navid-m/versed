package database

import "verse/models"

func CreateUser(email, password string) error {
	var (
		query  = `INSERT INTO users (email, password) VALUES (?, ?)`
		_, err = db.Exec(query, email, password)
	)
	return err
}

func GetUserByEmail(email string) (*models.User, error) {
	var (
		query = `SELECT id, email, password FROM users WHERE email = ?`
		row   = db.QueryRow(query, email)
		user  = &models.User{}
		err   = row.Scan(&user.ID, &user.Email, &user.Password)
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Adds a feed item to user's reading list
func SaveToReadingList(userID int, itemID string) error {
	query := `INSERT OR IGNORE INTO reading_list (user_id, item_id) VALUES (?, ?)`
	_, err := db.Exec(query, userID, itemID)
	return err
}

// Removes a feed item from user's reading list
func RemoveFromReadingList(userID int, itemID string) error {
	query := `DELETE FROM reading_list WHERE user_id = ? AND item_id = ?`
	_, err := db.Exec(query, userID, itemID)
	return err
}

// Checks if a feed item is in user's reading list
func IsInReadingList(userID int, itemID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM reading_list WHERE user_id = ? AND item_id = ?`
	err := db.QueryRow(query, userID, itemID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
