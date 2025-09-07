package database

import (
	"database/sql"
	"time"
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
	err := s.db.QueryRow(`
		SELECT data, expires_at FROM sessions 
		WHERE session_id = ? AND (expires_at IS NULL OR expires_at > datetime('now'))
	`, key).Scan(&data, &expiresAt)
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
	var expiresAt interface{}
	if exp > 0 {
		expiresAt = time.Now().Add(exp)
	} else {
		expiresAt = nil
	}

	data := string(val)
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO sessions (session_id, data, updated_at, expires_at) 
		VALUES (?, ?, datetime('now'), ?)
	`, key, data, expiresAt)
	return err
}

// Removes a session from the database
func (s *DBSessionStorage) Delete(key string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE session_id = ?`, key)
	return err
}

// Clears all sessions from the database
func (s *DBSessionStorage) Reset() error {
	_, err := s.db.Exec(`DELETE FROM sessions`)
	return err
}

// Close is a no-op for database storage
func (s *DBSessionStorage) Close() error {
	return nil
}
