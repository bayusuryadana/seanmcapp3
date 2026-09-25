package service

import (
	"seanmcapp/external"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPriceCachePrices(t *testing.T) {
	now := time.Now()
	cache := newPriceCache()
	cache.put("BBCA", []external.HistoricalPrice{
		{Date: now.AddDate(0, 0, -6), Close: 100}, {Date: now.AddDate(0, 0, -5), Close: 101},
		{Date: now.AddDate(0, 0, -4), Close: 102}, {Date: now.AddDate(0, 0, -3), Close: 103},
		{Date: now.AddDate(0, 0, -2), Close: 104}, {Date: now.AddDate(0, 0, -1), Close: 105},
		{Date: now, Close: 106},
	})

	latest, reference, err := cache.prices("BBCA", "1d")
	require.NoError(t, err)
	assert.Equal(t, 106.0, latest)
	assert.Equal(t, 105.0, reference)

	latest, reference, err = cache.prices("BBCA", "5d")
	require.NoError(t, err)
	assert.Equal(t, 106.0, latest)
	assert.Equal(t, 101.0, reference)

	_, _, err = cache.prices("BBCA", "invalid")
	assert.Error(t, err)
	_, _, err = cache.prices("MISSING", "1d")
	assert.Error(t, err)
}

func TestProgressionHelpers(t *testing.T) {
	points := []DashboardProgressPoint{
		{Date: "2026-01-30", Index: 0, Portfolio: 0},
		{Date: "2026-02-02", Index: 1, Portfolio: 2},
		{Date: "2026-02-04", Index: 3, Portfolio: 4},
		{Date: "2026-02-09", Index: 5, Portfolio: 6},
	}
	assert.Equal(t, points, aggregateProgression(points, "1mo"))
	weekly := aggregateProgression(points, "3mo")
	require.Len(t, weekly, 3)
	assert.Equal(t, "2026-02-04", weekly[1].Date)
	monthly := aggregateProgression(points, "1y")
	require.Len(t, monthly, 2)
	assert.Equal(t, "2026-02-09", monthly[1].Date)
}

func TestPriceCacheLongPeriods(t *testing.T) {
	now := time.Now()
	history := make([]external.HistoricalPrice, 401)
	for i := range history {
		history[i] = external.HistoricalPrice{Date: now.AddDate(0, 0, i-400), Close: float64(100 + i)}
	}
	cache := newPriceCache()
	cache.put("BBCA", history)

	for _, period := range []string{"1mo", "3mo", "6mo", "1y", "ytd"} {
		latest, reference, err := cache.prices("BBCA", period)
		require.NoError(t, err, period)
		assert.Greater(t, latest, reference, period)
	}
}
