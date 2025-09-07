package database

import (
	"database/sql"

	"github.com/navid-m/versed/models"

	"github.com/Masterminds/squirrel"
	"golang.org/x/crypto/bcrypt"
)

// Creates a new user in the database with a hashed password
func CreateUser(email, username, password string) error {
	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	sqlQuery, args, err := squirrel.Insert("users").
		Columns("email", "username", "password").
		Values(email, username, string(hashedPassword)).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(sqlQuery, args...)
	if err != nil {
		return err
	}
	user, err := GetUserByEmail(email)
	if err != nil {
		return err
	}

	return CreateDefaultCategoriesForUser(user.ID)
}

// Creates default categories and adds popular feeds for a new user
func CreateDefaultCategoriesForUser(userID int) error {
	defaultCategories := map[string][]struct {
		name string
		url  string
	}{
		"Technology": {
			{"Hacker News", "https://hnrss.org/frontpage"},
			{"Reddit - Programming", "https://www.reddit.com/r/programming/.rss"},
			{"Lobsters", "https://lobste.rs/rss"},
		},
		"News": {
			{"Reddit - Technology", "https://www.reddit.com/r/technology/.rss"},
			{"BBC News - Technology", "http://feeds.bbci.co.uk/news/technology/rss.xml"},
		},
		"Science": {
			{"Reddit - Science", "https://www.reddit.com/r/science/.rss"},
			{"Nature News", "https://www.nature.com/nature.rss"},
		},
	}

	for categoryName, feeds := range defaultCategories {
		category, err := CreateUserCategory(db, userID, categoryName, "Default "+categoryName+" feeds")
		if err != nil {
			continue
		}

		for _, feed := range feeds {
			feedSourceID, err := EnsureFeedSourceExists(feed.name, feed.url)
			if err != nil {
				continue
			}

			err = AddFeedToUserCategory(db, userID, category.ID, feedSourceID)
			if err != nil {
				continue
			}
		}
	}

	return nil
}

// Creates a feed source if it doesn't exist and returns its ID
func EnsureFeedSourceExists(name, url string) (int, error) {
	var id int
	err := db.QueryRow("SELECT id FROM feed_sources WHERE name = ?", name).Scan(&id)
	if err == nil {
		return id, nil
	}

	sqlQuery, args, err := squirrel.Insert("feed_sources").
		Columns("name", "url", "last_updated", "update_interval").
		Values(name, url, squirrel.Expr("datetime('2000-01-01 00:00:00')"), 3600).
		ToSql()
	if err != nil {
		return 0, err
	}

	result, err := db.Exec(sqlQuery, args...)
	if err != nil {
		return 0, err
	}

	id64, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id64), nil
}

// Retrieves the user object given some email address
func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	var isAdmin sql.NullBool

	err := db.QueryRow("SELECT id, email, username, password, is_admin FROM users WHERE email = ?", email).
		Scan(&user.ID, &user.Email, &user.Username, &user.Password, &isAdmin)

	if err != nil {
		return nil, err
	}

	// Convert sql.NullBool to bool with false as default
	user.IsAdmin = isAdmin.Bool

	return &user, nil
}

// VerifyPassword checks if the provided password matches the hashed password
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Updates user information in the database
func UpdateUser(userID int, email, username, password string) error {
	update := squirrel.Update("users").
		Set("email", email).
		Set("username", username).
		Where(squirrel.Eq{"id": userID})

	// Only update password if it's not empty
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		update = update.Set("password", string(hashedPassword))
	}

	sqlQuery, args, err := update.ToSql()
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
