package database

import (
	"verse/models"

	"github.com/Masterminds/squirrel"
)

// Creates a new user in the database
func CreateUser(email, username, password string) error {
	sqlQuery, args, err := squirrel.Insert("users").
		Columns("email", "username", "password").
		Values(email, username, password).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(sqlQuery, args...)
	return err
}

// Retrieves the user object given some email address
func GetUserByEmail(email string) (*models.User, error) {
	sqlQuery, args, err := squirrel.Select("id", "email", "username", "password").
		From("users").
		Where(squirrel.Eq{"email": email}).
		ToSql()

	if err != nil {
		return nil, err
	}

	row := db.QueryRow(sqlQuery, args...)
	user := &models.User{}
	err = row.Scan(&user.ID, &user.Email, &user.Username, &user.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Updates user information in the database
func UpdateUser(userID int, email, username, password string) error {
	sqlQuery, args, err := squirrel.Update("users").
		Set("email", email).
		Set("username", username).
		Set("password", password).
		Where(squirrel.Eq{"id": userID}).
		ToSql()

	if err != nil {
		return err
	}

	_, err = db.Exec(sqlQuery, args...)
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

	sqlQuery, args, err := squirrel.Insert("reading_list").
		Columns("user_id", "item_id").
		Values(userID, itemID).
		ToSql()

	if err != nil {
		return false, err
	}

	_, err = db.Exec(sqlQuery, args...)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Removes a feed item from user's reading list
func RemoveFromReadingList(userID int, itemID string) error {
	sqlQuery, args, err := squirrel.Delete("reading_list").
		Where(squirrel.Eq{"user_id": userID, "item_id": itemID}).
		ToSql()

	if err != nil {
		return err
	}

	_, err = db.Exec(sqlQuery, args...)
	return err
}

// Checks if a feed item is in user's reading list
func IsInReadingList(userID int, itemID string) (bool, error) {
	var count int

	sqlQuery, args, err := squirrel.Select("COUNT(*)").
		From("reading_list").
		Where(squirrel.Eq{"user_id": userID, "item_id": itemID}).
		ToSql()

	if err != nil {
		return false, err
	}

	err = db.QueryRow(sqlQuery, args...).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
