package repository

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStockGetAll(t *testing.T) {
	db, mock := newMockDB(t)
	repo := &StockRepoImpl{DB: db}

	rows := sqlmock.NewRows([]string{"name", "best_price", "current_price", "fair_price", "status", "buy_price", "lot"}).
		AddRow("BBCA", 100, 150, 200, true, 90, 5)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM stocks")).WillReturnRows(rows)

	got, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "BBCA", got[0].Name)
	assert.Equal(t, int64(150), *got[0].CurrentPrice)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStockCreate(t *testing.T) {
	db, mock := newMockDB(t)
	repo := &StockRepoImpl{DB: db}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO stocks")).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("BBCA"))

	name, err := repo.Create(Stock{Name: "BBCA", BestPrice: 100, FairPrice: 200, Status: true})
	require.NoError(t, err)
	assert.Equal(t, "BBCA", name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStockUpdate(t *testing.T) {
	t.Run("requires name", func(t *testing.T) {
		db, _ := newMockDB(t)
		repo := &StockRepoImpl{DB: db}
		_, err := repo.Update(Stock{})
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		repo := &StockRepoImpl{DB: db}
		mock.ExpectQuery(regexp.QuoteMeta("UPDATE stocks")).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("BBCA"))
		name, err := repo.Update(Stock{Name: "BBCA"})
		require.NoError(t, err)
		assert.Equal(t, "BBCA", name)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock := newMockDB(t)
		repo := &StockRepoImpl{DB: db}
		mock.ExpectQuery(regexp.QuoteMeta("UPDATE stocks")).WillReturnError(sql.ErrNoRows)
		_, err := repo.Update(Stock{Name: "BBCA"})
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestStockDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		repo := &StockRepoImpl{DB: db}
		mock.ExpectQuery(regexp.QuoteMeta("DELETE FROM stocks")).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("BBCA"))
		name, err := repo.Delete("BBCA")
		require.NoError(t, err)
		assert.Equal(t, "BBCA", name)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock := newMockDB(t)
		repo := &StockRepoImpl{DB: db}
		mock.ExpectQuery(regexp.QuoteMeta("DELETE FROM stocks")).WillReturnError(sql.ErrNoRows)
		_, err := repo.Delete("BBCA")
		assert.ErrorIs(t, err, ErrNotFound)
	})
}
