package repository

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstagramGetAllReadsPostTrackingFields(t *testing.T) {
	db, mock := newMockDB(t)
	repo := &InstagramAccountRepoImpl{DB: db}
	query := "SELECT id, username, COALESCE(last_shortcodes, ''), COALESCE(user_id, '') FROM instagram_accounts"
	rows := sqlmock.NewRows([]string{"id", "username", "last_shortcodes", "user_id"}).AddRow(1, "foo", "AAA", "123")
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(rows)

	accounts, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, accounts, 1)
	assert.Equal(t, "AAA", accounts[0].LastShortcodes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstagramAccountUpdates(t *testing.T) {
	db, mock := newMockDB(t)
	repo := &InstagramAccountRepoImpl{DB: db}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE instagram_accounts SET last_shortcodes = $1 WHERE username = $2")).WithArgs("AAA,BBB", "foo").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE instagram_accounts SET user_id = $1 WHERE username = $2")).WithArgs("123", "foo").WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpdateLastShortcodes("foo", "AAA,BBB"))
	require.NoError(t, repo.UpdateUserID("foo", "123"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstagramGetAllQueryError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := &InstagramAccountRepoImpl{DB: db}
	mock.ExpectQuery("SELECT id, username").WillReturnError(errors.New("query failed"))
	_, err := repo.GetAll()
	assert.Error(t, err)
}
