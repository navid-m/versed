package database

import (
	"database/sql"

	"github.com/navid-m/versed/models"
)

// CreateBannedIPTable creates the banned_ips table if it doesn't exist
func CreateBannedIPTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS banned_ips (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ip_address TEXT NOT NULL UNIQUE,
			banned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			banned_by INTEGER NOT NULL,
			reason TEXT,
			is_active BOOLEAN DEFAULT 1,
			unbanned_at DATETIME,
			unbanned_by INTEGER,
			FOREIGN KEY (banned_by) REFERENCES users(id),
			FOREIGN KEY (unbanned_by) REFERENCES users(id)
		)
	`
	_, err := db.Exec(query)
	return err
}

// BanIP bans an IP address
func BanIP(ipAddress, reason string, bannedBy int) error {
	query := `
		INSERT INTO banned_ips (ip_address, banned_by, reason, is_active)
		VALUES (?, ?, ?, 1)
		ON CONFLICT(ip_address) DO UPDATE SET
			banned_at = CURRENT_TIMESTAMP,
			banned_by = ?,
			reason = ?,
			is_active = 1,
			unbanned_at = NULL,
			unbanned_by = NULL
	`
	_, err := GetDB().Exec(query, ipAddress, bannedBy, reason, bannedBy, reason)
	return err
}

// UnbanIP unbans an IP address
func UnbanIP(ipAddress string, unbannedBy int) error {
	query := `
		UPDATE banned_ips
		SET is_active = 0,
			unbanned_at = CURRENT_TIMESTAMP,
			unbanned_by = ?
		WHERE ip_address = ? AND is_active = 1
	`
	_, err := GetDB().Exec(query, unbannedBy, ipAddress)
	return err
}

// IsIPBanned checks if an IP address is currently banned
func IsIPBanned(ipAddress string) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM banned_ips
		WHERE ip_address = ? AND is_active = 1
	`
	var isBanned bool
	err := GetDB().QueryRow(query, ipAddress).Scan(&isBanned)
	return isBanned, err
}

// GetAllBannedIPs retrieves all banned IP addresses
func GetAllBannedIPs() ([]models.BannedIP, error) {
	query := `
		SELECT id, ip_address, banned_at, banned_by, reason, is_active, unbanned_at, unbanned_by
		FROM banned_ips
		ORDER BY banned_at DESC
	`
	rows, err := GetDB().Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bannedIPs []models.BannedIP
	for rows.Next() {
		var bannedIP models.BannedIP
		var unbannedAt sql.NullTime
		var unbannedBy sql.NullInt64

		err := rows.Scan(
			&bannedIP.ID,
			&bannedIP.IPAddress,
			&bannedIP.BannedAt,
			&bannedIP.BannedBy,
			&bannedIP.Reason,
			&bannedIP.IsActive,
			&unbannedAt,
			&unbannedBy,
		)
		if err != nil {
			return nil, err
		}

		if unbannedAt.Valid {
			bannedIP.UnbannedAt = &unbannedAt.Time
		}
		if unbannedBy.Valid {
			unbannedByInt := int(unbannedBy.Int64)
			bannedIP.UnbannedBy = &unbannedByInt
		}

		bannedIPs = append(bannedIPs, bannedIP)
	}

	return bannedIPs, nil
}

// GetBannedIPByID retrieves a banned IP by its ID
func GetBannedIPByID(id int) (*models.BannedIP, error) {
	query := `
		SELECT id, ip_address, banned_at, banned_by, reason, is_active, unbanned_at, unbanned_by
		FROM banned_ips
		WHERE id = ?
	`
	var bannedIP models.BannedIP
	var unbannedAt sql.NullTime
	var unbannedBy sql.NullInt64

	err := GetDB().QueryRow(query, id).Scan(
		&bannedIP.ID,
		&bannedIP.IPAddress,
		&bannedIP.BannedAt,
		&bannedIP.BannedBy,
		&bannedIP.Reason,
		&bannedIP.IsActive,
		&unbannedAt,
		&unbannedBy,
	)
	if err != nil {
		return nil, err
	}

	if unbannedAt.Valid {
		bannedIP.UnbannedAt = &unbannedAt.Time
	}
	if unbannedBy.Valid {
		unbannedByInt := int(unbannedBy.Int64)
		bannedIP.UnbannedBy = &unbannedByInt
	}

	return &bannedIP, nil
}

// UpdateUserAdminStatus updates a user's admin status
func UpdateUserAdminStatus(userID int, isAdmin bool) error {
	query := `UPDATE users SET is_admin = ? WHERE id = ?`
	_, err := GetDB().Exec(query, isAdmin, userID)
	return err
}

// IsUserAdmin checks if a user is an admin
func IsUserAdmin(userID int) (bool, error) {
	query := `SELECT is_admin FROM users WHERE id = ?`
	var isAdmin bool
	err := GetDB().QueryRow(query, userID).Scan(&isAdmin)
	return isAdmin, err
}
