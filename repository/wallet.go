package repository

import (
	"database/sql"
	"errors"
	"log"
)

type Wallet struct {
	ID       *int   `db:"id"`
	Date     int    `db:"date"`
	Name     string `db:"name"`
	Category string `db:"category"`
	Currency string `db:"currency"`
	Amount   int    `db:"amount"`
	Done     bool   `db:"done"`
	Account  string `db:"account"`
}

type WalletRepo interface {
	GetAll() ([]Wallet, error)
	GetAllocations() (map[string]int, error)
	Insert(wallet Wallet) (int, error)
	Update(wallet Wallet) (int, error)
	Delete(id int) (int, error)
}

type WalletRepoImpl struct {
	DB *sql.DB
}

func (r *WalletRepoImpl) GetAll() ([]Wallet, error) {
	rows, err := r.DB.Query("SELECT * FROM wallets")
	if err != nil {
		log.Println("failed to fetch Wallet.GetAll", err)
		return nil, err
	}
	defer rows.Close()

	var wallets []Wallet
	for rows.Next() {
		var w Wallet
		if err := rows.Scan(&w.ID, &w.Date, &w.Name, &w.Category, &w.Currency, &w.Amount, &w.Done, &w.Account); err != nil {
			return nil, err
		}
		wallets = append(wallets, w)
	}
	return wallets, nil
}

func (r *WalletRepoImpl) GetAllocations() (map[string]int, error) {
	rows, err := r.DB.Query("SELECT category, amount FROM allocations")
	if err != nil {
		log.Println("failed to fetch Wallet.GetAllocations", err)
		return nil, err
	}
	defer rows.Close()

	allocations := make(map[string]int)
	for rows.Next() {
		var category string
		var amount int
		if err := rows.Scan(&category, &amount); err != nil {
			return nil, err
		}
		allocations[category] = amount
	}
	return allocations, nil
}

func (r *WalletRepoImpl) Insert(wallet Wallet) (int, error) {
	var id int
	err := r.DB.QueryRow(`
		INSERT INTO wallets (date, name, category, currency, amount, done, account)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		wallet.Date, wallet.Name, wallet.Category, wallet.Currency,
		wallet.Amount, wallet.Done, wallet.Account).Scan(&id)

	if err != nil {
		log.Println("failed to run Wallet.Insert", err)
		return -1, err
	}

	return id, err
}

func (r *WalletRepoImpl) Update(wallet Wallet) (int, error) {
	if wallet.ID == nil {
		return -1, errors.New("wallet ID is required")
	}
	var id int
	err := r.DB.QueryRow(`
		UPDATE wallets SET date=$1, name=$2, category=$3, currency=$4,
		amount=$5, done=$6, account=$7 WHERE id=$8 RETURNING id`,
		wallet.Date, wallet.Name, wallet.Category, wallet.Currency,
		wallet.Amount, wallet.Done, wallet.Account, *wallet.ID).Scan(&id)
	if err == sql.ErrNoRows {
		log.Println("failed to run Wallet.Update", err)
		return -1, nil
	}
	return id, err
}

func (r *WalletRepoImpl) Delete(id int) (int, error) {
	var deletedID int
	err := r.DB.QueryRow("DELETE FROM wallets WHERE id=$1 RETURNING id", id).Scan(&deletedID)
	if err == sql.ErrNoRows {
		log.Println("failed to run Wallet.Delete", err)
		return -1, nil
	}
	return deletedID, err
}
