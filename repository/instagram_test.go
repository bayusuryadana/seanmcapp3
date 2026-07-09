package repository

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstagramGetAll(t *testing.T) {
	db, mock := newMockDB(t)
	repo := &InstagramAccountRepoImpl{DB: db}

	rows := sqlmock.NewRows([]string{"username", "last_shortcodes", "user_id"}).
		AddRow("foo", "AAA,BBB", "123").
		AddRow("bar", "", "")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT username, COALESCE(last_shortcodes, ''), COALESCE(user_id, '') FROM instagram_accounts")).
		WillReturnRows(rows)

	got, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "foo", got[0].Username)
	assert.Equal(t, "AAA,BBB", got[0].LastShortcodes)
	assert.Equal(t, "123", got[0].UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstagramUpdateLastShortcodes(t *testing.T) {
	db, mock := newMockDB(t)
	repo := &InstagramAccountRepoImpl{DB: db}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE instagram_accounts SET last_shortcodes")).
		WithArgs("AAA,BBB", "foo").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateLastShortcodes("foo", "AAA,BBB")
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstagramUpdateUserID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := &InstagramAccountRepoImpl{DB: db}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE instagram_accounts SET user_id")).
		WithArgs("123", "foo").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateUserID("foo", "123")
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
