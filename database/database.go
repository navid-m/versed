package database

import (
	"database/sql"
)

var db *sql.DB

func InitDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "./users.db")
	if err != nil {
		return err
	}
	return createTables()
}

func CloseConnection() error {
	return db.Close()
}

func createTables() error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	)`
	_, err := db.Exec(query)
	return err
}
