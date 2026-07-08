package repository

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func TestWalletGetAll(t *testing.T) {
	db, mock := newMockDB(t)
	repo := &WalletRepoImpl{DB: db}

	rows := sqlmock.NewRows([]string{"id", "date", "name", "category", "currency", "amount", "done", "account"}).
		AddRow(1, 202406, "a", "Daily", "SGD", -100, true, "DBS")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM wallets")).WillReturnRows(rows)

	got, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "a", got[0].Name)
	assert.Equal(t, -100, got[0].Amount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWalletGetAllocations(t *testing.T) {
	db, mock := newMockDB(t)
	repo := &WalletRepoImpl{DB: db}

	rows := sqlmock.NewRows([]string{"category", "amount"}).AddRow("Daily", 1000).AddRow("Rent", 500)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT category, amount FROM allocations")).WillReturnRows(rows)

	got, err := repo.GetAllocations()
	require.NoError(t, err)
	assert.Equal(t, map[string]int{"Daily": 1000, "Rent": 500}, got)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWalletInsert(t *testing.T) {
	db, mock := newMockDB(t)
	repo := &WalletRepoImpl{DB: db}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO wallets")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(99))

	id, err := repo.Insert(Wallet{Date: 202406, Name: "a", Account: "DBS"})
	require.NoError(t, err)
	assert.Equal(t, 99, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWalletUpdate(t *testing.T) {
	t.Run("requires id", func(t *testing.T) {
		db, _ := newMockDB(t)
		repo := &WalletRepoImpl{DB: db}
		_, err := repo.Update(Wallet{})
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		repo := &WalletRepoImpl{DB: db}
		id := 5
		mock.ExpectQuery(regexp.QuoteMeta("UPDATE wallets")).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
		got, err := repo.Update(Wallet{ID: &id})
		require.NoError(t, err)
		assert.Equal(t, 5, got)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock := newMockDB(t)
		repo := &WalletRepoImpl{DB: db}
		id := 5
		mock.ExpectQuery(regexp.QuoteMeta("UPDATE wallets")).WillReturnError(sql.ErrNoRows)
		_, err := repo.Update(Wallet{ID: &id})
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestWalletDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		repo := &WalletRepoImpl{DB: db}
		mock.ExpectQuery(regexp.QuoteMeta("DELETE FROM wallets")).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
		got, err := repo.Delete(7)
		require.NoError(t, err)
		assert.Equal(t, 7, got)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock := newMockDB(t)
		repo := &WalletRepoImpl{DB: db}
		mock.ExpectQuery(regexp.QuoteMeta("DELETE FROM wallets")).WillReturnError(sql.ErrNoRows)
		_, err := repo.Delete(7)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}
