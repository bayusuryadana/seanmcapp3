package service

import (
	"errors"
	"seanmcapp/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T { return &v }

func TestWalletDashboard(t *testing.T) {
	wallets := []repository.Wallet{
		{ID: ptr(1), Date: 202405, Name: "a", Category: "Daily", Currency: "SGD", Amount: -100, Done: true, Account: "DBS"},
		{ID: ptr(2), Date: 202406, Name: "b", Category: "Rent", Currency: "SGD", Amount: -50, Done: true, Account: "DBS"},
		{ID: ptr(3), Date: 202406, Name: "c", Category: "Travel", Currency: "IDR", Amount: -25400, Done: true, Account: "BCA"},
		{ID: ptr(4), Date: 202404, Name: "d", Category: "Salary", Currency: "SGD", Amount: 5000, Done: true, Account: "DBS"},
	}
	repo := &fakeWalletRepo{
		getAllFn: func() ([]repository.Wallet, error) { return wallets, nil },
		getAllocationsFn: func() (map[string]int, error) {
			return map[string]int{"Daily": 1000, "Rent": 500}, nil
		},
	}
	svc := &WalletServiceImpl{WalletRepo: repo}

	view, err := svc.Dashboard(202406)
	require.NoError(t, err)

	// Savings = sum of Done entries per account.
	assert.Equal(t, 4850, view.Savings.DBS)
	assert.Equal(t, -25400, view.Savings.BCA)

	// Planned = entries whose date <= requested date.
	assert.Equal(t, 4850, view.Planned.SGD)
	assert.Equal(t, -25400, view.Planned.IDR)

	// Allocations follow the fixed category order with expense (sign-flipped) and alloc.
	expectedAlloc := []DashboardAllocations{
		{Name: "Daily", Expense: 100, Alloc: 1000},
		{Name: "Rent", Expense: 50, Alloc: 500},
		{Name: "Travel", Expense: 2, Alloc: 0},
		{Name: "Fashion", Expense: 0, Alloc: 0},
		{Name: "IT Stuff", Expense: 0, Alloc: 0},
		{Name: "Misc", Expense: 0, Alloc: 0},
		{Name: "Wellness", Expense: 0, Alloc: 0},
		{Name: "Funding", Expense: 0, Alloc: 0},
	}
	assert.Equal(t, expectedAlloc, view.Allocations)

	// Balance history: cumulative DBS totals up to the date, newest first.
	expectedBalance := []DashboardBalance{
		{Date: 202406, Sum: 4850},
		{Date: 202405, Sum: 4900},
		{Date: 202404, Sum: 5000},
	}
	assert.Equal(t, expectedBalance, view.Chart.BalanceHistory)

	// Detail wallets = entries matching exactly the requested date.
	require.Len(t, view.Wallets, 2)
	assert.Equal(t, "b", view.Wallets[0].Name)
	assert.Equal(t, "c", view.Wallets[1].Name)
}

func TestWalletDashboardErrors(t *testing.T) {
	boom := errors.New("boom")

	t.Run("GetAll fails", func(t *testing.T) {
		svc := &WalletServiceImpl{WalletRepo: &fakeWalletRepo{
			getAllFn: func() ([]repository.Wallet, error) { return nil, boom },
		}}
		_, err := svc.Dashboard(202406)
		assert.ErrorIs(t, err, boom)
	})

	t.Run("GetAllocations fails", func(t *testing.T) {
		svc := &WalletServiceImpl{WalletRepo: &fakeWalletRepo{
			getAllFn:         func() ([]repository.Wallet, error) { return nil, nil },
			getAllocationsFn: func() (map[string]int, error) { return nil, boom },
		}}
		_, err := svc.Dashboard(202406)
		assert.ErrorIs(t, err, boom)
	})
}

func TestWalletCreateUpdateDelete(t *testing.T) {
	t.Run("create success", func(t *testing.T) {
		repo := &fakeWalletRepo{insertFn: func(w repository.Wallet) (int, error) {
			assert.Equal(t, "x", w.Name)
			return 42, nil
		}}
		svc := &WalletServiceImpl{WalletRepo: repo}
		id, err := svc.Create(DashboardWallet{Name: "x"})
		require.NoError(t, err)
		assert.Equal(t, 42, id)
	})

	t.Run("update passes through ErrNotFound", func(t *testing.T) {
		repo := &fakeWalletRepo{updateFn: func(repository.Wallet) (int, error) {
			return -1, repository.ErrNotFound
		}}
		svc := &WalletServiceImpl{WalletRepo: repo}
		_, err := svc.Update(DashboardWallet{ID: ptr(1)})
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("delete success", func(t *testing.T) {
		repo := &fakeWalletRepo{deleteFn: func(id int) (int, error) { return id, nil }}
		svc := &WalletServiceImpl{WalletRepo: repo}
		id, err := svc.Delete(7)
		require.NoError(t, err)
		assert.Equal(t, 7, id)
	})
}
