package service

import (
	"errors"
	"seanmcapp/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStockGetAll(t *testing.T) {
	repo := &fakeStockRepo{getAllFn: func() ([]repository.Stock, error) {
		return []repository.Stock{
			{Name: "BBCA", BestPrice: 100, FairPrice: 200, Status: true, CurrentPrice: ptr[int64](150)},
		}, nil
	}}
	svc := &StockServiceImpl{StockRepo: repo}

	got, err := svc.GetAll()
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "BBCA", got[0].Name)
	assert.Equal(t, int64(150), *got[0].CurrentPrice)

	// error passthrough
	repo.getAllFn = func() ([]repository.Stock, error) { return nil, errors.New("db down") }
	_, err = svc.GetAll()
	assert.Error(t, err)
}

func TestStockCreateValidation(t *testing.T) {
	svc := &StockServiceImpl{StockRepo: &fakeStockRepo{
		createFn: func(s repository.Stock) (string, error) { return s.Name, nil },
	}}

	_, err := svc.Create(DashboardStock{Name: "X", BestPrice: 0, FairPrice: 10})
	assert.ErrorAs(t, err, &ValidationError{})

	_, err = svc.Create(DashboardStock{Name: "X", BestPrice: 10, FairPrice: 0})
	assert.ErrorAs(t, err, &ValidationError{})

	name, err := svc.Create(DashboardStock{Name: "BBCA", BestPrice: 100, FairPrice: 200})
	require.NoError(t, err)
	assert.Equal(t, "BBCA", name)
}

func TestStockUpdateAndDelete(t *testing.T) {
	t.Run("update validation", func(t *testing.T) {
		svc := &StockServiceImpl{StockRepo: &fakeStockRepo{}}
		_, err := svc.Update(DashboardStock{Name: "X", BestPrice: -1, FairPrice: 10})
		assert.ErrorAs(t, err, &ValidationError{})
	})

	t.Run("update passes through ErrNotFound", func(t *testing.T) {
		svc := &StockServiceImpl{StockRepo: &fakeStockRepo{
			updateFn: func(repository.Stock) (string, error) { return "", repository.ErrNotFound },
		}}
		_, err := svc.Update(DashboardStock{Name: "X", BestPrice: 1, FairPrice: 1})
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("delete success", func(t *testing.T) {
		svc := &StockServiceImpl{StockRepo: &fakeStockRepo{
			deleteFn: func(name string) (string, error) { return name, nil },
		}}
		name, err := svc.Delete("BBCA")
		require.NoError(t, err)
		assert.Equal(t, "BBCA", name)
	})
}

func TestStockRefreshPrices(t *testing.T) {
	stocks := []repository.Stock{
		{Name: "BBCA", BestPrice: 100, FairPrice: 200, Status: true, CurrentPrice: ptr[int64](150)},
		{Name: "TLKM", BestPrice: 300, FairPrice: 400, Status: false, CurrentPrice: ptr[int64](310)},
	}
	repo := &fakeStockRepo{getAllFn: func() ([]repository.Stock, error) { return stocks, nil }}
	client := &fakeStockClient{prices: map[string]int64{"BBCA": 155, "TLKM": 320}}
	svc := &StockServiceImpl{StockRepo: repo, StockClient: client}

	got, err := svc.RefreshPrices()
	require.NoError(t, err)
	assert.Len(t, got, 2)

	// Every stock's price was fetched and persisted.
	assert.ElementsMatch(t, []string{"BBCA", "TLKM"}, client.calls)
	require.Len(t, repo.updated, 2)
	assert.Equal(t, int64(155), *repo.updated[0].CurrentPrice)
}

func TestStockRunAlerts(t *testing.T) {
	// BBCA: wishlist (status=false) and current <= best  -> "hitting best price"
	// TLKM: owned   (status=true)  and current >= fair  -> "reaching fair price"
	// GOTO: current price nil -> skipped entirely
	stocks := []repository.Stock{
		{Name: "BBCA", BestPrice: 100, FairPrice: 200, Status: false, CurrentPrice: ptr[int64](90)},
		{Name: "TLKM", BestPrice: 300, FairPrice: 400, Status: true, CurrentPrice: ptr[int64](410)},
		{Name: "GOTO", BestPrice: 50, FairPrice: 80, Status: false, CurrentPrice: nil},
	}
	repo := &fakeStockRepo{getAllFn: func() ([]repository.Stock, error) { return stocks, nil }}
	client := &fakeStockClient{prices: map[string]int64{"BBCA": 90, "TLKM": 410, "GOTO": 60}}
	tg := &fakeTelegramClient{}
	svc := &StockServiceImpl{StockRepo: repo, StockClient: client, TelegramClient: tg, PersonalChatID: 99}

	svc.Run()

	require.Len(t, tg.messages, 1)
	assert.Equal(t, int64(99), tg.messages[0].chatID)
	assert.Equal(t, "BBCA hitting best price\nTLKM reaching fair price", tg.messages[0].text)
}

func TestStockRunNoAlerts(t *testing.T) {
	stocks := []repository.Stock{
		{Name: "BBCA", BestPrice: 100, FairPrice: 200, Status: false, CurrentPrice: ptr[int64](150)},
	}
	repo := &fakeStockRepo{getAllFn: func() ([]repository.Stock, error) { return stocks, nil }}
	client := &fakeStockClient{prices: map[string]int64{"BBCA": 150}}
	tg := &fakeTelegramClient{}
	svc := &StockServiceImpl{StockRepo: repo, StockClient: client, TelegramClient: tg}

	svc.Run()
	assert.Empty(t, tg.messages)
}
