package repository

import (
	"database/sql"
)

type InstagramAccount struct {
	Username       string `db:"username"`
	LastShortcodes string `db:"last_shortcodes"` // comma-separated, e.g. "ABC123,DEF456,..."
	UserID         string `db:"user_id"`         // numeric instagram user id; empty when not yet resolved
}

type InstagramAccountRepo interface {
	GetAll() ([]InstagramAccount, error)
	UpdateLastShortcodes(username string, shortcodes string) error
	UpdateUserID(username string, userID string) error
}

type InstagramAccountRepoImpl struct {
	DB *sql.DB
}

func (r *InstagramAccountRepoImpl) GetAll() ([]InstagramAccount, error) {
	rows, err := r.DB.Query("SELECT username, COALESCE(last_shortcodes, ''), COALESCE(user_id, '') FROM instagram_accounts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []InstagramAccount
	for rows.Next() {
		var a InstagramAccount
		if err := rows.Scan(&a.Username, &a.LastShortcodes, &a.UserID); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (r *InstagramAccountRepoImpl) UpdateLastShortcodes(username string, shortcodes string) error {
	_, err := r.DB.Exec("UPDATE instagram_accounts SET last_shortcodes = $1 WHERE username = $2", shortcodes, username)
	return err
}

func (r *InstagramAccountRepoImpl) UpdateUserID(username string, userID string) error {
	_, err := r.DB.Exec("UPDATE instagram_accounts SET user_id = $1 WHERE username = $2", userID, username)
	return err
}
