package repository

import (
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
