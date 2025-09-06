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
