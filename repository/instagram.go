package repository

import (
	"database/sql"
	"log"
)

type InstagramAccount struct {
	Username       string `db:"username"`
	LastShortcodes string `db:"last_shortcodes"` // comma-separated, e.g. "ABC123,DEF456,..."
}

type InstagramAccountRepo interface {
	GetAll() ([]InstagramAccount, error)
	UpdateLastShortcodes(username string, shortcodes string) error
}

type InstagramAccountRepoImpl struct {
	DB *sql.DB
}

func (r *InstagramAccountRepoImpl) GetAll() ([]InstagramAccount, error) {
	rows, err := r.DB.Query("SELECT username, COALESCE(last_shortcodes, '') FROM instagram_accounts")
	if err != nil {
		log.Println("failed to fetch instagram_accounts", err)
		return nil, err
	}
	defer rows.Close()

	var accounts []InstagramAccount
	for rows.Next() {
		var a InstagramAccount
		if err := rows.Scan(&a.Username, &a.LastShortcodes); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (r *InstagramAccountRepoImpl) UpdateLastShortcodes(username string, shortcodes string) error {
	_, err := r.DB.Exec("UPDATE instagram_accounts SET last_shortcodes = $1 WHERE username = $2", shortcodes, username)
	if err != nil {
		log.Println("failed to update last_shortcodes", err)
	}
	return err
}
