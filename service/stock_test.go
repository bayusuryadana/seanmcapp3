package service

import (
	"errors"
	"seanmcapp/external"
	"seanmcapp/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func history(closes ...float64) []external.HistoricalPrice {
	points := make([]external.HistoricalPrice, 0, len(closes))
	start := time.Now().AddDate(0, 0, -len(closes)+1)
	for i, close := range closes {
		points = append(points, external.HistoricalPrice{Date: start.AddDate(0, 0, i), Close: close})
	}
	return points
}

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

	repo.getAllFn = func() ([]repository.Stock, error) { return nil, errors.New("db down") }
	_, err = svc.GetAll()
	assert.Error(t, err)
}

func TestStockGetJKSE(t *testing.T) {
	svc := &StockServiceImpl{StockClient: &fakeStockClient{jkse: external.IndexQuote{CurrentPrice: 7123.45, PreviousClose: 7000}}}

	quote, err := svc.GetJKSE()
	require.NoError(t, err)
	assert.Equal(t, 7123.45, quote.CurrentPrice)
	assert.Equal(t, 7000.0, quote.PreviousClose)
}

func TestStockSummaryAndProgression(t *testing.T) {
	stocks := []repository.Stock{{Name: "BBCA", Status: true, Lot: ptr[int64](1)}}
	svc := &StockServiceImpl{
		StockRepo:  &fakeStockRepo{getAllFn: func() ([]repository.Stock, error) { return stocks, nil }},
		priceCache: newPriceCache(),
	}
	indexHistory := history(100, 101, 102, 103, 104, 105, 106)
	stockHistory := make([]external.HistoricalPrice, len(indexHistory))
	for i, point := range indexHistory {
		stockHistory[i] = external.HistoricalPrice{Date: point.Date, Close: point.Close * 10}
	}
	svc.priceCache.put("^JKSE", indexHistory)
	svc.priceCache.put("BBCA", stockHistory)

	summary, err := svc.GetSummary("1d")
	require.NoError(t, err)
	assert.Equal(t, 1.0, summary.JKSE.Delta)
	assert.InDelta(t, 0.952, summary.Portfolio.Percentage, 0.001)
	assert.Equal(t, 1000.0, summary.Positions["BBCA"].Delta)

	progression, err := svc.GetProgression("all")
	require.NoError(t, err)
	require.Len(t, progression, 6)
	assert.Equal(t, 0.0, progression[0].Index)
	assert.InDelta(t, 4.95, progression[len(progression)-1].Portfolio, 0.01)

	_, err = svc.GetSummary("all")
	assert.Error(t, err)
	_, err = svc.GetProgression("bad")
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

	assert.ElementsMatch(t, []string{"BBCA", "TLKM"}, client.calls)
	require.Len(t, repo.updated, 2)
	assert.Equal(t, int64(155), *repo.updated[0].CurrentPrice)
}

func TestStockRunAlerts(t *testing.T) {
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

func TestStockRefreshPricesErrors(t *testing.T) {
	t.Run("GetAll error propagates", func(t *testing.T) {
		repo := &fakeStockRepo{getAllFn: func() ([]repository.Stock, error) { return nil, errors.New("db") }}
		svc := &StockServiceImpl{StockRepo: repo, StockClient: &fakeStockClient{}}
		_, err := svc.RefreshPrices()
		assert.Error(t, err)
	})

	t.Run("client price error is skipped, still returns stocks", func(t *testing.T) {
		stocks := []repository.Stock{{Name: "BBCA", BestPrice: 100, FairPrice: 200}}
		repo := &fakeStockRepo{getAllFn: func() ([]repository.Stock, error) { return stocks, nil }}
		client := &fakeStockClient{err: errors.New("fetch failed")}
		svc := &StockServiceImpl{StockRepo: repo, StockClient: client}

		got, err := svc.RefreshPrices()
		require.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Empty(t, repo.updated)
	})
}

func TestStockRunGetAllError(t *testing.T) {
	repo := &fakeStockRepo{getAllFn: func() ([]repository.Stock, error) { return nil, errors.New("db") }}
	tg := &fakeTelegramClient{}
	svc := &StockServiceImpl{StockRepo: repo, StockClient: &fakeStockClient{}, TelegramClient: tg}

	svc.Run()
	assert.Empty(t, tg.messages)
}
