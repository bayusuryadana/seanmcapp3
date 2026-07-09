package repository

import (
	"database/sql"
)

type InstagramAccount struct {
	Username       string `db:"username"`
	LastShortcodes string `db:"last_shortcodes"` // comma-separated post shortcodes, e.g. "ABC123,DEF456,..."
	UserID         string `db:"user_id"`         // numeric instagram user id; empty when not yet resolved
	LastStoryIDs   string `db:"last_story_ids"`  // comma-separated story pks currently seen
}

type InstagramAccountRepo interface {
	GetAll() ([]InstagramAccount, error)
	UpdateLastShortcodes(username string, shortcodes string) error
	UpdateUserID(username string, userID string) error
	UpdateLastStoryIDs(username string, storyIDs string) error
}

type InstagramAccountRepoImpl struct {
	DB *sql.DB
}

func (r *InstagramAccountRepoImpl) GetAll() ([]InstagramAccount, error) {
	rows, err := r.DB.Query("SELECT username, COALESCE(last_shortcodes, ''), COALESCE(user_id, ''), COALESCE(last_story_ids, '') FROM instagram_accounts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []InstagramAccount
	for rows.Next() {
		var a InstagramAccount
		if err := rows.Scan(&a.Username, &a.LastShortcodes, &a.UserID, &a.LastStoryIDs); err != nil {
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

func (r *InstagramAccountRepoImpl) UpdateLastStoryIDs(username string, storyIDs string) error {
	_, err := r.DB.Exec("UPDATE instagram_accounts SET last_story_ids = $1 WHERE username = $2", storyIDs, username)
	return err
}
