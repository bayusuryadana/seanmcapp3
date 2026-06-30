package repository

import (
	"database/sql"
	"errors"
)

type Stock struct {
	Name         string `db:"name"`
	BestPrice    int64  `db:"best_price"`
	CurrentPrice *int64 `db:"current_price"`
	FairPrice    int64  `db:"fair_price"`
	Status       bool   `db:"status"` // 0 -> wishlist, 1 -> bought
	BuyPrice     *int64 `db:"buy_price"`
	Lot          *int64 `db:"lot"`
}

type StockRepo interface {
	GetAll() ([]Stock, error)
	Create(stock Stock) (string, error)
	Update(stock Stock) (string, error)
	Delete(name string) (string, error)
}

type StockRepoImpl struct {
	DB *sql.DB
}

func (r *StockRepoImpl) GetAll() ([]Stock, error) {
	rows, err := r.DB.Query("SELECT * FROM stocks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []Stock
	for rows.Next() {
		var s Stock
		if err := rows.Scan(&s.Name, &s.BestPrice, &s.CurrentPrice, &s.FairPrice, &s.Status, &s.BuyPrice, &s.Lot); err != nil {
			return nil, err
		}
		stocks = append(stocks, s)
	}
	return stocks, nil
}

func boolToBit(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func (r *StockRepoImpl) Create(stock Stock) (string, error) {
	var name string
	err := r.DB.QueryRow(`
		INSERT INTO stocks (name, best_price, current_price, fair_price, status, buy_price, lot)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING name`,
		stock.Name, stock.BestPrice, stock.CurrentPrice, stock.FairPrice, boolToBit(stock.Status), stock.BuyPrice, stock.Lot).Scan(&name)
	return name, err
}

func (r *StockRepoImpl) Update(stock Stock) (string, error) {
	if stock.Name == "" {
		return "", errors.New("stock name is required")
	}
	var name string
	err := r.DB.QueryRow(`
		UPDATE stocks SET best_price=$1, current_price=$2, fair_price=$3, status=$4, buy_price=$5, lot=$6
		WHERE name=$7 RETURNING name`,
		stock.BestPrice, stock.CurrentPrice, stock.FairPrice, boolToBit(stock.Status), stock.BuyPrice, stock.Lot, stock.Name).Scan(&name)
	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	return name, err
}

func (r *StockRepoImpl) Delete(name string) (string, error) {
	var deletedName string
	err := r.DB.QueryRow("DELETE FROM stocks WHERE name=$1 RETURNING name", name).Scan(&deletedName)
	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	return deletedName, err
}
