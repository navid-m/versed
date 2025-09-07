package database

import "verse/models"

// Creates a new user in the database
func CreateUser(email, username, password string) error {
	var (
		query  = `INSERT INTO users (email, username, password) VALUES (?, ?, ?)`
		_, err = db.Exec(query, email, username, password)
	)
	return err
}

// Retrieves the user object given some email address
func GetUserByEmail(email string) (*models.User, error) {
	var (
		query = `SELECT id, email, username, password FROM users WHERE email = ?`
		row   = db.QueryRow(query, email)
		user  = &models.User{}
		err   = row.Scan(&user.ID, &user.Email, &user.Username, &user.Password)
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Updates user information in the database
func UpdateUser(userID int, email, username, password string) error {
	query := `UPDATE users SET email = ?, username = ?, password = ? WHERE id = ?`
	_, err := db.Exec(query, email, username, password, userID)
	return err
}

// Adds a feed item to user's reading list
//
// Returns (saved or not -> bool, error)
func SaveToReadingList(userID int, itemID string) (bool, error) {
	exists, err := IsInReadingList(userID, itemID)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}
	query := `INSERT INTO reading_list (user_id, item_id) VALUES (?, ?)`
	_, err = db.Exec(query, userID, itemID)
	if err != nil {
		return false, err
	}
	return true, nil
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
