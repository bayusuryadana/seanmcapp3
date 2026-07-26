package repository

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigGetValues(t *testing.T) {
	db, mock := newMockDB(t)
	repo := &ConfigRepoImpl{DB: db}

	query := "SELECT key, value FROM app_config WHERE key IN ($1, $2)"
	rows := sqlmock.NewRows([]string{"key", "value"}).
		AddRow(ConfigKeyIGSessionID, "session").
		AddRow(ConfigKeyIGCSRFToken, "csrf")
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(ConfigKeyIGSessionID, ConfigKeyIGCSRFToken).
		WillReturnRows(rows)

	got, err := repo.GetValues(ConfigKeyIGSessionID, ConfigKeyIGCSRFToken)
	require.NoError(t, err)
	assert.Equal(t, "session", got[ConfigKeyIGSessionID])
	assert.Equal(t, "csrf", got[ConfigKeyIGCSRFToken])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConfigGetValuesQueryError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := &ConfigRepoImpl{DB: db}

	mock.ExpectQuery("SELECT key, value FROM app_config").
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetValues("key")
	assert.Error(t, err)
}

func TestConfigSetValues(t *testing.T) {
	db, mock := newMockDB(t)
	repo := &ConfigRepoImpl{DB: db}

	query := "INSERT INTO app_config (key, value) VALUES ($1, $2), ($3, $4) " +
		"ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()"
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(ConfigKeyIGCSRFToken, "csrf", ConfigKeyIGSessionID, "session").
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.SetValues(map[string]string{
		ConfigKeyIGSessionID: "session",
		ConfigKeyIGCSRFToken: "csrf",
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConfigEmptyWritesAreNoOps(t *testing.T) {
	db, mock := newMockDB(t)
	repo := &ConfigRepoImpl{DB: db}

	require.NoError(t, repo.SetValues(nil))
	require.NoError(t, repo.SetValues(map[string]string{"": "value", "key": ""}))
	assert.NoError(t, mock.ExpectationsWereMet())
}
