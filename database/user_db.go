package database

import (
	"github.com/navid-m/versed/models"

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
