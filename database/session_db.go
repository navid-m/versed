package database

import (
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
)

// Implements database-backed session storage
type DBSessionStorage struct {
	db *sql.DB
}

// Creates a new database-backed session storage
func NewDBSessionStorage(db *sql.DB) *DBSessionStorage {
	return &DBSessionStorage{db: db}
}

// Retrieves a session from the database
func (s *DBSessionStorage) Get(key string) ([]byte, error) {
	var data string
	var expiresAt sql.NullTime

	sqlQuery, args, err := squirrel.Select("data", "expires_at").
		From("sessions").
		Where(squirrel.Eq{"session_id": key}).
		Where(squirrel.Or{
			squirrel.Eq{"expires_at": nil},
			squirrel.Expr("expires_at > datetime('now')"),
		}).ToSql()

	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(sqlQuery, args...).Scan(&data, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

// Stores a session in the database
func (s *DBSessionStorage) Set(key string, val []byte, exp time.Duration) error {
	var expiresAt any
	if exp > 0 {
		expiresAt = time.Now().Add(exp)
	} else {
		expiresAt = nil
	}

	data := string(val)

	sqlQuery, args, err := squirrel.Insert("sessions").
		Columns("session_id", "data", "updated_at", "expires_at").
		Values(key, data, squirrel.Expr("datetime('now')"), expiresAt).
		Suffix("ON CONFLICT(session_id) DO UPDATE SET data = EXCLUDED.data, updated_at = EXCLUDED.updated_at, expires_at = EXCLUDED.expires_at").
		ToSql()

	if err != nil {
		return err
	}

	_, err = s.db.Exec(sqlQuery, args...)
	return err
}

// Removes a session from the database
func (s *DBSessionStorage) Delete(key string) error {
	sqlQuery, args, err := squirrel.Delete("sessions").
		Where(squirrel.Eq{"session_id": key}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(sqlQuery, args...)
	return err
}

// Clears all sessions from the database
func (s *DBSessionStorage) Reset() error {
	sqlQuery, args, err := squirrel.Delete("sessions").ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(sqlQuery, args...)
	return err
}

// Close is a no-op for database storage
func (s *DBSessionStorage) Close() error {
	return nil
}
